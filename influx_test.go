package common_test

//go:generate mockgen --build_flags=--mod=mod -destination=./mock/influx.go -package=mock -mock_names=InfluxClient=InfluxClient  github.com/cloudtrust/common-healthcheck InfluxClient

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

type influxReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

func TestInfluxDisabled(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInfluxClient(mockCtrl)

	var (
		enabled = false
		m       = NewInfluxModule(mockInflux, enabled)
	)

	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []influxReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "influx", r.Name)
	assert.Equal(t, "Deactivated", r.Status)
	assert.Zero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestInfluxPing(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInfluxClient(mockCtrl)

	var (
		enabled = true
		m       = NewInfluxModule(mockInflux, enabled)
		d       = 1 * time.Second
	)

	mockInflux.EXPECT().Ping(5*time.Second).Return(d, "", nil).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []influxReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestInfluxAllChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInfluxClient(mockCtrl)

	var (
		enabled = true
		m       = NewInfluxModule(mockInflux, enabled)
		d       = 1 * time.Second
	)

	mockInflux.EXPECT().Ping(5*time.Second).Return(d, "", nil).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []influxReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestInfluxFailure(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInfluxClient(mockCtrl)

	var (
		enabled = true
		m       = NewInfluxModule(mockInflux, enabled)
		d       = 0 * time.Second
	)

	mockInflux.EXPECT().Ping(5*time.Second).Return(d, "", fmt.Errorf("fail")).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []influxReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "KO", r.Status)
	assert.NotZero(t, r.Duration)
	assert.NotZero(t, r.Error)
}

func TestInfluxUnkownHealthCheck(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockInflux = mock.NewInfluxClient(mockCtrl)

	var (
		enabled         = true
		healthCheckName = "unknown"
		m               = NewInfluxModule(mockInflux, enabled)
	)

	var f = func() {
		m.HealthCheck(context.Background(), healthCheckName)
	}
	assert.Panics(t, f)
}
