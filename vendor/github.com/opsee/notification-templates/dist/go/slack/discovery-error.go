package slack
var DiscoveryError = `{
  "text": "*DISCOVERY ERROR*",
  "username": "ErrorBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-red.png",
  "attachments": [
    {
      "text": "Customer: {{customer_id}}",
      "color": "#f44336",
      "fields": [
        {
          "title": "User ID",
          "value": "{{user_id}}",
          "short": true
        },
        {
          "title": "AWS Region",
          "value": "{{region}}",
          "short": true
        },
        {
          "title": "Instance Errors",
          "value": "{{instance_error_count}}",
          "short": true
        },
        {
          "title": "Group Errors",
          "value": "{{group_error_count}}",
          "short": true
        },
        {
          "title": "Last Error",
          "value": "{{last_error}}"
        }
      ]
    }
  ]
}
`
