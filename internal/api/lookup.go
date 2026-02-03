package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// SystemCollectionAll represents all raindrops.
const SystemCollectionAll = 0

// SystemCollectionUnsorted represents unsorted raindrops.
const SystemCollectionUnsorted = -1

// SystemCollectionTrash represents trashed raindrops.
const SystemCollectionTrash = -99

// ResolveCollection converts a name or ID string to a collection ID.
// Supports: numeric ID, "all" (0), "unsorted" (-1), "trash" (-99), or name lookup.
func (c *Client) ResolveCollection(ctx context.Context, nameOrID string) (int, error) {
	s := strings.TrimSpace(strings.ToLower(nameOrID))

	// Handle system collections
	switch s {
	case "", "all", "0":
		return SystemCollectionAll, nil
	case "unsorted", "-1":
		return SystemCollectionUnsorted, nil
	case "trash", "-99":
		return SystemCollectionTrash, nil
	}

	// Handle numeric IDs
	if id, err := strconv.Atoi(nameOrID); err == nil {
		return id, nil
	}

	// Name lookup
	collections, err := c.ListAllCollections(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetch collections for lookup: %w", err)
	}

	// Case-insensitive title match
	for _, col := range collections {
		if strings.EqualFold(col.Title, nameOrID) {
			return col.ID, nil
		}
	}

	return 0, &APIError{
		StatusCode: 404,
		Message:    "collection not found",
		Details:    nameOrID,
	}
}
