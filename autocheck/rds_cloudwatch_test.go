package autocheck

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/stretchr/testify/assert"
	"testing"
)

var rdsTests = []struct {
	rds *rds.DBInstance
}{
	{
		rds: &rds.DBInstance{
			DBInstanceIdentifier: aws.String("opsee-test-db"),
			DBInstanceClass:      aws.String("db.m3.xlarge"),
		},
	},
}

func TestRDSGenerate(t *testing.T) {
	assert := assert.New(t)

	for _, t := range rdsTests {
		_, err := NewTarget(t.rds).Generate()
		assert.NoError(err)
	}
}
