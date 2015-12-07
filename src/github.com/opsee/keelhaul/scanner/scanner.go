package scanner

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/opsee/keelhaul/com"
	"github.com/opsee/keelhaul/util"
	"net"
	"sort"
)

const (
	// this is necessary due to the aws api being wrong
	attachmentStatusAvailable = "available"
	theInternet               = "0.0.0.0/0"
)

func ScanRegion(region string, session *session.Session) (*com.Region, error) {
	ec2Client := ec2.New(session)

	var (
		nextToken      *string
		vpcIPs         = make(map[string][]string)
		vpcRouteTables = make(map[string][]*ec2.RouteTable)
		vpcGateways    = make(map[string][]*ec2.InternetGateway)
		vpcCounts      = make(map[string]int)
		subnetCounts   = make(map[string]int)
	)

	for {
		instancesOutput, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		})

		if err != nil {
			return nil, err
		}

		nextToken = instancesOutput.NextToken
		for _, res := range instancesOutput.Reservations {
			for _, instance := range res.Instances {
				if *instance.State.Name != ec2.InstanceStateNameTerminated {
					vips, ok := vpcIPs[*instance.VpcId]
					if !ok {
						vips = make([]string, 0)
					}

					vpcIPs[*instance.VpcId] = append(vips, *instance.PrivateIpAddress)
					vpcCounts[*instance.VpcId]++
					subnetCounts[*instance.SubnetId]++
				}
			}
		}

		if nextToken == nil {
			break
		}
	}

	internetGatewaysOutput, err := ec2Client.DescribeInternetGateways(nil)
	if err != nil {
		return nil, err
	}

	for _, igw := range internetGatewaysOutput.InternetGateways {
		for _, igwatt := range igw.Attachments {
			st := *igwatt.State
			if st == attachmentStatusAvailable || st == ec2.AttachmentStatusAttached {
				vigw, ok := vpcGateways[*igwatt.VpcId]
				if !ok {
					vigw = make([]*ec2.InternetGateway, 0)
				}

				vpcGateways[*igwatt.VpcId] = append(vigw, igw)
			}
		}
	}

	routeTablesOutput, err := ec2Client.DescribeRouteTables(nil)
	if err != nil {
		return nil, err
	}

	for _, rt := range routeTablesOutput.RouteTables {
		vrts, ok := vpcRouteTables[*rt.VpcId]
		if !ok {
			vrts = make([]*ec2.RouteTable, 0)
		}

		vpcRouteTables[*rt.VpcId] = append(vrts, rt)
	}

	vpcOutput, err := ec2Client.DescribeVpcs(nil)
	if err != nil {
		return nil, err
	}

	vpcs := make([]*com.VPC, len(vpcOutput.Vpcs))
	for vi, v := range vpcOutput.Vpcs {
		tags := make([]*com.Tag, len(v.Tags))
		copyTags(tags, v.Tags)

		vpc := &com.VPC{}
		util.Copy(vpc, v)

		vpc.Tags = tags
		vpc.InstanceCount = vpcCounts[*vpc.VpcId]
		vpcs[vi] = vpc
	}

	subnetOutput, err := ec2Client.DescribeSubnets(nil)
	if err != nil {
		return nil, err
	}

	subnets := make([]*com.Subnet, len(subnetOutput.Subnets))
	for si, s := range subnetOutput.Subnets {
		tags := make([]*com.Tag, len(s.Tags))
		copyTags(tags, s.Tags)

		subnet := &com.Subnet{}
		util.Copy(subnet, s)

		subnet.Tags = tags
		subnet.InstanceCount = subnetCounts[*subnet.SubnetId]

		routing, err := determineRouting(s, vpcRouteTables[*s.VpcId], vpcIPs[*s.VpcId], vpcGateways[*s.VpcId])
		if err != nil {
			return nil, err
		}

		subnet.Routing = routing
		subnets[si] = subnet
	}

	// now sort them in the order we want to select for bastion install
	sort.Sort(com.SubnetsByPreference(subnets))

	accountOutput, err := ec2Client.DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{
		AttributeNames: []*string{
			aws.String("supported-platforms"),
		},
	})
	if err != nil {
		return nil, err
	}

	supportedPlatforms := make([]*string, 0)
	for _, a := range accountOutput.AccountAttributes {
		if *a.AttributeName == "supported-platforms" {
			for _, v := range a.AttributeValues {
				supportedPlatforms = append(supportedPlatforms, v.AttributeValue)
			}
		}
	}

	return &com.Region{
		Region:             region,
		SupportedPlatforms: supportedPlatforms,
		VPCs:               vpcs,
		Subnets:            subnets,
	}, nil
}

func determineRouting(subnet *ec2.Subnet, routeTables []*ec2.RouteTable, instanceIPs []string, gateways []*ec2.InternetGateway) (string, error) {
	// RouteTable.Associations.{Main(*bool),RouteTableId(*string),SubnetId(*string)}
	var (
		associatedTable *ec2.RouteTable
		mainTable       *ec2.RouteTable
		internetRoute   *ec2.Route
		cidrs           = make([]string, 0)
		routeToIPs      = make(map[string]bool)
	)

	for _, rt := range routeTables {
		for _, asso := range rt.Associations {
			if *asso.Main {
				mainTable = rt
			}

			if asso.SubnetId != nil && *asso.SubnetId == *subnet.SubnetId {
				associatedTable = rt
			}
		}
	}

	if associatedTable == nil {
		associatedTable = mainTable
	}

	for _, route := range associatedTable.Routes {
		if *route.State != ec2.RouteStateActive {
			continue
		}

		if *route.DestinationCidrBlock == theInternet {
			internetRoute = route
			continue
		}

		cidrs = append(cidrs, *route.DestinationCidrBlock)
	}

	// no route to 0.0.0.0/0, so we're private. no need to check if
	// we have a route to other instances
	if internetRoute == nil {
		return com.RoutingStatePrivate, nil
	}

	// TODO: scan network ACLs here? YES. going to push this out first tho

	// verify that we can reach all of the instance ips? i say yes
	for _, cid := range cidrs {
		_, network, err := net.ParseCIDR(cid)
		if err != nil {
			return "", err
		}

		for _, ip := range instanceIPs {
			if network.Contains(net.ParseIP(ip)) {
				routeToIPs[ip] = true
			}
		}
	}

	// we must not be able to reach all the instance ips,
	// so we're going to mark this subnet as occluded
	if len(routeToIPs) < len(instanceIPs) {
		return com.RoutingStateOccluded, nil
	}

	// nat. pretty straightforward i guess,
	// going to punt on verifying that the nat instance itself has a route?
	if internetRoute.InstanceId != nil {
		return com.RoutingStateNAT, nil
	}

	// public _or_ routing through a customer gateway, in which case we'll consider it NAT,
	// since the only case we'll need a public ip is if we're going through an internet gateway
	if internetRoute.GatewayId != nil {
		for _, igw := range gateways {
			if *internetRoute.GatewayId == *igw.InternetGatewayId {
				return com.RoutingStatePublic, nil
			}
		}

		return com.RoutingStateGateway, nil
	}

	// we're in a weird state. the routing table is possibly attached to an eni, but not an instance
	return com.RoutingStateOccluded, nil
}

func copyTags(tags []*com.Tag, eTags []*ec2.Tag) {
	for i, t := range eTags {
		tags[i] = &com.Tag{
			Key:   t.Key,
			Value: t.Value,
		}
	}
}
