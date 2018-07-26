package common_test

//go:generate mockgen -destination=./mock/jaeger.go -package=mock -mock_names=SystemDConn=SystemDConn github.com/cloudtrust/common-healthcheck SystemDConn

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/cloudtrust/common-healthcheck"
	mock "github.com/cloudtrust/common-healthcheck/mock"
	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type jaegerReport struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

func TestJaegerDisabled(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var (
		enabled = false
		url     = s.URL[7:] // strip http:// from URL
		m       = NewJaegerModule(mockSystemDConn, s.Client(), url, enabled)
	)

	var jsonReport, err = m.HealthCheck(context.Background(), "agent")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []jaegerReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "jaeger", r.Name)
	assert.Equal(t, "Deactivated", r.Status)
	assert.Zero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestJaegerAgent(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var (
		enabled = true
		url     = s.URL[7:] // strip http:// from URL
		m       = NewJaegerModule(mockSystemDConn, s.Client(), url, enabled)
		units   = []dbus.UnitStatus{{Name: "agent.service", ActiveState: "active"}}
	)

	mockSystemDConn.EXPECT().ListUnitsByNames([]string{"agent.service"}).Return(units, nil).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "agent")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []jaegerReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "agent systemd unit", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestJaegerCollector(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var (
		enabled = true
		url     = s.URL[7:] // strip http:// from URL
		m       = NewJaegerModule(mockSystemDConn, s.Client(), url, enabled)
	)

	var jsonReport, err = m.HealthCheck(context.Background(), "collector")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []jaegerReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "ping collector", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestJaegerAllChecks(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var (
		enabled = true
		url     = s.URL[7:] // strip http:// from URL
		m       = NewJaegerModule(mockSystemDConn, s.Client(), url, enabled)
		units   = []dbus.UnitStatus{{Name: "agent.service", ActiveState: "active"}}
	)

	mockSystemDConn.EXPECT().ListUnitsByNames([]string{"agent.service"}).Return(units, nil).Times(1)
	var jsonReport, err = m.HealthCheck(context.Background(), "")
	assert.Nil(t, err)

	// Check that the report is a valid json
	var report = []jaegerReport{}
	assert.Nil(t, json.Unmarshal(jsonReport, &report))

	var r = report[0]
	assert.Equal(t, "agent systemd unit", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)

	r = report[1]
	assert.Equal(t, "ping collector", r.Name)
	assert.Equal(t, "OK", r.Status)
	assert.NotZero(t, r.Duration)
	assert.Zero(t, r.Error)
}

func TestJaegerAgentFailure(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var (
		enabled = true
		url     = s.URL[7:] // strip http:// from URL
		m       = NewJaegerModule(mockSystemDConn, s.Client(), url, enabled)
	)

	var tsts = []struct {
		mockUnitsStatus []dbus.UnitStatus
		mockError       error
	}{
		// Systemd conn error
		{nil, fmt.Errorf("fail")},
		// Empty systemd unit list
		{[]dbus.UnitStatus{}, nil},
		// Unit status not 'active'
		{[]dbus.UnitStatus{{ActiveState: "inactive"}}, nil},
	}

	for _, tst := range tsts {
		mockSystemDConn.EXPECT().ListUnitsByNames([]string{"agent.service"}).Return(tst.mockUnitsStatus, tst.mockError).Times(1)
		var jsonReport, err = m.HealthCheck(context.Background(), "agent")
		assert.Nil(t, err)

		// Check that the report is a valid json
		var report = []jaegerReport{}
		assert.Nil(t, json.Unmarshal(jsonReport, &report))

		var r = report[0]
		assert.Equal(t, "agent systemd unit", r.Name)
		assert.Equal(t, "KO", r.Status)
		assert.NotZero(t, r.Duration)
		assert.NotZero(t, r.Error)
	}
}

func TestJaegerUnkownHealthCheck(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockSystemDConn = mock.NewSystemDConn(mockCtrl)

	var s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var (
		enabled         = true
		url             = s.URL[7:] // strip http:// from URL
		healthCheckName = "unknown"
		m               = NewJaegerModule(mockSystemDConn, s.Client(), url, enabled)
	)

	var f = func() {
		m.HealthCheck(context.Background(), healthCheckName)
	}
	assert.Panics(t, f)
}
