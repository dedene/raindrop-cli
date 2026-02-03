package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListOptions configures list/search requests.
type ListOptions struct {
	Search  string
	Sort    string
	Page    int
	PerPage int
}

// RaindropResponse wraps a single raindrop response.
type RaindropResponse struct {
	Item   Raindrop `json:"item"`
	Result bool     `json:"result"`
}

// RaindropsResponse wraps a list of raindrops.
type RaindropsResponse struct {
	Items []Raindrop `json:"items"`
	Count int        `json:"count"`
}

// CreateRaindropRequest is the payload for creating a raindrop.
type CreateRaindropRequest struct {
	Link       string   `json:"link,omitempty"`
	Title      string   `json:"title,omitempty"`
	Excerpt    string   `json:"excerpt,omitempty"`
	Note       string   `json:"note,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Important  bool     `json:"important,omitempty"`
	Collection struct {
		ID int `json:"$id"`
	} `json:"collection,omitempty"`
	PleaseParse bool `json:"pleaseParse,omitempty"` //nolint:tagliatelle // API uses camelCase
}

// UpdateRaindropRequest is the payload for updating a raindrop.
type UpdateRaindropRequest struct {
	Link       string   `json:"link,omitempty"`
	Title      string   `json:"title,omitempty"`
	Excerpt    string   `json:"excerpt,omitempty"`
	Note       string   `json:"note,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Important  *bool    `json:"important,omitempty"`
	Collection *struct {
		ID int `json:"$id"`
	} `json:"collection,omitempty"`
}

// ListRaindrops fetches raindrops from a collection.
// collectionID: 0 = all, -1 = unsorted, -99 = trash
func (c *Client) ListRaindrops(ctx context.Context, collectionID int, opts ListOptions) (*RaindropsResponse, error) {
	params := url.Values{}

	if opts.Search != "" {
		params.Set("search", opts.Search)
	}

	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}

	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}

	perPage := opts.PerPage
	if perPage == 0 {
		perPage = 50
	}

	params.Set("perpage", strconv.Itoa(perPage))

	path := fmt.Sprintf("/raindrops/%d", collectionID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp RaindropsResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetRaindrop fetches a single raindrop by ID.
func (c *Client) GetRaindrop(ctx context.Context, id int) (*Raindrop, error) {
	var resp RaindropResponse
	if err := c.Get(ctx, fmt.Sprintf("/raindrop/%d", id), &resp); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// CreateRaindrop creates a new raindrop.
func (c *Client) CreateRaindrop(ctx context.Context, req *CreateRaindropRequest) (*Raindrop, error) {
	var resp RaindropResponse
	if err := c.Post(ctx, "/raindrop", req, &resp); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// UpdateRaindrop updates an existing raindrop.
func (c *Client) UpdateRaindrop(ctx context.Context, id int, req *UpdateRaindropRequest) (*Raindrop, error) {
	var resp RaindropResponse
	if err := c.Put(ctx, fmt.Sprintf("/raindrop/%d", id), req, &resp); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// DeleteRaindrop moves a raindrop to trash, or permanently deletes if permanent=true.
func (c *Client) DeleteRaindrop(ctx context.Context, id int, permanent bool) error {
	path := fmt.Sprintf("/raindrop/%d", id)
	if permanent {
		path += "?permanent=true"
	}

	return c.Delete(ctx, path)
}

// BulkCreateRequest wraps bulk create payload.
type BulkCreateRequest struct {
	Items []CreateRaindropRequest `json:"items"`
}

// BulkCreateResponse wraps bulk create response.
type BulkCreateResponse struct {
	Items  []Raindrop `json:"items"`
	Result bool       `json:"result"`
}

// CreateRaindropsBulk creates multiple raindrops (max 100 per call).
func (c *Client) CreateRaindropsBulk(ctx context.Context, items []CreateRaindropRequest) ([]Raindrop, error) {
	req := BulkCreateRequest{Items: items}

	var resp BulkCreateResponse
	if err := c.Post(ctx, "/raindrops", &req, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}
