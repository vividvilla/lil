package store

import "errors"

var (
	// ErrNotFound raised when something not found.
	ErrNotFound = errors.New("Not found")
)

// Meta represents.
type Meta struct {
	OGTags []*OGTag `json:"og_tags,omitempty"`
	// Page title
	Title string `json:"title,omitempty"`
}

// OGTag represents og tags.
type OGTag struct {
	Property string `json:"property"`
	Content  string `json:"content"`
}

// Store represents backend storage for storing short urls.
type Store interface {
	// Get id and meta for given full URL.
	GetID(url string) (id string, err error)
	// Get retrives payload from given id.
	Get(id string) (url string, meta *Meta, err error)
	// Set stores full url against short url (stores reverse map if needed).
	Set(id string, url string, meta *Meta) (err error)
	// Delete removes short url and meta from store.
	Del(id string) (err error)
}
