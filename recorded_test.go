package dexcomshare

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func beforeSaveHook(i *cassette.Interaction) error {
	// Redact the credentials the client sends in the request body.
	tmp := map[string]any{}
	if err := json.Unmarshal([]byte(i.Request.Body), &tmp); err == nil {
		for _, key := range []string{"accountId", "accountName", "password", "sessionId"} {
			if _, ok := tmp[key]; ok {
				tmp[key] = "REDACTED"
			}
		}

		b, err := json.Marshal(tmp)
		if err != nil {
			return err
		}

		i.Request.Body = string(b)
	}

	// Redact the account ID and session ID the API returns. Match on the endpoint
	// suffix so a recording made against either regional host is covered; keying
	// off BaseURLUS alone let the session ID leak once already.
	switch {
	case strings.HasSuffix(i.Request.URL, "/"+authenticateEndpoint):
		i.Response.Body = `"accountID"`
	case strings.HasSuffix(i.Request.URL, "/"+loginIDEndpoint):
		i.Response.Body = `"sessionID"`
	}

	return nil
}

// matchByMethodAndURL replaces the default matcher, which also compares request
// bodies. Recorded bodies have their credentials redacted and never match.
func matchByMethodAndURL(r *http.Request, i cassette.Request) bool {
	return r.Method == i.Method && r.URL.String() == i.URL
}

// TestBeforeSaveHookRedacts exercises the redaction hook directly. TestRecordedSession
// replays in recorded mode and so never runs the hook, which is how a live session ID
// once reached the committed cassette; this test keeps the redaction honest in CI.
func TestBeforeSaveHookRedacts(t *testing.T) {
	t.Run("authenticate request and response", func(t *testing.T) {
		i := &cassette.Interaction{}
		i.Request.URL = BaseURLUS + "/" + authenticateEndpoint
		i.Request.Body = `{"accountName":"me@example.com","applicationId":"app","password":"hunter2"}`
		i.Response.Body = `"real-account-id"`

		require.NoError(t, beforeSaveHook(i))

		got := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(i.Request.Body), &got))
		assert.Equal(t, "REDACTED", got["accountName"])
		assert.Equal(t, "REDACTED", got["password"])
		assert.Equal(t, "app", got["applicationId"], "non-secret fields must survive")
		assert.Equal(t, `"accountID"`, i.Response.Body)
	})

	t.Run("login request and response", func(t *testing.T) {
		i := &cassette.Interaction{}
		i.Request.URL = BaseURLUS + "/" + loginIDEndpoint
		i.Request.Body = `{"accountId":"real-account-id","applicationId":"app","password":"hunter2"}`
		i.Response.Body = `"3c348745-d2b3-4e3e-a8fe-306c0c18eae6"`

		require.NoError(t, beforeSaveHook(i))

		got := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(i.Request.Body), &got))
		assert.Equal(t, "REDACTED", got["accountId"])
		assert.Equal(t, "REDACTED", got["password"])
		assert.Equal(t, `"sessionID"`, i.Response.Body, "the session ID must never reach the cassette")
	})

	t.Run("login response is redacted for the outside-US host too", func(t *testing.T) {
		i := &cassette.Interaction{}
		i.Request.URL = BaseURLOutsideUS + "/" + loginIDEndpoint
		i.Response.Body = `"3c348745-d2b3-4e3e-a8fe-306c0c18eae6"`

		require.NoError(t, beforeSaveHook(i))
		assert.Equal(t, `"sessionID"`, i.Response.Body)
	})

	t.Run("read glucose redacts the session ID and leaves data intact", func(t *testing.T) {
		i := &cassette.Interaction{}
		i.Request.URL = BaseURLUS + "/" + readGlucoseEndpoint
		i.Request.Body = `{"sessionId":"3c348745-d2b3-4e3e-a8fe-306c0c18eae6","minutes":1440,"maxCount":1}`
		i.Response.Body = `[{"Value":160,"Trend":"FortyFiveUp"}]`

		require.NoError(t, beforeSaveHook(i))

		got := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(i.Request.Body), &got))
		assert.Equal(t, "REDACTED", got["sessionId"])
		assert.Equal(t, `[{"Value":160,"Trend":"FortyFiveUp"}]`, i.Response.Body)
	})
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
	assert.Equal(t, "sessionID", c.sessionID)

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
