package api

import (
	"context"
	"net/url"
)

// ParseURL fetches metadata for a URL (title, excerpt, type, cover).
func (c *Client) ParseURL(ctx context.Context, rawURL string) (*ParsedURL, error) {
	params := url.Values{}
	params.Set("url", rawURL)

	path := "/import/url/parse?" + params.Encode()

	var resp ParsedURL
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
