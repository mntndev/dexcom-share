package dexcomshare

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func glucoseResponses(t *testing.T, response stubResponse) map[string]stubResponse {
	t.Helper()

	responses := sessionResponses()
	responses[readGlucoseEndpoint] = response

	return responses
}

func TestReadGlucose(t *testing.T) {
	fixture, err := os.ReadFile("testdata/glucose.json")
	require.NoError(t, err)

	client, server := newStubClient(t, glucoseResponses(t, okResponse(string(fixture))))

	entries, err := client.ReadGlucose(t.Context(), 1440, 3)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	// The session ID from login must be attached to the request.
	requests := server.requestsFor(readGlucoseEndpoint)
	require.Len(t, requests, 1)
	assert.JSONEq(t, `{"sessionId":"session-id","minutes":1440,"maxCount":3}`, requests[0])

	first := entries[0]
	assert.Equal(t, 160, first.Value)
	assert.Equal(t, TrendFortyFiveUp, first.Trend)
	assert.InDelta(t, 8.9, first.MmolL(), 0.001)

	// DT carries the device's UTC offset; WT and ST do not.
	assert.Equal(t, time.UnixMilli(1680591428000).UTC(), first.WT.UTC())
	assert.Equal(t, time.UnixMilli(1680591428000).UTC(), first.ST.UTC())
	assert.Equal(t, "2023-04-03T23:57:08-07:00", first.DT.Format(time.RFC3339))
	assert.Equal(t, "2023-04-04T06:57:08Z", first.WT.Format(time.RFC3339))

	assert.Equal(t, TrendFlat, entries[2].Trend)
}

func TestReadGlucoseValidation(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		maxCount int
		sentinel error
	}{
		{name: "minutes too low", minutes: 0, maxCount: 1, sentinel: ErrInvalidMinutes},
		{name: "minutes too high", minutes: 1441, maxCount: 1, sentinel: ErrInvalidMinutes},
		{name: "minutes negative", minutes: -10, maxCount: 1, sentinel: ErrInvalidMinutes},
		{name: "max count too low", minutes: 1440, maxCount: 0, sentinel: ErrInvalidMaxCount},
		{name: "max count too high", minutes: 1440, maxCount: 289, sentinel: ErrInvalidMaxCount},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, server := newStubClient(t, glucoseResponses(t, okResponse(`[]`)))

			entries, err := client.ReadGlucose(t.Context(), test.minutes, test.maxCount)
			assert.Nil(t, entries)
			assert.ErrorIs(t, err, test.sentinel)
			assert.Empty(t, server.requestsFor(readGlucoseEndpoint), "invalid arguments must not reach the API")
		})
	}

	t.Run("bounds are accepted", func(t *testing.T) {
		client, _ := newStubClient(t, glucoseResponses(t, okResponse(`[]`)))

		entries, err := client.ReadGlucose(t.Context(), 1, 1)
		require.NoError(t, err)
		assert.Empty(t, entries)

		entries, err = client.ReadGlucose(t.Context(), 1440, 288)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

func TestReadGlucoseErrors(t *testing.T) {
	tests := []struct {
		name     string
		response stubResponse
		apiError *APIError
	}{
		{
			name: "session expired",
			response: stubResponse{
				status: http.StatusUnauthorized,
				body:   `{"Code":"SessionIdNotFound","Message":"Session ID not found"}`,
			},
			apiError: &APIError{
				StatusCode: http.StatusUnauthorized,
				Code:       "SessionIdNotFound",
				Message:    "Session ID not found",
			},
		},
		{
			name:     "server error",
			response: stubResponse{status: http.StatusInternalServerError, body: ``},
			apiError: &APIError{StatusCode: http.StatusInternalServerError},
		},
		{
			name:     "malformed json",
			response: okResponse(`[{"Value":`),
		},
		{
			name:     "unparsable timestamp",
			response: okResponse(`[{"Value":100,"Trend":"Flat","WT":"yesterday"}]`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := newStubClient(t, glucoseResponses(t, test.response))

			entries, err := client.ReadGlucose(t.Context(), 1440, 1)
			assert.Nil(t, entries)
			require.ErrorIs(t, err, ErrReadGlucoseFailed)

			var apiErr *APIError
			if test.apiError == nil {
				assert.NotErrorAs(t, err, &apiErr)

				return
			}

			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, test.apiError, apiErr)
		})
	}
}

func TestMmolL(t *testing.T) {
	tests := []struct {
		value int
		mmolL float64
	}{
		{value: 0, mmolL: 0},
		{value: 40, mmolL: 2.2},
		{value: 100, mmolL: 5.6},
		{value: 144, mmolL: 8.0},
		{value: 160, mmolL: 8.9},
		{value: 400, mmolL: 22.2},
	}

	for _, test := range tests {
		entry := GlucoseEntry{Value: test.value}
		assert.InDelta(t, test.mmolL, entry.MmolL(), 0.001)
	}
}
