package slack

import "github.com/hoisie/mustache"

var CheckPassing = `{
  "token": "{{token}}",
  "channel":"{{channel}}",
  "username": "OpseeBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-green.png",
  "attachments": [
    {
      "pretext": "Passing check",
      "title": "{{check_name}} passing in {{group_name}}",
      "title_link": "https://app.opsee.com/check/{{check_id}}{{json_url}}utm_source=notification&utm_medium=slack&utm_campaign=app",
      "text": "{{instance_count}} Passing",
      "color": "#69a92c"
    }
  ]
}
`

func init() {
	tmpl, err := mustache.ParseString(CheckPassing)
	if err != nil {
		panic(err)
	}
	Templates["check-passing"] = tmpl
}
