package launcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hoisie/mustache"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Notifier interface {
	UserID() int
	NotifyVars() interface{}
}

const (
	emailLaunchTemplate = "discovery-completion"
	emailErrorTemplate  = "discovery-failure"
)

var (
	slackLaunchTemplate *mustache.Template
	slackErrorTemplate  *mustache.Template
)

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
}

func (l *launcher) NotifySuccess(n Notifier) error {
	err := l.notifyEmail(n, emailLaunchTemplate)
	if err != nil {
		return err
	}

	return l.notifySlack(n, slackLaunchTemplate)
}

func (l *launcher) NotifyError(n Notifier) error {
	err := l.notifyEmail(n, emailErrorTemplate)
	if err != nil {
		return err
	}

	return l.notifySlack(n, slackErrorTemplate)
}

func (l *launcher) notifyEmail(n Notifier, template string) error {
	log.Info("requested email notification")

	if l.config.VapeEndpoint == "" {
		log.Warn("not sending email notification since VAPE_ENDPOINT is not set")
		return nil
	}

	requestJSON, err := json.Marshal(map[string]interface{}{
		"user_id":  n.UserID(),
		"template": template,
		"vars":     n.NotifyVars(),
	})

	if err != nil {
		return err
	}

	resp, err := http.Post(l.config.VapeEndpoint, "application/json", bytes.NewBuffer(requestJSON))
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

func (l *launcher) notifySlack(n Notifier, template *mustache.Template) error {
	log.Info("requested slack notification")

	if l.config.SlackEndpoint == "" {
		log.Warn("not sending slack notification since SLACK_ENDPOINT is not set")
		return nil
	}

	templateVars := make(map[string]interface{})

	j, err := json.Marshal(n.NotifyVars())
	if err != nil {
		return err
	}

	err = json.Unmarshal(j, &templateVars)
	if err != nil {
		return err
	}

	body := bytes.NewBufferString(template.Render(templateVars))
	resp, err := http.Post(l.config.SlackEndpoint, "application/json", body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	log.WithField("status", resp.StatusCode).Info("sent slack request")

	return nil
}
