package api

import "net/http"

// Client represents an HTTP client with authentication capabilities.
type Client struct {
	HTTPClient *http.Client
	Token      string
	BaseURL    string
}

func NewClient(token string) *Client {
	return &Client{
		HTTPClient: &http.Client{},
		Token:      token,
		BaseURL:    "https://api.openai.com/v1",
	}
}

// NewRequest creates a new authenticated HTTP request.
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// Do sends an HTTP request and decodes the response.
func (c *Client) Do(req *http.Request, v interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: Implement response decoding logic here

	return nil
}
