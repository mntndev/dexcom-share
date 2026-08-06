package dexcomshare

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func beforeSaveHook(i *cassette.Interaction) error {
	// Remove account ID, name, and password from the request body
	tmp := map[string]any{}

	err := json.Unmarshal([]byte(i.Request.Body), &tmp)
	if err == nil {
		if _, ok := tmp["accountId"]; ok {
			tmp["accountId"] = "REDACTED"
		}

		if _, ok := tmp["accountName"]; ok {
			tmp["accountName"] = "REDACTED"
		}

		if _, ok := tmp["password"]; ok {
			tmp["password"] = "REDACTED"
		}

		if _, ok := tmp["sessionId"]; ok {
			tmp["sessionId"] = "REDACTED"
		}

		b, err := json.Marshal(tmp)
		if err != nil {
			return err
		}

		i.Request.Body = string(b)
	}

	if i.Request.URL == BaseURLUS+"/"+authenticateEndpoint {
		i.Response.Body = `"accountID"`
	}

	if i.Request.URL == BaseURLUS+"/"+loginIDEndpoint {
		i.Response.Body = `"sessionID"`
	}

	return nil
}

// matchByMethodAndURL replaces the default matcher, which also compares request
// bodies. Recorded bodies have their credentials redacted and never match.
func matchByMethodAndURL(r *http.Request, i cassette.Request) bool {
	return r.Method == i.Method && r.URL.String() == i.URL
}

// TestRecordedSession replays a session recorded against the real Share API, so
// the package keeps decoding responses in the shape Dexcom actually sends. The
// cassette is replayed in order: the maxCount=1 read must precede the
// maxCount=100 read.
func TestRecordedSession(t *testing.T) {
	r, err := recorder.New(
		"testdata/TestRecordedSession",
		recorder.WithHook(beforeSaveHook, recorder.BeforeSaveHook),
		recorder.WithMatcher(matchByMethodAndURL),
		recorder.WithSkipRequestLatency(true),
	)
	require.NoError(t, err)
	defer func() { assert.NoError(t, r.Stop()) }()

	// Guard against the cassette going missing and this test quietly turning
	// into a live call against the real API.
	require.False(t, r.IsNewCassette())
	require.False(t, r.IsRecording())

	c, err := NewClient(t.Context(), "username", "password", WithHTTPClient(r.GetDefaultClient()))
	require.NoError(t, err)
	assert.Equal(t, "3c348745-d2b3-4e3e-a8fe-306c0c18eae6", c.sessionID)

	entries, err := c.ReadGlucose(t.Context(), 1440, 1)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, 160, entries[0].Value)
	assert.Equal(t, TrendFortyFiveUp, entries[0].Trend)
	assert.Equal(t, "2023-04-03T23:57:08-07:00", entries[0].DT.Format(time.RFC3339))
	assert.Equal(t, "2023-04-04T06:57:08Z", entries[0].WT.Format(time.RFC3339))

	entries, err = c.ReadGlucose(t.Context(), 1440, 100)
	require.NoError(t, err)
	require.Len(t, entries, 100)

	// Every reading in the recording decodes into a trend this package knows about.
	for _, entry := range entries {
		assert.True(t, entry.Trend.Known(), "unknown trend %q", entry.Trend)
		assert.False(t, entry.WT.IsZero())
	}
}
