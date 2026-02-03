package api

import (
	"context"
	"fmt"
)

// CollectionsResponse wraps a list of collections.
type CollectionsResponse struct {
	Items  []Collection `json:"items"`
	Result bool         `json:"result"`
}

// CollectionResponse wraps a single collection.
type CollectionResponse struct {
	Item   Collection `json:"item"`
	Result bool       `json:"result"`
}

// CreateCollectionRequest is the payload for creating a collection.
type CreateCollectionRequest struct {
	Title  string `json:"title"`
	Parent *struct {
		ID int `json:"$id"`
	} `json:"parent,omitempty"`
	Color string `json:"color,omitempty"`
}

// UpdateCollectionRequest is the payload for updating a collection.
type UpdateCollectionRequest struct {
	Title  string `json:"title,omitempty"`
	Parent *struct {
		ID int `json:"$id"`
	} `json:"parent,omitempty"`
	Color string `json:"color,omitempty"`
}

// ListRootCollections fetches all root-level collections.
func (c *Client) ListRootCollections(ctx context.Context) ([]Collection, error) {
	var resp CollectionsResponse
	if err := c.Get(ctx, "/collections", &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// ListChildCollections fetches all child collections.
func (c *Client) ListChildCollections(ctx context.Context) ([]Collection, error) {
	var resp CollectionsResponse
	if err := c.Get(ctx, "/collections/childrens", &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// ListAllCollections fetches both root and child collections.
func (c *Client) ListAllCollections(ctx context.Context) ([]Collection, error) {
	root, err := c.ListRootCollections(ctx)
	if err != nil {
		return nil, err
	}

	children, err := c.ListChildCollections(ctx)
	if err != nil {
		return nil, err
	}

	return append(root, children...), nil
}

// GetCollection fetches a single collection by ID.
func (c *Client) GetCollection(ctx context.Context, id int) (*Collection, error) {
	var resp CollectionResponse
	if err := c.Get(ctx, fmt.Sprintf("/collection/%d", id), &resp); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// CreateCollection creates a new collection.
func (c *Client) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*Collection, error) {
	var resp CollectionResponse
	if err := c.Post(ctx, "/collection", req, &resp); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// UpdateCollection updates an existing collection.
func (c *Client) UpdateCollection(ctx context.Context, id int, req *UpdateCollectionRequest) (*Collection, error) {
	var resp CollectionResponse
	if err := c.Put(ctx, fmt.Sprintf("/collection/%d", id), req, &resp); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// DeleteCollection deletes a collection.
func (c *Client) DeleteCollection(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/collection/%d", id))
}
