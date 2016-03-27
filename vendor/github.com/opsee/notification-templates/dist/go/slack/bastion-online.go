package slack
var BastionOnline = `{
  "text": "*BASTION ONLINE*",
  "username": "BastionTracker",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-green.png",
  "attachments": [
    {
      "color": "#81C784",
      "fields": [
        {
          "title": "User",
          "value": "{{name}}",
          "short": true
        },
        {
          "title": "Email",
          "value": "{{email}}",
          "short": true
        },
        {
          "title": "Customer ID",
          "value": "{{customer_id}}",
          "short": true
        },
        {
          "title": "Bastion ID",
          "value": "{{bastion_id}}",
          "short": true
        },
        {
          "title": "Last Seen",
          "value": "{{last_seen}}",
          "short": true
        },
        {
          "title": "Current State",
          "value": "{{current_state}}",
          "short": true
        }
      ]
    }
  ]
}
`
