package dexcomshare

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	// ErrAuthenticationFailed is returned when an account cannot be authenticated.
	ErrAuthenticationFailed = errors.New("dexcomshare: authentication failed")
	// ErrLoginFailed is returned when a session cannot be established.
	ErrLoginFailed = errors.New("dexcomshare: login failed")
	// ErrReadGlucoseFailed is returned when glucose readings cannot be retrieved.
	ErrReadGlucoseFailed = errors.New("dexcomshare: read glucose failed")
)

// APIError describes an unsuccessful response from the Share API. The API
// reports failures as a JSON body containing a Code and a Message.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	switch {
	case e.Code != "" && e.Message != "":
		return fmt.Sprintf("%s: %s (http %d)", e.Code, e.Message, e.StatusCode)
	case e.Code != "":
		return fmt.Sprintf("%s (http %d)", e.Code, e.StatusCode)
	default:
		return fmt.Sprintf("unexpected response (http %d)", e.StatusCode)
	}
}

func newAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}

	var payload struct {
		Code    string
		Message string
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		apiErr.Code = payload.Code
		apiErr.Message = payload.Message
	}

	return apiErr
}
