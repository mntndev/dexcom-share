package dexcomshare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// BaseURLUS serves accounts registered inside the United States. It is the default.
	BaseURLUS = "https://share2.dexcom.com/ShareWebServices/Services"
	// BaseURLOutsideUS serves accounts registered outside the United States.
	BaseURLOutsideUS = "https://shareous1.dexcom.com/ShareWebServices/Services"

	authenticateEndpoint = "General/AuthenticatePublisherAccount"
	loginIDEndpoint      = "General/LoginPublisherAccountById"
	readGlucoseEndpoint  = "Publisher/ReadPublisherLatestGlucoseValues"

	// applicationID identifies the Dexcom Share mobile app to the API.
	applicationID = "d89443d2-327c-4a6f-89e5-496bbb0317db"
)

// Client is a client for the Dexcom Share API.
type Client struct {
	username   string
	password   string
	baseURL    string
	accountID  string // accountID is the account ID returned by the authentication endpoint.
	sessionID  string // sessionID is the session ID returned by the login endpoint.
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient configures a Client with a custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL configures a Client to talk to a different Share host, such as
// BaseURLOutsideUS.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// NewClient creates a new Dexcom Share client and establishes a session.
func NewClient(ctx context.Context, username, password string, options ...Option) (*Client, error) {
	client := &Client{
		username:   username,
		password:   password,
		baseURL:    BaseURLUS,
		httpClient: &http.Client{},
	}

	for _, option := range options {
		option(client)
	}

	if err := client.authenticate(ctx); err != nil {
		return nil, err
	}

	if err := client.login(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

type authenticateRequest struct {
	AccountName string `json:"accountName"`
	Password    string `json:"password"`
	Application string `json:"applicationId"`
}

// authenticate exchanges the account name and password for an account ID.
func (c *Client) authenticate(ctx context.Context) error {
	request := authenticateRequest{
		AccountName: c.username,
		Password:    c.password,
		Application: applicationID,
	}

	if err := c.postJSON(ctx, authenticateEndpoint, request, &c.accountID); err != nil {
		return fmt.Errorf("%w: %w", ErrAuthenticationFailed, err)
	}

	return nil
}

type loginRequest struct {
	AccountID   string `json:"accountId"`
	Password    string `json:"password"`
	Application string `json:"applicationId"`
}

// login exchanges the account ID for a session ID. It must be called after authenticate.
func (c *Client) login(ctx context.Context) error {
	request := loginRequest{
		AccountID:   c.accountID,
		Password:    c.password,
		Application: applicationID,
	}

	if err := c.postJSON(ctx, loginIDEndpoint, request, &c.sessionID); err != nil {
		return fmt.Errorf("%w: %w", ErrLoginFailed, err)
	}

	return nil
}

// postJSON posts in as JSON to endpoint and decodes the response into out.
func (c *Client) postJSON(ctx context.Context, endpoint string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/"+endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close() //nolint:errcheck // nothing actionable on a response body close

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return newAPIError(response.StatusCode, data)
	}

	return json.Unmarshal(data, out)
}
