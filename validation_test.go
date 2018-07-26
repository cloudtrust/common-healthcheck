package common_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	. "github.com/cloudtrust/common-healthcheck"
	"github.com/cloudtrust/common-healthcheck/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestValidationMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockHealthChecker = mock.NewHealthChecker(mockCtrl)

	var (
		validValues = map[string]struct{}{
			"valid1": struct{}{},
			"valid2": struct{}{},
			"valid3": struct{}{},
		}
		m   = MakeValidationMiddleware(validValues)(mockHealthChecker)
		rep = json.RawMessage(`{"key":"value"}`)
	)

	var tsts = []struct {
		name    string
		isValid bool
	}{
		{"valid1", true},
		{"valid2", true},
		{"valid3", true},
		{"invalid1", false},
		{"invalid2", false},
		{"invalid3", false},
	}

	for _, tst := range tsts {
		if tst.isValid {
			mockHealthChecker.EXPECT().HealthCheck(context.Background(), gomock.Any()).Return(rep, nil).Times(1)
		}

		var report, err = m.HealthCheck(context.Background(), tst.name)

		if tst.isValid {
			assert.Nil(t, err)
			assert.NotNil(t, report)
		} else {
			assert.IsType(t, &ErrInvalidHCName{}, err)
			assert.Nil(t, report)
		}
	}
}
