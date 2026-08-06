package dexcomshare

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResponse struct {
	status int
	body   string
}

func okResponse(body string) stubResponse {
	return stubResponse{status: http.StatusOK, body: body}
}

// stubServer stands in for the Share API and records the request bodies it receives.
type stubServer struct {
	*httptest.Server

	mu       sync.Mutex
	requests map[string][]string
}

func newStubServer(t *testing.T, responses map[string]stubResponse) *stubServer {
	t.Helper()

	server := &stubServer{requests: map[string][]string{}}

	mux := http.NewServeMux()
	for endpoint, response := range responses {
		mux.HandleFunc("/"+endpoint, func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)

			server.mu.Lock()
			server.requests[endpoint] = append(server.requests[endpoint], string(body))
			server.mu.Unlock()

			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(response.status)
			_, err = io.WriteString(w, response.body)
			assert.NoError(t, err)
		})
	}

	server.Server = httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

func (s *stubServer) requestsFor(endpoint string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.requests[endpoint]
}

// sessionResponses are the authenticate and login responses of a healthy session.
func sessionResponses() map[string]stubResponse {
	return map[string]stubResponse{
		authenticateEndpoint: okResponse(`"account-id"`),
		loginIDEndpoint:      okResponse(`"session-id"`),
	}
}

// newStubClient returns a client with an established session against a stub server.
func newStubClient(t *testing.T, responses map[string]stubResponse) (*Client, *stubServer) {
	t.Helper()

	server := newStubServer(t, responses)

	client, err := NewClient(t.Context(), "username", "password", WithBaseURL(server.URL))
	require.NoError(t, err)

	return client, server
}
