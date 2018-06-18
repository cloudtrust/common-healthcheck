package common_test

//go:generate mockgen -destination=./mock/influx.go -package=mock -mock_names=Influx=Influx  github.com/cloudtrust/common-healthcheck Influx

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/common-healthcheck"
	mock "github.com/cloudtrust/common-healthcheck/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInfluxHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInflux(mockCtrl)

	var m = NewInfluxModule(mockInflux, true)

	// HealthChecks.
	{
		mockInflux.EXPECT().Ping(5*time.Second).Return(1*time.Second, "", nil).Times(1)
		var report = m.HealthChecks(context.Background())[0]
		assert.Equal(t, "ping", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, OK, report.Status)
		assert.Zero(t, report.Error)
	}

	// HealthChecks error.
	{
		mockInflux.EXPECT().Ping(5*time.Second).Return(0*time.Second, "", fmt.Errorf("fail")).Times(1)
		var report = m.HealthChecks(context.Background())[0]
		assert.Equal(t, "ping", report.Name)
		assert.NotZero(t, report.Duration)
		assert.Equal(t, KO, report.Status)
		assert.NotZero(t, report.Error)
	}
}

func TestNoopInfluxHealthChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInflux(mockCtrl)

	var m = NewInfluxModule(mockInflux, false)

	// HealthChecks.
	var report = m.HealthChecks(context.Background())[0]
	assert.Equal(t, "influx", report.Name)
	assert.Zero(t, report.Duration)
	assert.Equal(t, Deactivated, report.Status)
	assert.Zero(t, report.Error)
}

func TestInfluxModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInflux(mockCtrl)

	var module = NewInfluxModule(mockInflux, true)
	var mockLogger = mock.NewLogger(mockCtrl)

	var m = MakeInfluxModuleLoggingMW(mockLogger)(module)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	mockInflux.EXPECT().Ping(gomock.Any()).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthChecks", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthChecks(ctx)

	// Without correlation ID.
	mockInflux.EXPECT().Ping(gomock.Any()).Times(1)
	var f = func() {
		m.HealthChecks(context.Background())
	}
	assert.Panics(t, f)
}

func TestInfluxReportMarshalJSON(t *testing.T) {
	var report = &InfluxReport{
		Name:     "Influx",
		Duration: 1 * time.Second,
		Status:   OK,
		Error:    fmt.Errorf("Error"),
	}

	json, err := report.MarshalJSON()

	assert.Nil(t, err)
	assert.Equal(t, "{\"name\":\"Influx\",\"duration\":\"1s\",\"status\":\"OK\",\"error\":\"Error\"}", string(json))
}
