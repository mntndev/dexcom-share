package dexcomshare

import "fmt"

// GlucoseEntry is a single glucose reading.
type GlucoseEntry struct {
	Value int    `json:"Value"` // Value is the glucose reading in mg/dL.
	Trend string `json:"Trend"`
	DT    string `json:"DT"`
	WT    string `json:"WT"`
	ST    string `json:"ST"`
}

type readGlucoseRequest struct {
	SessionID string `json:"sessionId"`
	Minutes   int    `json:"minutes"`
	MaxCount  int    `json:"maxCount"`
}

// ReadGlucose returns up to maxCount glucose entries recorded within the last
// minutes minutes, most recent first.
func (c *Client) ReadGlucose(minutes, maxCount int) ([]GlucoseEntry, error) {
	request := readGlucoseRequest{
		SessionID: c.sessionID,
		Minutes:   minutes,
		MaxCount:  maxCount,
	}

	var entries []GlucoseEntry
	if err := c.postJSON(readGlucoseEndpoint, request, &entries); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadGlucoseFailed, err)
	}

	return entries, nil
}
