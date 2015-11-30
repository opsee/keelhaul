package com

import (
	"bytes"
	"text/template"
)

type BastionConfig struct {
	OwnerID       string `json:"owner_id"`
	Tag           string `json:"tag"`
	KeyPair       string `json:"keypair"`
	VPNRemote     string `json:"vpn_remote"`
	DNSServer     string `json:"dns_server"`
	NSQDHost      string `json:"nsqd_host"`
	BartnetHost   string `json:"bartnet_host"`
	AuthType      string `json:"auth_type"`
	ModifiedIndex uint64 `json:"modified_index"`
}

const userdata = `
#cloud-config
write_files:
  path: "/etc/opsee/bastion-env.sh"
  permissions: "0644"
  owner: root
  content: |
    CUSTOMER_ID={{.User.CustomerID}}
    CUSTOMER_EMAIL={{.User.Email}}
    BASTION_VERSION={{.Config.Tag}}
    BASTION_ID={{.Bastion.ID}}
    VPN_PASSWORD={{.Bastion.Password}}
    VPN_REMOTE={{.Config.VPNRemote}}
    DNS_SERVER={{.Config.DNSServer}}
    NSQD_HOST={{.Config.NSQDHost}}
    BARTNET_HOST={{.Config.BartnetHost}}
    BASTION_AUTH_TYPE={{.Config.AuthType}}
coreos:
  update:
    reboot-strategy: etcd-lock
    group: beta
`

var userdataTmpl = template.Must(template.New("userdata").Parse(userdata))

func (config *BastionConfig) GenerateUserData(user *User, bastion *Bastion) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	var ud = struct {
		User    *User
		Config  *BastionConfig
		Bastion *Bastion
	}{
		user,
		config,
		bastion,
	}

	err := userdataTmpl.Execute(buf, ud)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
