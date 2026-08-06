package dexcomshare

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		instant time.Time
		offset  int
	}{
		{
			name:    "without offset",
			raw:     `"Date(1680591428000)"`,
			instant: time.UnixMilli(1680591428000),
			offset:  0,
		},
		{
			name:    "negative offset",
			raw:     `"Date(1680591428000-0700)"`,
			instant: time.UnixMilli(1680591428000),
			offset:  -7 * 3600,
		},
		{
			name:    "positive offset",
			raw:     `"Date(1680591428000+0530)"`,
			instant: time.UnixMilli(1680591428000),
			offset:  5*3600 + 30*60,
		},
		{
			name:    "zero offset is UTC",
			raw:     `"Date(1680591428000+0000)"`,
			instant: time.UnixMilli(1680591428000),
			offset:  0,
		},
		{
			name:    "epoch",
			raw:     `"Date(0)"`,
			instant: time.UnixMilli(0),
			offset:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var parsed Time
			require.NoError(t, json.Unmarshal([]byte(test.raw), &parsed))

			assert.True(t, parsed.Equal(test.instant), "expected %s, got %s", test.instant, parsed)

			_, offset := parsed.Zone()
			assert.Equal(t, test.offset, offset)
		})
	}
}

func TestTimeUnmarshalJSONErrors(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "not a string", raw: `1680591428000`},
		{name: "not a date wrapper", raw: `"1680591428000"`},
		{name: "missing closing paren", raw: `"Date(1680591428000"`},
		{name: "empty wrapper", raw: `"Date()"`},
		{name: "milliseconds not a number", raw: `"Date(yesterday)"`},
		{name: "truncated offset", raw: `"Date(1680591428000-07)"`},
		{name: "nonsense offset", raw: `"Date(1680591428000-ab00)"`},
		{name: "out of range offset", raw: `"Date(1680591428000+9900)"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var parsed Time
			assert.Error(t, json.Unmarshal([]byte(test.raw), &parsed))
		})
	}
}

func TestTimeMarshalJSON(t *testing.T) {
	for _, raw := range []string{
		`"Date(1680591428000)"`,
		`"Date(1680591428000-0700)"`,
		`"Date(1680591428000+0530)"`,
	} {
		t.Run(raw, func(t *testing.T) {
			var parsed Time
			require.NoError(t, json.Unmarshal([]byte(raw), &parsed))

			encoded, err := json.Marshal(parsed)
			require.NoError(t, err)
			assert.JSONEq(t, raw, string(encoded))
		})
	}
}

func TestTimeEmbedsStdlibTime(t *testing.T) {
	var parsed Time
	require.NoError(t, json.Unmarshal([]byte(`"Date(1680591428000-0700)"`), &parsed))

	assert.Equal(t, 2023, parsed.Year())
	assert.Equal(t, time.April, parsed.Month())
	assert.True(t, parsed.After(time.UnixMilli(1680591427000)))
}
