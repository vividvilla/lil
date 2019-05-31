package store

import "errors"

var (
	// ErrNotFound raised when something not found.
	ErrNotFound = errors.New("Not found")
)

// Store represents backend storage for storing short urls.
type Store interface {
	// GetFullURL retrives full url for given short url it exists else send ErrNotFound.
	GetFullURL(shortURL string) (string, error)
	// GetShortURL retrives short url from a long url if it exists else send ErrNotFound.
	GetShortURL(fullURL string) (string, error)
	// Set stores full url against short url (stores reverse map if needed).
	Set(shortURL string, fullURL string) error
	// Delete removes shorturl-fullurl if it exists.
	Delete(shortURL string) error
}
