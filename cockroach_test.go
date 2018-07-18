package common_test

//go:generate mockgen -destination=./mock/cockroach.go -package=mock -mock_names=CockroachClient=CockroachClient  github.com/cloudtrust/common-healthcheck CockroachClient

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

type cockroachReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

func TestCockroachDisabled(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroachClient(mockCtrl)

	var (
		enabled = false
		m       = NewCockroachModule(mockCockroach, enabled)
	)

	var report, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = cockroachReport{}
	assert.Nil(t, json.Unmarshal(report, &r))
	assert.Equal(t, "cockroach", r.Name)
	assert.Equal(t, "Deactivated", r.Status)
	assert.Zero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestCockroachPing(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroachClient(mockCtrl)

	var (
		enabled = true
		m       = NewCockroachModule(mockCockroach, enabled)
	)

	mockCockroach.EXPECT().Ping().Return(nil).Times(1)
	var report, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = cockroachReport{}
	assert.Nil(t, json.Unmarshal(report, &r))
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestCockroachAllChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroachClient(mockCtrl)

	var (
		enabled = true
		m       = NewCockroachModule(mockCockroach, enabled)
	)

	mockCockroach.EXPECT().Ping().Return(nil).Times(1)
	var report, err = m.HealthCheck(context.Background(), "")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = []cockroachReport{}
	assert.Nil(t, json.Unmarshal(report, &r))

	var pingReport = r[0]
	assert.Equal(t, "ping", pingReport.Name)
	assert.Equal(t, "OK", pingReport.Status)
	assert.NotZero(t, pingReport.Duration)
	assert.Zero(t, pingReport.Error)
}

func TestCockroachFailure(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroachClient(mockCtrl)

	var (
		enabled = true
		m       = NewCockroachModule(mockCockroach, enabled)
	)

	mockCockroach.EXPECT().Ping().Return(fmt.Errorf("fail")).Times(1)
	var report, err = m.HealthCheck(context.Background(), "ping")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var r = cockroachReport{}
	assert.Nil(t, json.Unmarshal(report, &r))
	assert.Equal(t, "ping", r.Name)
	assert.Equal(t, "KO", r.Status)
	assert.NotZero(t, r.Duration)
	assert.NotZero(t, r.Error)
}

func TestCockroachUnkownHealthCheck(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockCockroach = mock.NewCockroachClient(mockCtrl)

	var (
		enabled         = true
		healthCheckName = "unknown"
		m               = NewCockroachModule(mockCockroach, enabled)
	)

	var f = func() {
		m.HealthCheck(context.Background(), healthCheckName)
	}
	assert.Panics(t, f)
}
