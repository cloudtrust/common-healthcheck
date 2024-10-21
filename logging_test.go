package common_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "github.com/cloudtrust/common-healthcheck"
	"github.com/cloudtrust/common-healthcheck/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen --build_flags=--mod=mod -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockHealthChecker = mock.NewHealthChecker(mockCtrl)

	var (
		m               = MakeHealthCheckerLoggingMW(mockLogger)(mockHealthChecker)
		corrID          = strconv.FormatUint(rand.Uint64(), 10)
		ctx             = context.WithValue(context.Background(), "correlation_id", corrID)
		healthCheckName = "name"
		rep             = json.RawMessage(`{"key":"value"}`)
	)

	mockHealthChecker.EXPECT().HealthCheck(ctx, healthCheckName).Return(rep, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "HealthCheck", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.HealthCheck(ctx, healthCheckName)

	// Without correlation ID.
	mockHealthChecker.EXPECT().HealthCheck(context.Background(), healthCheckName).Return(rep, nil).Times(1)
	var f = func() {
		m.HealthCheck(context.Background(), healthCheckName)
	}
	assert.Panics(t, f)
}
