package slack
var CheckPassing = `{
  "token": "{{token}}",
  "channel":"{{channel}}",
  "username": "OpseeBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-green.png",
  "attachments": [
    {
      "pretext": "Passing check",
      "title": "{{check_name}} passing in {{group_name}}",
      "title_link": "https://app.opsee.com/check/{{check_id}}",
      "text": "{{instance_count}} Passing",
      "color": "#69a92c"
    }
  ]
}
`
