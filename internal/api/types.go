package api

import "time"

// User represents a Raindrop.io user.
type User struct {
	ID       int    `json:"_id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Pro      bool   `json:"pro"`
}

// CollectionRef is a reference to a collection (used in API responses).
type CollectionRef struct {
	ID int `json:"$id"`
}

// Collection represents a Raindrop.io collection.
type Collection struct {
	ID       int            `json:"_id"`
	Title    string         `json:"title"`
	Count    int            `json:"count"`
	Color    string         `json:"color,omitempty"`
	Parent   *CollectionRef `json:"parent,omitempty"`
	Expanded bool           `json:"expanded"`
	Created  time.Time      `json:"created"`
	Updated  time.Time      `json:"lastUpdate"`
}

// ParentID returns the parent collection ID, or 0 if no parent.
func (c *Collection) ParentID() int {
	if c.Parent == nil {
		return 0
	}

	return c.Parent.ID
}

// Raindrop represents a bookmark in Raindrop.io.
type Raindrop struct {
	ID         int            `json:"_id"`
	Link       string         `json:"link"`
	Title      string         `json:"title"`
	Excerpt    string         `json:"excerpt"`
	Note       string         `json:"note"`
	Type       string         `json:"type"`
	Tags       []string       `json:"tags"`
	Important  bool           `json:"important"`
	Collection *CollectionRef `json:"collection,omitempty"`
	Cover      string         `json:"cover"`
	Domain     string         `json:"domain"`
	Created    time.Time      `json:"created"`
	Updated    time.Time      `json:"lastUpdate"`
	Highlights []Highlight    `json:"highlights,omitempty"`
}

// CollectionID returns the collection ID, or 0 if none.
func (r *Raindrop) CollectionID() int {
	if r.Collection == nil {
		return 0
	}

	return r.Collection.ID
}

// Highlight represents a text highlight in a raindrop.
type Highlight struct {
	ID      string    `json:"_id"`
	Text    string    `json:"text"`
	Note    string    `json:"note,omitempty"`
	Color   string    `json:"color,omitempty"`
	Created time.Time `json:"created"`
}

// Tag represents a tag with usage count.
type Tag struct {
	Tag   string `json:"_id"`
	Count int    `json:"count"`
}

// ParsedURL represents the response from /import/url/parse.
type ParsedURL struct {
	Result bool `json:"result"`
	Item   struct {
		Title   string `json:"title"`
		Excerpt string `json:"excerpt"`
		Type    string `json:"type"`
		Cover   string `json:"cover"`
	} `json:"item"`
}
