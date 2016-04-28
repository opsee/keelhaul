package launcher

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/basic/com"
	"github.com/opsee/basic/schema"
	"github.com/opsee/basic/service"
	"github.com/opsee/keelhaul/autocheck"
	"github.com/opsee/keelhaul/bus"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
	log "github.com/sirupsen/logrus"
)

const (
	commandLaunchBastion  = "launch-bastion"
	commandConnectBastion = "connect-bastion"
	commandDiscovery      = "discovery"

	stateInProgress = "in-progress"
	stateComplete   = "complete"
	stateFailed     = "failed"

	userdata = `#cloud-config
write_files:
  - path: "/etc/opsee/bastion-env.sh"
    permissions: "0644"
    owner: "root"
    content: |
      CUSTOMER_ID={{.User.CustomerId}}
      CUSTOMER_EMAIL={{.User.Email}}
      BASTION_VERSION={{.Config.Tag}}
      BASTION_ID={{.Bastion.ID}}
      VPN_PASSWORD={{.Bastion.Password}}
      VPN_REMOTE={{.Config.VPNRemote}}
      DNS_SERVER={{.Config.DNSServer}}
      NSQD_HOST={{.Config.NSQDHost}}
      BARTNET_HOST={{.Config.BartnetHost}}
      BASTION_AUTH_TYPE={{.Config.AuthType}}
      GODEBUG=netdns=cgo
{{ with .BastionUsers }}users:{{ range . }}
  - name: "{{ .Username }}"
    groups:
      - "sudo"
    ssh-authorized-keys:
      - "{{ .Key }}{{ end }}{{ end }}"
coreos:
  units:
    - name: "docker.service"
      drop-ins:
        - name: "10-cgroupfs.conf"
          content: |
            [Service]
            Environment="DOCKER_OPTS=--exec-opt=native.cgroupdriver=cgroupfs"
        - name: "50-reboot.conf"
          content: |
            [Service]
            FailureAction=reboot-force
  update:
    reboot-strategy: "off"
    group: "beta"
`
)

var userdataTmpl = template.Must(template.New("userdata").Parse(userdata))

type BastionConfig struct {
	OwnerID       string `json:"owner_id"`
	Tag           string `json:"tag"`
	KeyPair       string `json:"keypair"`
	VPNRemote     string `json:"vpn_remote"`
	DNSServer     string `json:"dns_server"`
	NSQDHost      string `json:"nsqd_host"`
	BartnetHost   string `json:"bartnet_host"`
	AuthType      string `json:"auth_type"`
	ModifiedIndex uint64 `json:"modified_index"`
}

type Stage interface {
	Execute(*Launch)
}

type Launch struct {
	Bastion                   *com.Bastion
	User                      *schema.User
	Autochecks                *autocheck.Pool
	EventChan                 chan *Event
	Err                       error
	VPCEnvironment            *VPCEnvironment
	BastionIngressTemplateURL string
	ImageID                   string
	ImageTag                  string
	InstanceType              string
	state                     int
	stateMut                  *sync.RWMutex
	session                   *session.Session
	logger                    *log.Entry
	db                        store.Store
	router                    router.Router
	etcd                      etcd.KeysAPI
	spanx                     service.SpanxClient
	config                    *config.Config
	bastionConfig             *BastionConfig
	sqsClient                 sqsiface.SQSAPI
	snsClient                 snsiface.SNSAPI
	cloudformationClient      cloudformationiface.CloudFormationAPI
	subscribeOutput           *sns.SubscribeOutput
	createTopicOutput         *sns.CreateTopicOutput
	createQueueOutput         *sqs.CreateQueueOutput
	getQueueAttributesOutput  *sqs.GetQueueAttributesOutput
	setQueueAttributesOutput  *sqs.SetQueueAttributesOutput
	createStackOutput         *cloudformation.CreateStackOutput
	connectAttempts           float64
}

type Event struct {
	Err     error
	Message *bus.Message
}

type VPCEnvironment struct {
	SecurityGroupCount    int   `json:"security_group_count"`
	DBSecurityGroupCount  int   `json:"db_security_group_count"`
	LoadBalancerCount     int   `json:"load_balancer_count"`
	AutoscalingGroupCount int   `json:"autoscaling_group_count"`
	InstanceCount         int   `json:"instance_count"`
	DBInstanceCount       int   `json:"db_instance_count"`
	GroupErrorCount       int   `json:"group_error_count"`
	InstanceErrorCount    int   `json:"instance_error_count"`
	LastError             error `json:"last_error"`
}

