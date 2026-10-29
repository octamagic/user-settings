// Package userdirectory is an HTTP client for the upstream user directory,
// the system of record for a user's date of birth and country.
package userdirectory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// User is the subset of a directory record this service needs.
type User struct {
	URN         string  `json:"urn"`
	DateOfBirth *string `json:"date_of_birth"`
	CountryCode *string `json:"country_code"`
}

// Client fetches users from the directory over HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New returns a Client for the directory at baseURL.
func New(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
	}
}

// FetchUsers returns the directory records for the given URNs. URNs the
// directory does not know are simply absent from the result.
func (c *Client) FetchUsers(ctx context.Context, userUrns []string) ([]User, error) {
	query := url.Values{"urns": {strings.Join(userUrns, ",")}}
	endpoint := c.baseURL + "/users?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("building user directory request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling user directory: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user directory returned status %d", res.StatusCode)
	}

	var users []User
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decoding user directory response: %w", err)
	}
	return users, nil
}
