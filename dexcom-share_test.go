package dexcomshare

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func beforeSaveHook(i *cassette.Interaction) error {
	// Remove account ID, name, and password from the request body
	tmp := map[string]interface{}{}

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

func Test_Client(t *testing.T) {
	r, err := recorder.New(
		"testdata/Test_NewClient",
		recorder.WithHook(beforeSaveHook, recorder.BeforeSaveHook),
		recorder.WithMatcher(matchByMethodAndURL),
	)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, r.Stop()) }()

	client := r.GetDefaultClient()

	c, err := NewClient("username", "password", WithHTTPClient(client))
	assert.NoError(t, err)
	assert.NotNil(t, c)

	entries, err := c.ReadGlucose(1440, 1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entries, err = c.ReadGlucose(1440, 100)
	assert.NoError(t, err)
	assert.Len(t, entries, 100)
}