func NewLaunch(db store.Store, router router.Router, etcdKAPI etcd.KeysAPI, spanx service.SpanxClient, cfg *config.Config, sess *session.Session, user *schema.User) *Launch {
	logger := log.WithFields(log.Fields{
		"customer_id": user.CustomerId,
		"user_id":     user.Id,
	})

	return &Launch{
		User:            user,
		EventChan:       make(chan *Event),
		VPCEnvironment:  &VPCEnvironment{},
		Autochecks:      autocheck.NewPool(autocheck.NewBartnetSink(cfg.BartnetEndpoint, cfg.HugsEndpoint, user), logger),
		state:           3,
		stateMut:        &sync.RWMutex{},
		db:              db,
		router:          router,
		etcd:            etcdKAPI,
		spanx:           spanx,
		config:          cfg,
		session:         sess,
		logger:          logger,
		connectAttempts: float64(1),
	}
}

func (launch *Launch) NotifyVars() interface{} {
	vars := struct {
		*VPCEnvironment
		Error        string `json:"error"`
		UserID       int    `json:"user_id"`
		UserEmail    string `json:"user_email"`
		CustomerID   string `json:"customer_id"`
		Region       string `json:"region"`
		ImageID      string `json:"image_id"`
		VPCID        string `json:"vpc_id"`
		SubnetID     string `json:"subnet_id"`
		InstanceID   string `json:"instance_id"`
		GroupID      string `json:"group_id"`
		InstanceName string `json:"instance_name"`
		GroupName    string `json:"group_name"`
		CheckCount   int    `json:"check_count"`
	}{
		VPCEnvironment: launch.VPCEnvironment,
		UserID:         int(launch.User.Id),
		UserEmail:      launch.User.Email,
		CustomerID:     launch.User.CustomerId,
		Region:         *launch.session.Config.Region,
		VPCID:          launch.Bastion.VPCID,
		SubnetID:       launch.Bastion.SubnetID,
		ImageID:        launch.Bastion.ImageID.String,
		InstanceID:     launch.Bastion.InstanceID.String,
		GroupID:        launch.Bastion.GroupID.String,
		InstanceName:   "Opsee Instance",
		GroupName:      "Opsee Instance Security Group",
		CheckCount:     launch.Autochecks.SuccessCount(),
	}

	if launch.Err != nil {
		vars.Error = launch.Err.Error()
	}

	return vars
}

// these events happen synchronously in the request cycle, so they are not part of launch stages
func (launch *Launch) CreateBastion(region, vpcID, subnetID, subnetRouting, instanceType string) error {
	bastion, err := com.NewBastion(int(launch.User.Id), launch.User.CustomerId, region, vpcID, subnetID, subnetRouting, instanceType)
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed creating bastion credentials",
		})
		return err
	}

	err = launch.db.PutBastion(bastion)
	if err != nil {
		launch.error(err, &bus.Message{
			Command: commandLaunchBastion,
			Message: "failed saving bastion in database",
		})
		return err
	}

	// fold the bastion id into the logger k/v
	launch.logger = launch.logger.WithField("bastion-id", bastion.ID)
	launch.Bastion = bastion
	launch.event(&bus.Message{
		State:   stateInProgress,
		Command: commandLaunchBastion,
		Message: "created bastion object",
	})

	return nil
}

func (launch *Launch) Launch(imageTag string, createRole bool) {
	if createRole {
		launch.stage(putRole{})
	}

	launch.ImageTag = "beta"
	launch.stage(getBastionConfig{})
	launch.stage(getLatestImageID{})
	launch.stage(createTopic{})
	launch.stage(createQueue{})
	launch.stage(getQueueAttributes{})
	launch.stage(setQueueAttributes{})
	launch.stage(subscribe{})
	launch.stage(createStack{})
	launch.stage(bastionLaunchingState{})
	launch.stage(consumeSQS{})
	launch.stage(bastionActiveState{})
	go launch.stage(vpcDiscovery{})
	go launch.stage(waitConnect{})
	//TODO add stage post launch
}

func (launch *Launch) State() string {
	launch.stateMut.RLock()
	defer launch.stateMut.RUnlock()

	if launch.Err != nil {
		return stateFailed
	}

	if launch.state > 0 {
		return stateInProgress
	}

	return stateComplete
}

