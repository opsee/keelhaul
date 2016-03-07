package launcher

import "testing"

func TestGetIngressTemplateUrl(t *testing.T) {
	east1 := ingressTemplateUrl("us-east-1", "opsee-bastion-cf", "beta/opsee-bastion-cf.template")
	if east1 != "https://s3.amazonaws.com/opsee-bastion-cf/beta/opsee-bastion-cf.template" {
		t.FailNow()
	}
	west1 := ingressTemplateUrl("us-west-1", "opsee-bastion-cf", "beta/opsee-bastion-cf.template")
	t.Log(west1)
	if west1 != "https://s3-us-west-1.amazonaws.com/opsee-bastion-cf/beta/opsee-bastion-cf.template" {
		t.FailNow()
	}
}
