package slack

var LaunchError = `{
  "text": "*BASTION LAUNCH ERROR*",
  "username": "ErrorBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-red.png",
  "attachments": [
    {
      "text": "Customer: {{customer_id}}",
      "color": "#f44336",
      "fields": [
        {
          "title": "User Email",
          "value": "{{user_email}}",
          "short": true
        },
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
          "title": "AMI ID",
          "value": "{{image_id}}",
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
          "value": "{{error}}"
        }
      ]
    }
  ]
}
`
