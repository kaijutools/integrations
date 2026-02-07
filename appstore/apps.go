package appstore

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ListApps fetches all apps associated with the account
func (c *Client) ListApps() (*AppsResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/apps", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	// Add query params if you want to sort or limit fields
	q := req.URL.Query()
	q.Add("limit", "20")
	q.Add("sort", "name")
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result AppsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	return &result, nil
}
