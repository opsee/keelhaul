package autocheck

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/opsee/basic/schema"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var rdsTests = []struct {
	rds *rds.RDS
}{
	{
		rds: &rds.RDS{
			DBInstanceIdentifier: aws.String("opsee-test-db"),
			DBInstanceClass:      aws.String("db.m3.xlarge"),
		},
	},
}

func TestRDSGenerate(t *testing.T) {
	assert := assert.New(t)

	for _, t := range rdsTests {
		checks, err := NewTarget(t.rds).Generate()
		assert.NoError(err)
	}
}