func (launch *Launch) GenerateUserData() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	var ud = struct {
		User         *schema.User
		BastionUsers []*com.BastionUser
		Config       *BastionConfig
		Bastion      *com.Bastion
	}{
		launch.User,
		[]*com.BastionUser{
			{
				"opsee",
				"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDP+VmyztGmJJTe6YtMtrKazGy3tQC/Pku156Ae10TMzCjvtiol+eL11FKyvNvlENM5EWwIQEng5w3J616kRa92mWr9OWALBn4HJZcztS2YLAXyiC+GLauil6W6xnGzS0DmU5RiYSSPSrmQEwHvmO2umbG190srdaDn/ZvAwptC1br/zc/7ya3XqxHugw1V9kw+KXzTWSC95nPkhOFoaA3nLcMvYWfoTbsU/G08qQy8medqyK80LJJntedpFAYPUrVdGY2J7F2y994YLfapPGzDjM7nR0sRWAZbFgm/BSD0YM8KA0mfGZuKPwKSLMtTUlsmv3l6GJl5a7TkyOlK3zzYtVGO6dnHdZ3X19nldreE3DywpjDrKIfYF2L42FKnpTGFgvunsg9vPdYOiJyIfk6lYsGE6h451OAmV0dxeXhtbqpw4/DsSHtLm5kKjhjRwunuQXEg8SfR3kesJjq6rmhCjLc7bIKm3rSU07zbXSR40JHO1Mc9rqzg2bCk3inJmCKWbMnDvWU1RD475eATEKoG/hv0/7EOywDnFe1m4yi6yZh7XlvakYsxDBPO9/FMlZm2T+cn+TyTmDiw9tEAIEAEiiu18CUNIii1em7XtFDmXjGFWfvteQG/2A98/uDGbmlXd64F2OtU/ulDRJXFGaji8tqxQ/To+2zIeIptLjtqBw==",
			},
		},
		launch.bastionConfig,
		launch.Bastion,
	}

	err := userdataTmpl.Execute(buf, ud)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (launch *Launch) stage(st Stage) {
	if launch.State() == stateFailed {
		return
	}

	st.Execute(launch)
}

func (launch *Launch) handleEvent(event *Event) {
	launch.stateMut.Lock()
	defer launch.stateMut.Unlock()

	event.Message.CustomerID = launch.User.CustomerId
	if launch.Bastion != nil {
		event.Message.BastionID = launch.Bastion.ID
	}

	if launch.EventChan != nil {
		launch.EventChan <- event
	}

	if event.Err != nil {
		launch.Err = event.Err
	}

	if event.Message.State == stateComplete {
		launch.state = launch.state - 1
	}

	if launch.state == 0 || launch.Err != nil {
		launch.cleanup()
	}
}

func (launch *Launch) event(msg *bus.Message) {
	launch.loggerWithAttributes(msg.Attributes).Info(
		fmt.Sprintf("[%s](%s): %s", msg.Command, msg.State, msg.Message),
	)

	launch.handleEvent(&Event{Message: msg})
}

func (launch *Launch) error(err error, msg *bus.Message) {
	launch.loggerWithAttributes(msg.Attributes).WithError(err).Error(msg.Message)

	msg.State = stateFailed
	launch.handleEvent(&Event{Err: err, Message: msg})
}

func (launch *Launch) loggerWithAttributes(attrs map[string]interface{}) *log.Entry {
	logger := launch.logger
	if attrs != nil {
		logger = logger.WithFields(log.Fields(attrs))
	}

	return logger
}

func (launch *Launch) cleanup() {
	defer func() {
		close(launch.EventChan)
		launch.EventChan = nil
	}()

	var err error
	if launch.createTopicOutput != nil {
		_, err = launch.snsClient.DeleteTopic(&sns.DeleteTopicInput{
			TopicArn: launch.createTopicOutput.TopicArn,
		})

		if err != nil {
			launch.logger.WithError(err).Error("failed to delete sns topic")
		} else {
			launch.logger.Info("cleaned up sns topic")
		}
	}

	if launch.createQueueOutput != nil {
		_, err = launch.sqsClient.DeleteQueue(&sqs.DeleteQueueInput{
			QueueUrl: launch.createQueueOutput.QueueUrl,
		})

		if err != nil {
			launch.logger.WithError(err).Error("failed to delete sqs queue")
		} else {
			launch.logger.Info("cleaned up sqs queue")
		}
	}

	if launch.Err != nil && launch.Bastion != nil {
		err = launch.db.UpdateBastion(launch.Bastion.Fail())

		if err != nil {
			launch.logger.WithError(err).Error("failed to mark bastion object as failed")
		} else {
			launch.logger.Info("marked bastion object as failed")
		}
	}

	if err != nil {
		launch.logger.Error("launch cleanup failed")
	} else {
		launch.logger.Info("launch successfuly cleaned up")
	}
}
