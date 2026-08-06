package dexcomshare

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		message  string
	}{
		{
			name:     "code and message",
			apiError: &APIError{StatusCode: 401, Code: "SessionNotValid", Message: "Session ID not valid"},
			message:  "SessionNotValid: Session ID not valid (http 401)",
		},
		{
			name:     "code only",
			apiError: &APIError{StatusCode: 401, Code: "SessionNotValid"},
			message:  "SessionNotValid (http 401)",
		},
		{
			name:     "neither",
			apiError: &APIError{StatusCode: 502},
			message:  "unexpected response (http 502)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.message, test.apiError.Error())
		})
	}
}

func TestNewAPIError(t *testing.T) {
	apiErr := newAPIError(http.StatusUnauthorized, []byte(`{"Code":"AccountNotFound","Message":"Account not found"}`))
	assert.Equal(t, &APIError{
		StatusCode: http.StatusUnauthorized,
		Code:       "AccountNotFound",
		Message:    "Account not found",
	}, apiErr)

	// A body that is not the documented error shape still yields the status code.
	apiErr = newAPIError(http.StatusBadGateway, []byte(`<html>bad gateway</html>`))
	assert.Equal(t, &APIError{StatusCode: http.StatusBadGateway}, apiErr)
}

func TestInvalidBaseURL(t *testing.T) {
	_, err := NewClient(t.Context(), "username", "password", WithBaseURL("://not a url"))
	assert.ErrorIs(t, err, ErrAuthenticationFailed)
}
