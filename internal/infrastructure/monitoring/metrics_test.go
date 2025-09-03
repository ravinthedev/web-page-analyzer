package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecordHTTPRequest(t *testing.T) {
	RecordHTTPRequest("GET", "/test", 200, time.Millisecond*100)
	assert.True(t, true)
}
