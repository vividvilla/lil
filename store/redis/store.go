package redis

import (
	"crypto/md5"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"gitlab.zerodha.tech/commons/lil/store"
)

// Store represents redis session store for simple sessions.
// Each session is stored as redis hashmap.
type Store struct {
	// Prefix for session id.
	shortURLPrefix string
	fullURLPrefix  string
	// Redis pool
	pool *redis.Pool
}

const (
	defaultShortURLPrefix = "SHORT:"
	defaultFullURLPrefix  = "FULL:"
)

// New creates a new in-memory store instance
func New(pool *redis.Pool) *Store {
	return &Store{
		pool:           pool,
		shortURLPrefix: defaultShortURLPrefix,
		fullURLPrefix:  defaultFullURLPrefix,
	}
}

func (s *Store) md5(val string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(val)))
}

func (s *Store) shortURLKey(url string) string {
	return fmt.Sprintf("%s%s", s.shortURLPrefix, s.md5(url))
}

func (s *Store) fullURLKey(url string) string {
	return fmt.Sprintf("%s%s", s.fullURLPrefix, s.md5(url))
}

// GetFullURL retrives full url for given short url it exists else send ErrNotFound.
func (s *Store) GetFullURL(shortURL string) (string, error) {
	conn := s.pool.Get()
	defer conn.Close()
	v, err := redis.String(conn.Do("GET", s.shortURLKey(shortURL)))
	if err == redis.ErrNil {
		return "", store.ErrNotFound
	} else if err != nil {
		return "", err
	}
	return v, nil
}

// GetShortURL retrives short url from a long url if it exists else send ErrNotFound.
func (s *Store) GetShortURL(fullURL string) (string, error) {
	conn := s.pool.Get()
	defer conn.Close()
	v, err := redis.String(conn.Do("GET", s.fullURLKey(fullURL)))
	if err == redis.ErrNil {
		return "", store.ErrNotFound
	} else if err != nil {
		return "", err
	}
	return v, nil
}

// Set stores full url against short url (stores reverse map if needed).
func (s *Store) Set(shortURL string, fullURL string) error {
	conn := s.pool.Get()
	defer conn.Close()
	conn.Send("MULTI")
	conn.Send("SET", s.shortURLKey(shortURL), fullURL)
	conn.Send("SET", s.fullURLKey(fullURL), shortURL)
	rep, err := redis.Values(conn.Do("EXEC"))
	// Check if there are any errors.
	for _, r := range rep {
		if _, ok := r.(redis.Error); ok {
			return err
		}
	}
	return nil
}

// Delete removes shorturl and fullurl map.
func (s *Store) Delete(shortURL string) error {
	fullURL, err := s.GetFullURL(shortURL)
	if err != nil {
		return err
	}
	conn := s.pool.Get()
	defer conn.Close()
	conn.Send("MULTI")
	conn.Send("DEL", s.shortURLKey(shortURL))
	conn.Send("DEL", s.fullURLKey(fullURL))
	rep, err := redis.Values(conn.Do("EXEC"))
	// Check if there are any errors.
	for _, r := range rep {
		if _, ok := r.(redis.Error); ok {
			return err
		}
	}
	return nil
}
