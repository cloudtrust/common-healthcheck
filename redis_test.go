package common_test

//go:generate mockgen -destination=./mock/redis.go -package=mock -mock_names=RedisClient=RedisClient github.com/cloudtrust/common-healthcheck RedisClient

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/cloudtrust/common-healthcheck"
	mock "github.com/cloudtrust/common-healthcheck/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type redisReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

func TestRedisDisabled(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRedis = mock.NewRedisClient(mockCtrl)

	var (
		enabled = false
		m       = NewRedisModule(mockRedis, enabled)
	)

	var report, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = redisReport{}
	assert.Nil(t, json.Unmarshal(report, &r))
	assert.Equal(t, "redis", r.Name)
	assert.Equal(t, "Deactivated", r.Status)
	assert.Zero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestRedisPing(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRedis = mock.NewRedisClient(mockCtrl)

	var (
		enabled = true
		m       = NewRedisModule(mockRedis, enabled)
	)

	mockRedis.EXPECT().Do("PING").Return(nil, nil).Times(1)
	var report, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = redisReport{}
	assert.Nil(t, json.Unmarshal(report, &r))
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestRedisAllChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRedis = mock.NewRedisClient(mockCtrl)

	var (
		enabled = true
		m       = NewRedisModule(mockRedis, enabled)
	)

	mockRedis.EXPECT().Do("PING").Return(nil, nil).Times(1)
	var report, err = m.HealthCheck(context.Background(), "")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = []redisReport{}
	assert.Nil(t, json.Unmarshal(report, &r))

	var pingReport = r[0]
	assert.Equal(t, "ping", pingReport.Name)
	assert.Equal(t, "OK", pingReport.Status)
	assert.NotZero(t, pingReport.Duration)
	assert.Zero(t, pingReport.Error)
}

func TestRedisFailure(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRedis = mock.NewRedisClient(mockCtrl)

	var (
		enabled = true
		m       = NewRedisModule(mockRedis, enabled)
	)

	mockRedis.EXPECT().Do("PING").Return(nil, fmt.Errorf("fail")).Times(1)
	var report, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = redisReport{}
	assert.Nil(t, json.Unmarshal(report, &r))
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "KO", r.Status)
	assert.NotZero(t, r.Duration)
	assert.NotZero(t, r.Error)
}

func TestRedisUnkownHealthCheck(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRedis = mock.NewRedisClient(mockCtrl)

	var (
		enabled         = true
		healthCheckName = "unknown"
		m               = NewRedisModule(mockRedis, enabled)
	)

	var f = func() {
		m.HealthCheck(context.Background(), healthCheckName)
	}
	assert.Panics(t, f)
}
