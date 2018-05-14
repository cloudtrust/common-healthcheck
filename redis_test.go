package common_test

//go:generate mockgen -destination=./mock/redis.go -package=mock -mock_names=RedisClient=RedisClient  github.com/cloudtrust/common-healthcheck RedisClient

import (
	"context"
	"fmt"
	"testing"
	"time"
	"math/rand"
	"strconv"
 
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

func TestRedisModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRedis = mock.NewRedisClient(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var module = NewRedisModule(mockRedis, true) 
	var m = MakeRedisModuleLoggingMW(mockLogger)(module)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	mockRedis.EXPECT().Do(gomock.Any(), gomock.Any()).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)

	// Without correlation ID.
	mockRedis.EXPECT().Do(gomock.Any(), gomock.Any()).Times(1)
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}

func TestRedisReportMarshalJSON(t *testing.T) {
	var report = &JaegerReport{
		Name:     "Redis",
		Duration: 1 * time.Second,
		Status:   OK,
		Error:    fmt.Errorf("Error"),
	}

	json, err := report.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, "{\"name\":\"Redis\",\"duration\":\"1s\",\"status\":\"OK\",\"error\":\"Error\"}", string(json))
}