package slack

import "github.com/hoisie/mustache"

var NewRoleInstall = `{
  "text": ":bread: :garf: CROSS ACCOUNT ROLL HOT OUT OF THE OVEN :garf: :bread:",
  "username": "CustomerBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-green.png",
  "attachments": [
    {
      "text": "{{user_name}} - {{user_email}}",
      "color": "#81C784"{{#referrer}},
      "fields": [
          {
              "title": "Referrer",
              "value": "{{referrer}}",
              "short": true
          }
          ]{{/referrer}}
    }
  ]
}
`

func init() {
	tmpl, err := mustache.ParseString(NewRoleInstall)
	if err != nil {
		panic(err)
	}
	Templates["new-role-install"] = tmpl
}
