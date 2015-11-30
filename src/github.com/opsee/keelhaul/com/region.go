package com

type Region struct {
	CustomerID         string    `json:"-" db:"customer_id"`
	Region             string    `json:"region"`
	SupportedPlatforms []*string `json:"supported_platforms"`
	VPCs               []*VPC    `json:"vpcs"`
	Subnets            []*Subnet `json:"subnets"`
}

type VPC struct {
	CidrBlock       *string `json:"cidr_block"`
	DhcpOptionsId   *string `json:"dhcp_options_id"`
	InstanceTenancy *string `json:"instance_tenancy"`
	IsDefault       *bool   `json:"is_default"`
	State           *string `json:"state"`
	VpcId           *string `json:"vpc_id"`
}

type Subnet struct {
	AvailabilityZone        *string `json:"availability_zone"`
	AvailableIpAddressCount *int64  `json:"available_ip_address_count"`
	CidrBlock               *string `json:"cidr_block"`
	DefaultForAz            *bool   `json:"default_for_az"`
	MapPublicIpOnLaunch     *bool   `json:"map_public_ip_on_launch"`
	State                   *string `json:"state"`
	SubnetId                *string `json:"subnet_id"`
	VpcId                   *string `json:"vpc_id"`
}
