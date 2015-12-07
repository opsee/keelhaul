-- // subnet has no route to the internet
-- RoutingStatePrivate = "private"
-- // subnet has a route to the internet via an aws internet gateway
-- RoutingStatePublic = "public"
-- // subnet has a route to the internet via a NAT instance
-- RoutingStateNAT = "nat"
-- // subnet has a route to the internet via a customer gateway
-- RoutingStateGateway = "gateway"
-- // subnet may have a route to the internet, but can't communicate with
-- // 100% of instances in a VPC
-- RoutingStateOccluded = "occluded"

create type bastion_routing as enum ('private', 'public', 'nat', 'gateway', 'occluded');

alter table bastions add column subnet_routing bastion_routing not null default 'private';
