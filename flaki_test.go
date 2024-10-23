package common_test

//go:generate mockgen --build_flags=--mod=mod -destination=./mock/flaki.go -package=mock -mock_names=FlakiClient=FlakiClient  github.com/cloudtrust/common-healthcheck FlakiClient

import (
	"context"
	"encoding/json"
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

type flakiReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

func TestFlakiDisabled(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)

	var (
		enabled = false
		m       = NewFlakiModule(mockFlakiClient, enabled)
	)

	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []flakiReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "flaki", r.Name)
	assert.Equal(t, "Deactivated", r.Status)
	assert.Zero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestFlakiPing(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)

	var (
		enabled = true
		m       = NewFlakiModule(mockFlakiClient, enabled)
		id      = strconv.FormatUint(rand.Uint64(), 10)
	)

	mockFlakiClient.EXPECT().NextID(context.Background()).Return(id, nil).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []flakiReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "nextid", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestFlakiAllChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)

	var (
		enabled = true
		m       = NewFlakiModule(mockFlakiClient, enabled)
		id      = strconv.FormatUint(rand.Uint64(), 10)
	)

	mockFlakiClient.EXPECT().NextID(context.Background()).Return(id, nil).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []flakiReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "nextid", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestFlakiFailure(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)

	var (
		enabled = true
		m       = NewFlakiModule(mockFlakiClient, enabled)
	)

	mockFlakiClient.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []flakiReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "nextid", r.Name)
	assert.Equal(t, "KO", r.Status)
	assert.NotZero(t, r.Duration)
	assert.NotZero(t, r.Error)
}

func TestFlakiUnkownHealthCheck(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlakiClient = mock.NewFlakiClient(mockCtrl)

	var (
		enabled         = true
		healthCheckName = "unknown"
		m               = NewFlakiModule(mockFlakiClient, enabled)
	)

	var f = func() {
		m.HealthCheck(context.Background(), healthCheckName)
	}
	assert.Panics(t, f)
}
