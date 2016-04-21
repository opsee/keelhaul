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
		cz, err := NewTarget(t.rds).Generate()
		assert.NoError(err)
		assert.NotNil(cz[0])
		assert.EqualValues(cz[0].Interval, 60)
		assert.EqualValues(cz[0].Assertions[0].Operand, "0.950")
	}
}
