package common_test

//go:generate mockgen -destination=./mock/redis.go -package=mock -mock_names=RedisClient=RedisClient  github.com/cloudtrust/common-healthcheck RedisClient

import (
	"context"
	"fmt"
	"testing"
 
	. "github.com/cloudtrust/common-healthcheck"
	mock "github.com/cloudtrust/common-healthcheck/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRedisHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRedis = mock.NewRedisClient(mockCtrl)

	var m = NewRedisModule(mockRedis, true) 

	// HealthChecks
	{
		mockRedis.EXPECT().Do("PING").Return(nil, nil).Times(1)
		var report = m.HealthChecks(context.Background())[0]
		assert.Equal(t, "ping", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, OK, report.Status)
		assert.Zero(t, report.Error)
	}

	// Redis fail.
	{
		mockRedis.EXPECT().Do("PING").Return(nil, fmt.Errorf("fail")).Times(1)
		var report = m.HealthChecks(context.Background())[0]
		assert.Equal(t, "ping", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, KO, report.Status)
		assert.NotZero(t, report.Error)
	}
}

func TestNoopRedisHealthChecks(t *testing.T) {
	var m = NewRedisModule(nil, false)

	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "ping", report.Name)
	assert.Zero(t, report.Duration)
	assert.Equal(t, Deactivated, report.Status)
	assert.Zero(t, report.Error)
}
