package slack

import "github.com/hoisie/mustache"

var CheckFailing = `{
  "token": "{{token}}",
  "channel":"{{channel}}",
  "username": "OpseeBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-red.png",
  "attachments": [
    {
      "pretext": "Failing check",
      "title": "{{check_name}} failing in {{group_name}}",
      "title_link": "https://app.opsee.com/check/{{check_id}}{{json_url}}utm_source=notification&utm_medium=slack&utm_campaign=app",
      "text": "{{fail_count}} of {{instance_count}} Failing",
      "color": "#f44336"
    }
  ]
}
`

func init() {
	tmpl, err := mustache.ParseString(CheckFailing)
	if err != nil {
		panic(err)
	}
	Templates["check-failing"] = tmpl
}
