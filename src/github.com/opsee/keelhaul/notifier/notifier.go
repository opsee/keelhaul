package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hoisie/mustache"
	"github.com/opsee/keelhaul/config"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Notifier interface {
	NotifyError(int, interface{}) error
	NotifySuccess(int, interface{}) error
	NotifySlackBastionState(bool, string, map[string]interface{}) error
}

type notifier struct {
	LaunchesSlackEndpoint string
	TrackerSlackEndpoint  string
	VapeEmailEndpoint     string
	VapeUserInfoEndpoint  string
}

const (
	emailLaunchTemplate = "discovery-completion"
	emailErrorTemplate  = "discovery-failure"
)

var (
	slackLaunchTemplate      *mustache.Template
	slackErrorTemplate       *mustache.Template
	slackBastionUpTemplate   *mustache.Template
	slackBastionDownTemplate *mustache.Template
)

func New(c *config.Config) *notifier {
	return &notifier{
		VapeEmailEndpoint:     c.VapeEmailEndpoint,
		VapeUserInfoEndpoint:  c.VapeUserInfoEndpoint,
		LaunchesSlackEndpoint: c.LaunchesSlackEndpoint,
		TrackerSlackEndpoint:  c.TrackerSlackEndpoint,
	}
}

func init() {
	tmpl, err := mustache.ParseString(slacktmpl.NewCustomer)
	if err != nil {
		panic(err)
	}
	slackLaunchTemplate = tmpl

	tmpl, err = mustache.ParseString(slacktmpl.LaunchError)
	if err != nil {
		panic(err)
	}
	slackErrorTemplate = tmpl

	tmpl, err = mustache.ParseString(slacktmpl.BastionOnline)
	if err != nil {
		panic(err)
	}
	slackBastionUpTemplate = tmpl

	tmpl, err = mustache.ParseString(slacktmpl.BastionOffline)
	if err != nil {
		panic(err)
	}
	slackBastionDownTemplate = tmpl
}

func (n *notifier) NotifySlackBastionState(isUp bool, custID string, notifyVars map[string]interface{}) error {
	resp, err := http.Get(n.VapeUserInfoEndpoint + "/" + custID)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("Bad response from Vape user info endpoint: %s", resp.Status)
	}

	response := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&response)
	if err != nil {
		return err
	}

	_, ok := response["email"]
	if !ok {
		return fmt.Errorf("error response from vape")
	}

	notifyVars["email"] = response["email"]
	notifyVars["name"] = response["name"]

	if isUp {
		return n.notifySlack(notifyVars, slackBastionUpTemplate, n.TrackerSlackEndpoint)
	}
	return n.notifySlack(notifyVars, slackBastionDownTemplate, n.TrackerSlackEndpoint)
}

func (n *notifier) NotifySuccess(userID int, notifyVars interface{}) error {
	err := n.notifyEmail(userID, notifyVars, emailLaunchTemplate)
	if err != nil {
		return err
	}

	return n.notifySlack(notifyVars, slackLaunchTemplate, n.LaunchesSlackEndpoint)
}

func (n *notifier) NotifyError(userID int, notifyVars interface{}) error {
	err := n.notifyEmail(userID, notifyVars, emailErrorTemplate)
	if err != nil {
		return err
	}

	return n.notifySlack(notifyVars, slackErrorTemplate, n.LaunchesSlackEndpoint)
}

func (n *notifier) notifyEmail(userID int, notifyVars interface{}, template string) error {
	log.Info("requested email notification")

	requestJSON, err := json.Marshal(map[string]interface{}{
		"user_id":  userID,
		"template": template,
		"vars":     notifyVars,
	})

	if err != nil {
		return err
	}

	resp, err := http.Post(n.VapeEmailEndpoint, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	log.WithField("status", resp.StatusCode).Info("sent vape request")

	if resp.StatusCode > 299 {
		return fmt.Errorf("Bad response from Vape notification endpoint: %s", resp.Status)
	}

	response := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&response)
	if err != nil {
		return err
	}

	_, ok := response["user"]
	if !ok {
		return fmt.Errorf("error response from vape")
	}

	log.Info("user response from vape")

	return nil
}

func (n *notifier) notifySlack(notifyVars interface{}, template *mustache.Template, endpoint string) error {

	templateVars := make(map[string]interface{})

	j, err := json.Marshal(notifyVars)
	if err != nil {
		return err
	}

	err = json.Unmarshal(j, &templateVars)
	if err != nil {
		return err
	}

	body := bytes.NewBufferString(template.Render(templateVars))
	resp, err := http.Post(endpoint, "application/json", body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	log.WithField("status", resp.StatusCode).Info("sent slack request")

	return nil
}
