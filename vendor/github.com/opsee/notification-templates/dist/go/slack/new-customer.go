package slack
var NewCustomer = `{
  "text": "*NEW CUSTOMER*",
  "username": "CustomerBot",
  "icon_url": "https://s3-us-west-1.amazonaws.com/opsee-public-images/slack-avi-48-green.png",
  "attachments": [
    {
      "text": "Customer: {{customer_id}}",
      "color": "#81C784",
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
          "title": "VPC ID",
          "value": "{{vpc_id}}",
          "short": true  
        },
        {
          "title": "Subnet ID",
          "value": "{{subnet_id}}",
          "short": true  
        },
        {
          "title": "Instance ID",
          "value": "{{instance_id}}",
          "short": true  
        },
        {
          "title": "Group ID",
          "value": "{{group_id}}",
          "short": true  
        },
        {
          "title": "EC2 Instances",
          "value": "{{instance_count}}",
          "short": true
        },
        {
          "title": "RDS Instances",
          "value": "{{db_instance_count}}",
          "short": true
        },
        {
          "title": "Security Groups",
          "value": "{{security_group_count}}",
          "short": true
        },
        {
          "title": "RDS Security Groups",
          "value": "{{db_security_group_count}}",
          "short": true
        },
        {
          "title": "Load Balancers",
          "value": "{{load_balancer_count}}",
          "short": true
        },
        {
          "title": "Autoscaling Groups",
          "value": "{{autoscaling_group_count}}",
          "short": true
        },
        {
          "title": "Opsee Created Checks",
          "value": "{{check_count}}",
          "short": true
        }
      ]
    }
  ]
}
`
