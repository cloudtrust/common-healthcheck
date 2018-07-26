package common

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestStr(t *testing.T) {
	var vals = []struct {
		in       error
		expected string
	}{
		{nil, ""},
		{fmt.Errorf("fail"), "fail"},
		{fmt.Errorf("error"), "error"},
		{fmt.Errorf("unexpected failure"), "unexpected failure"},
	}

	for _, v := range vals {
		assert.Equal(t, v.expected, str(v.in))
	}
}
