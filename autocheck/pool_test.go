package autocheck

import (
	"github.com/opsee/basic/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testSink struct{}

func (s *testSink) Send(check *schema.Check) error {
	return nil
}

func TestPool(t *testing.T) {
	assert := assert.New(t)
	expected := 0

	pool := NewPool(&testSink{}, nil)
	for _, test := range elbtests {
		pool.AddTarget(test.elb)
		expected = expected + len(test.checks)
	}
	pool.Drain()

	assert.Equal(expected, pool.SuccessCount())
}
