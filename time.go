package dexcomshare

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Time is a Share API timestamp. The API reports timestamps in an ASP.NET
// flavored format holding Unix milliseconds and an optional UTC offset, such as
// Date(1680591428000) or Date(1680591428000-0700).
//
// It embeds time.Time, so the usual time methods are available directly.
type Time struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Time) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, err := parseTime(raw)
	if err != nil {
		return err
	}

	*t = parsed

	return nil
}

// MarshalJSON implements json.Marshaler, emitting the format the Share API uses.
func (t Time) MarshalJSON() ([]byte, error) {
	millis := t.UnixMilli()

	_, offset := t.Zone()
	if offset == 0 {
		return json.Marshal(fmt.Sprintf("Date(%d)", millis))
	}

	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}

	return json.Marshal(fmt.Sprintf("Date(%d%s%02d%02d)", millis, sign, offset/3600, offset%3600/60))
}

func parseTime(raw string) (Time, error) {
	body, ok := strings.CutPrefix(raw, "Date(")
	if !ok {
		return Time{}, fmt.Errorf("dexcomshare: unrecognized timestamp %q", raw)
	}

	body, ok = strings.CutSuffix(body, ")")
	if !ok {
		return Time{}, fmt.Errorf("dexcomshare: unrecognized timestamp %q", raw)
	}

	// The offset, when present, is appended directly to the milliseconds.
	millis, location := body, time.UTC
	if index := strings.IndexAny(body, "+-"); index > 0 {
		zone, err := parseZone(body[index:])
		if err != nil {
			return Time{}, fmt.Errorf("dexcomshare: unrecognized timestamp %q: %w", raw, err)
		}

		millis, location = body[:index], zone
	}

	parsed, err := strconv.ParseInt(millis, 10, 64)
	if err != nil {
		return Time{}, fmt.Errorf("dexcomshare: unrecognized timestamp %q: %w", raw, err)
	}

	return Time{time.UnixMilli(parsed).In(location)}, nil
}

func parseZone(offset string) (*time.Location, error) {
	zone, err := time.Parse("-0700", offset)
	if err != nil {
		return nil, fmt.Errorf("unrecognized UTC offset %q", offset)
	}

	_, seconds := zone.Zone()
	if seconds == 0 {
		return time.UTC, nil
	}

	return time.FixedZone(offset, seconds), nil
}
