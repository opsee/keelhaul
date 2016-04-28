package service

import (
	"errors"
)

var (
	errMissingUser          = errors.New("no user provided")
	errMissingVpc           = errors.New("no vpc id provided")
	errMissingSubnet        = errors.New("no subnet id provided")
	errMissingSubnetRouting = errors.New("no subnet routing provided")
	errUnauthorized         = errors.New("unauthorized.")
	errAWSUnauthorized      = errors.New("Your AWS credentials could not be validated, please check to ensure they are correct.")
	errMissingAccessKey     = errors.New("missing access_key.")
	errMissingSecretKey     = errors.New("missing secret_key.")
	errMissingRegion        = errors.New("missing region.")
	errBadRequest           = errors.New("bad request.")
	errUnknown              = errors.New("unknown error.")
)
