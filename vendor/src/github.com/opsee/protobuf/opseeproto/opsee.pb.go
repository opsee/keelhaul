// Code generated by protoc-gen-gogo.
// source: opsee.proto
// DO NOT EDIT!

/*
Package opseeproto is a generated protocol buffer package.

It is generated from these files:
	opsee.proto

It has these top-level messages:
*/
package opseeproto

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

var E_Graphql = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FileOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         64040,
	Name:          "opseeproto.graphql",
	Tag:           "varint,64040,opt,name=graphql",
}

var E_Required = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         60000,
	Name:          "opseeproto.required",
	Tag:           "varint,60000,opt,name=required",
}

func init() {
	proto.RegisterExtension(E_Graphql)
	proto.RegisterExtension(E_Required)
}
