package dexcomshare

// Trend describes the direction a glucose reading is heading. The Share API
// reports it as a string; values this package does not recognize are preserved
// as-is rather than discarded.
type Trend string

// Trends reported by the Share API.
const (
	TrendNone           Trend = "None"
	TrendDoubleUp       Trend = "DoubleUp"
	TrendSingleUp       Trend = "SingleUp"
	TrendFortyFiveUp    Trend = "FortyFiveUp"
	TrendFlat           Trend = "Flat"
	TrendFortyFiveDown  Trend = "FortyFiveDown"
	TrendSingleDown     Trend = "SingleDown"
	TrendDoubleDown     Trend = "DoubleDown"
	TrendNotComputable  Trend = "NotComputable"
	TrendRateOutOfRange Trend = "RateOutOfRange"
)

var trends = map[Trend]struct {
	arrow       string
	description string
}{
	TrendNone:           {"", ""},
	TrendDoubleUp:       {"↑↑", "rising quickly"},
	TrendSingleUp:       {"↑", "rising"},
	TrendFortyFiveUp:    {"↗", "rising slightly"},
	TrendFlat:           {"→", "steady"},
	TrendFortyFiveDown:  {"↘", "falling slightly"},
	TrendSingleDown:     {"↓", "falling"},
	TrendDoubleDown:     {"↓↓", "falling quickly"},
	TrendNotComputable:  {"?", "unable to determine trend"},
	TrendRateOutOfRange: {"-", "trend unavailable"},
}

// Arrow returns the arrow the Dexcom apps display for the trend. It is empty
// for TrendNone and for unrecognized trends.
func (t Trend) Arrow() string {
	return trends[t].arrow
}

// Description returns a human readable description of the trend, or "unknown"
// for an unrecognized trend.
func (t Trend) Description() string {
	trend, ok := trends[t]
	if !ok {
		return "unknown"
	}

	return trend.description
}

// Known reports whether the trend is one this package recognizes.
func (t Trend) Known() bool {
	_, ok := trends[t]

	return ok
}
