package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Jira API error: %d - %s", e.StatusCode, e.Message)
}
func (c *Client) buildURL(endpoint string, params url.Values) string {
	return fmt.Sprintf("%s%s%s?%s", c.config.BaseURL, c.config.VersionAPI, endpoint, params.Encode())
}

func (c *Client) doRequest(ctx context.Context, url string, result interface{}) error {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err = json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
