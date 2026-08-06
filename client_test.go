package dexcomshare

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client, server := newStubClient(t, sessionResponses())

	assert.Equal(t, "account-id", client.accountID)
	assert.Equal(t, "session-id", client.sessionID)

	authenticateRequests := server.requestsFor(authenticateEndpoint)
	require.Len(t, authenticateRequests, 1)
	assert.JSONEq(t,
		`{"accountName":"username","password":"password","applicationId":"`+applicationID+`"}`,
		authenticateRequests[0],
	)

	// The login request must carry the account ID handed back by authenticate.
	loginRequests := server.requestsFor(loginIDEndpoint)
	require.Len(t, loginRequests, 1)
	assert.JSONEq(t,
		`{"accountId":"account-id","password":"password","applicationId":"`+applicationID+`"}`,
		loginRequests[0],
	)
}

func TestNewClientErrors(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]stubResponse
		sentinel  error
		apiError  *APIError
	}{
		{
			name: "authenticate rejected",
			responses: map[string]stubResponse{
				authenticateEndpoint: {
					status: http.StatusUnauthorized,
					body:   `{"Code":"AccountPasswordInvalid","Message":"Password not valid"}`,
				},
			},
			sentinel: ErrAuthenticationFailed,
			apiError: &APIError{
				StatusCode: http.StatusUnauthorized,
				Code:       "AccountPasswordInvalid",
				Message:    "Password not valid",
			},
		},
		{
			name: "authenticate returns unparsable error body",
			responses: map[string]stubResponse{
				authenticateEndpoint: {status: http.StatusInternalServerError, body: `<html>oops</html>`},
			},
			sentinel: ErrAuthenticationFailed,
			apiError: &APIError{StatusCode: http.StatusInternalServerError},
		},
		{
			name: "authenticate returns malformed json",
			responses: map[string]stubResponse{
				authenticateEndpoint: okResponse(`{"not":"a string"}`),
			},
			sentinel: ErrAuthenticationFailed,
		},
		{
			name: "login rejected",
			responses: map[string]stubResponse{
				authenticateEndpoint: okResponse(`"account-id"`),
				loginIDEndpoint: {
					status: http.StatusUnauthorized,
					body:   `{"Code":"SessionNotValid","Message":"Session ID not valid"}`,
				},
			},
			sentinel: ErrLoginFailed,
			apiError: &APIError{
				StatusCode: http.StatusUnauthorized,
				Code:       "SessionNotValid",
				Message:    "Session ID not valid",
			},
		},
		{
			name: "login returns malformed json",
			responses: map[string]stubResponse{
				authenticateEndpoint: okResponse(`"account-id"`),
				loginIDEndpoint:      okResponse(`nonsense`),
			},
			sentinel: ErrLoginFailed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newStubServer(t, test.responses)

			client, err := NewClient(t.Context(), "username", "password", WithBaseURL(server.URL))
			assert.Nil(t, client)
			require.ErrorIs(t, err, test.sentinel)

			var apiErr *APIError
			if test.apiError == nil {
				assert.NotErrorAs(t, err, &apiErr)

				return
			}

			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, test.apiError, apiErr)
			assert.Contains(t, err.Error(), apiErr.Error())
		})
	}
}

func TestNewClientUnreachableServer(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	url := server.URL
	server.Close()

	_, err := NewClient(t.Context(), "username", "password", WithBaseURL(url))
	assert.ErrorIs(t, err, ErrAuthenticationFailed)
}

func TestNewClientCancelledContext(t *testing.T) {
	server := newStubServer(t, sessionResponses())

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := NewClient(ctx, "username", "password", WithBaseURL(server.URL))
	assert.ErrorIs(t, err, context.Canceled)
	assert.ErrorIs(t, err, ErrAuthenticationFailed)
	assert.Empty(t, server.requestsFor(authenticateEndpoint))
}

func TestOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		client, _ := newStubClient(t, sessionResponses())

		assert.NotNil(t, client.httpClient)
		assert.Equal(t, "username", client.username)
		assert.Equal(t, "password", client.password)
	})

	t.Run("base url defaults to the US host", func(t *testing.T) {
		client := &Client{baseURL: BaseURLUS}
		assert.Equal(t, "https://share2.dexcom.com/ShareWebServices/Services", client.baseURL)

		WithBaseURL(BaseURLOutsideUS)(client)
		assert.Equal(t, "https://shareous1.dexcom.com/ShareWebServices/Services", client.baseURL)
	})

	t.Run("with http client", func(t *testing.T) {
		transport := &countingTransport{}
		server := newStubServer(t, sessionResponses())

		_, err := NewClient(t.Context(), "username", "password",
			WithBaseURL(server.URL),
			WithHTTPClient(&http.Client{Transport: transport}),
		)
		require.NoError(t, err)
		assert.Equal(t, 2, transport.calls, "authenticate and login should go through the supplied client")
	})
}

type countingTransport struct {
	calls int
}

func (t *countingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.calls++

	return http.DefaultTransport.RoundTrip(request)
}
