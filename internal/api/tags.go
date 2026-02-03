package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// TagsResponse wraps a list of tags.
type TagsResponse struct {
	Items []Tag `json:"items"`
}

// ListTags fetches all tags for a collection.
// collectionID: 0 = all collections
func (c *Client) ListTags(ctx context.Context, collectionID int) ([]Tag, error) {
	var resp TagsResponse
	if err := c.Get(ctx, fmt.Sprintf("/tags/%d", collectionID), &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// RenameTagRequest is the payload for renaming a tag.
type RenameTagRequest struct {
	Tags    []string `json:"tags"`
	Replace string   `json:"replace"`
}

// RenameTags renames tags in a collection.
func (c *Client) RenameTags(ctx context.Context, collectionID int, oldTags []string, newName string) error {
	req := RenameTagRequest{
		Tags:    oldTags,
		Replace: newName,
	}

	return c.Put(ctx, fmt.Sprintf("/tags/%d", collectionID), &req, nil)
}

// DeleteTagsRequest is the payload for deleting tags.
type DeleteTagsRequest struct {
	Tags []string `json:"tags"`
}

// DeleteTags removes tags from a collection.
func (c *Client) DeleteTags(ctx context.Context, collectionID int, tags []string) error {
	// Raindrop API uses DELETE with tags as query param
	return c.Delete(ctx, fmt.Sprintf("/tags/%d?tags=%s", collectionID, encodeTagsParam(tags)))
}

func encodeTagsParam(tags []string) string {
	// URL-encode each tag and join with commas
	encoded := make([]string, len(tags))

	for i, t := range tags {
		encoded[i] = url.QueryEscape(t)
	}

	return strings.Join(encoded, ",")
}
