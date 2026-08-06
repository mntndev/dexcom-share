package dexcomshare

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrend(t *testing.T) {
	tests := []struct {
		trend       Trend
		arrow       string
		description string
		known       bool
	}{
		{trend: TrendNone, arrow: "", description: "", known: true},
		{trend: TrendDoubleUp, arrow: "↑↑", description: "rising quickly", known: true},
		{trend: TrendSingleUp, arrow: "↑", description: "rising", known: true},
		{trend: TrendFortyFiveUp, arrow: "↗", description: "rising slightly", known: true},
		{trend: TrendFlat, arrow: "→", description: "steady", known: true},
		{trend: TrendFortyFiveDown, arrow: "↘", description: "falling slightly", known: true},
		{trend: TrendSingleDown, arrow: "↓", description: "falling", known: true},
		{trend: TrendDoubleDown, arrow: "↓↓", description: "falling quickly", known: true},
		{trend: TrendNotComputable, arrow: "?", description: "unable to determine trend", known: true},
		{trend: TrendRateOutOfRange, arrow: "-", description: "trend unavailable", known: true},
		{trend: Trend("SomethingNew"), arrow: "", description: "unknown", known: false},
		{trend: Trend(""), arrow: "", description: "unknown", known: false},
	}

	for _, test := range tests {
		t.Run(string(test.trend), func(t *testing.T) {
			assert.Equal(t, test.arrow, test.trend.Arrow())
			assert.Equal(t, test.description, test.trend.Description())
			assert.Equal(t, test.known, test.trend.Known())
		})
	}
}

// Unrecognized trends must survive a decode/encode round trip rather than being
// flattened to a known value.
func TestTrendRoundTrip(t *testing.T) {
	for _, raw := range []string{"Flat", "SomethingNew"} {
		var trend Trend
		require.NoError(t, json.Unmarshal([]byte(`"`+raw+`"`), &trend))
		assert.Equal(t, Trend(raw), trend)

		encoded, err := json.Marshal(trend)
		require.NoError(t, err)
		assert.JSONEq(t, `"`+raw+`"`, string(encoded))
	}
}
