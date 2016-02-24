#!/usr/bin/env bash

# This is an example of how to use the cfn utility to launch a bastion
# it assumes you have VPN_PASSWORD and BASTION_ID in your environment.
CUSTOMER_ID="5963d7bc-6ba2-11e5-8603-6ba085b2f5b5"
CUSTOMER_EMAIL="dan@opsee.co"
BASTION_VERSION="stable"
VPN_REMOTE="bastion.opsee.com"

bin/cfn -bastion_id=$BASTION_ID -role=opsee-role-140c5346-5d57-11e5-9947-9f9fcf62725e -user_id=13 -customer_id=$CUSTOMER_ID -email=$CUSTOMER_EMAIL -password=$VPN_PASSWORD -region=us-west-1 -subnet_id=subnet-0378a966 -vpc_id=vpc-79b1491c -template_path=etc/bastion-cf.template -public=True -delete=false -key_name=bastion-testing
