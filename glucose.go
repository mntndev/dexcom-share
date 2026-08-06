package dexcomshare

import (
	"context"
	"fmt"
	"math"
)

const (
	maxMinutes  = 1440
	maxMaxCount = 288

	// mmolLConversionFactor converts mg/dL to mmol/L.
	mmolLConversionFactor = 0.0555
)

// GlucoseEntry is a single glucose reading.
type GlucoseEntry struct {
	Value int   `json:"Value"` // Value is the glucose reading in mg/dL.
	Trend Trend `json:"Trend"`
	DT    Time  `json:"DT"` // DT is the display time, in the device's own UTC offset.
	WT    Time  `json:"WT"` // WT is the wall time reported by the transmitter.
	ST    Time  `json:"ST"` // ST is the system time reported by the transmitter.
}

// MmolL returns the reading in mmol/L, rounded to one decimal place the way the
// Dexcom apps display it.
func (e GlucoseEntry) MmolL() float64 {
	return math.Round(float64(e.Value)*mmolLConversionFactor*10) / 10
}

type readGlucoseRequest struct {
	SessionID string `json:"sessionId"`
	Minutes   int    `json:"minutes"`
	MaxCount  int    `json:"maxCount"`
}

// ReadGlucose returns up to maxCount glucose entries recorded within the last
// minutes minutes, most recent first. minutes must be between 1 and 1440, and
// maxCount between 1 and 288.
func (c *Client) ReadGlucose(ctx context.Context, minutes, maxCount int) ([]GlucoseEntry, error) {
	if minutes < 1 || minutes > maxMinutes {
		return nil, ErrInvalidMinutes
	}

	if maxCount < 1 || maxCount > maxMaxCount {
		return nil, ErrInvalidMaxCount
	}

	request := readGlucoseRequest{
		SessionID: c.sessionID,
		Minutes:   minutes,
		MaxCount:  maxCount,
	}

	var entries []GlucoseEntry
	if err := c.postJSON(ctx, readGlucoseEndpoint, request, &entries); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadGlucoseFailed, err)
	}

	return entries, nil
}
