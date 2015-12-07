package launcher

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	etcd "github.com/coreos/etcd/client"
	"github.com/opsee/keelhaul/com"
	"github.com/opsee/keelhaul/config"
	"github.com/opsee/keelhaul/router"
	"github.com/opsee/keelhaul/store"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Stage interface {
	Execute(*Launch)
}

type Launch struct {
	Bastion                  *com.Bastion
	User                     *com.User
	EventChan                chan *Event
	Err                      error
	VPCEnvironment           *VPCEnvironment
	state                    int
	stateMut                 *sync.RWMutex
	session                  *session.Session
	logger                   *log.Entry
	db                       store.Store
	router                   router.Router
	etcd                     etcd.KeysAPI
	config                   *config.Config
	bastionConfig            *com.BastionConfig
	imageID                  string
	sqsClient                sqsiface.SQSAPI
	snsClient                snsiface.SNSAPI
	cloudformationClient     cloudformationiface.CloudFormationAPI
	subscribeOutput          *sns.SubscribeOutput
	createTopicOutput        *sns.CreateTopicOutput
	createQueueOutput        *sqs.CreateQueueOutput
	getQueueAttributesOutput *sqs.GetQueueAttributesOutput
	setQueueAttributesOutput *sqs.SetQueueAttributesOutput
	createStackOutput        *cloudformation.CreateStackOutput
	connectAttempts          float64
}

type Event struct {
	Err     error
	Message *com.Message
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

const (
	commandLaunchBastion  = "launch-bastion"
	commandConnectBastion = "connect-bastion"
	commandDiscovery      = "discovery"

	stateInProgress = "in-progress"
	stateComplete   = "complete"
	stateFailed     = "failed"
)

func NewLaunch(db store.Store, router router.Router, etcdKAPI etcd.KeysAPI, cfg *config.Config, sess *session.Session, user *com.User) *Launch {
	return &Launch{
		User:           user,
		EventChan:      make(chan *Event),
		VPCEnvironment: &VPCEnvironment{},
		state:          3,
		stateMut:       &sync.RWMutex{},
		db:             db,
		router:         router,
		etcd:           etcdKAPI,
		config:         cfg,
		session:        sess,
		logger: log.WithFields(log.Fields{
			"customer-id": user.CustomerID,
			"user-id":     user.ID,
		}),
		sqsClient:            sqs.New(sess),
		snsClient:            sns.New(sess),
		cloudformationClient: cloudformation.New(sess),
		connectAttempts:      float64(1),
	}
}

func (launch *Launch) UserID() int {
	return launch.User.ID
}

func (launch *Launch) NotifyVars() interface{} {
	vars := struct {
		*VPCEnvironment
		UserID     int    `json:"user_id"`
		UserEmail  string `json:"user_email"`
		CustomerID string `json:"customer_id"`
		Region     string `json:"region"`
		ImageID    string `json:"image_id"`
		VPCID      string `json:"vpc_id"`
		SubnetID   string `json:"subnet_id"`
		InstanceID string `json:"instance_id"`
		GroupID    string `json:"group_id"`
		Error      string `json:"error"`
	}{
		VPCEnvironment: launch.VPCEnvironment,
		UserID:         launch.User.ID,
		UserEmail:      launch.User.Email,
		CustomerID:     launch.User.CustomerID,
		Region:         *launch.session.Config.Region,
		VPCID:          launch.Bastion.VPCID,
		SubnetID:       launch.Bastion.SubnetID,
		ImageID:        launch.Bastion.ImageID.String,
		InstanceID:     launch.Bastion.InstanceID.String,
		GroupID:        launch.Bastion.GroupID.String,
	}

	if launch.Err != nil {
		vars.Error = launch.Err.Error()
	}

	return vars
}

// these events happen synchronously in the request cycle, so they are not part of launch stages
func (launch *Launch) CreateBastion(vpcID, subnetID, subnetRouting, instanceType string) error {
	bastion, err := com.NewBastion(launch.User.ID, launch.User.CustomerID, vpcID, subnetID, subnetRouting, instanceType)
	if err != nil {
		launch.error(err, &com.Message{
			Command: commandLaunchBastion,
			Message: "failed creating bastion credentials",
		})
		return err
	}

	err = launch.db.PutBastion(bastion)
	if err != nil {
		launch.error(err, &com.Message{
			Command: commandLaunchBastion,
			Message: "failed saving bastion in database",
		})
		return err
	}

	// fold the bastion id into the logger k/v
	launch.logger = launch.logger.WithField("bastion-id", bastion.ID)
	launch.Bastion = bastion
	launch.event(&com.Message{
		State:   stateInProgress,
		Command: commandLaunchBastion,
		Message: "created bastion object",
	})

	return nil
}

func (launch *Launch) Launch() {
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

func (launch *Launch) stage(st Stage) {
	if launch.State() == stateFailed {
		return
	}

	st.Execute(launch)
}

func (launch *Launch) handleEvent(event *Event) {
	launch.stateMut.Lock()
	defer launch.stateMut.Unlock()

	event.Message.CustomerID = launch.User.CustomerID
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

func (launch *Launch) event(msg *com.Message) {
	launch.loggerWithAttributes(msg.Attributes).Info(
		fmt.Sprintf("[%s](%s): %s", msg.Command, msg.State, msg.Message),
	)

	launch.handleEvent(&Event{Message: msg})
}

func (launch *Launch) error(err error, msg *com.Message) {
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
