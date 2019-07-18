package redis

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"gitlab.zerodha.tech/commons/lil/store"
)

// Store represents redis session store for simple sessions.
// Each session is stored as redis hashmap.
type Store struct {
	// Prefix for session id.
	idPrefix  string
	urlPrefix string
	// Redis pool
	pool *redis.Pool
}

const (
	defaultIDPrefix  = "LIL:ID:"
	defaultURLPrefix = "LIL:URL:"
)

// New creates a new in-memory store instance
func New(pool *redis.Pool) *Store {
	return &Store{
		pool:      pool,
		idPrefix:  defaultIDPrefix,
		urlPrefix: defaultURLPrefix,
	}
}

func (s *Store) md5(val string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(val)))
}

func (s *Store) idKey(url string) string {
	return fmt.Sprintf("%s%s", s.idPrefix, s.md5(url))
}

func (s *Store) urlKey(url string) string {
	return fmt.Sprintf("%s%s", s.urlPrefix, s.md5(url))
}

// Get retrives full url for given short url it exists else send ErrNotFound.
func (s *Store) Get(id string) (string, *store.Meta, error) {
	conn := s.pool.Get()
	defer conn.Close()
	vals, err := redis.Values(conn.Do("HMGET", s.idKey(id), "url", "meta"))
	if err != nil {
		return "", nil, err
	}
	var (
		url      string
		metaJSON []byte
	)
	if _, err := redis.Scan(vals, &url, &metaJSON); err != nil {
		return "", nil, err
	}
	if url == "" {
		return "", nil, store.ErrNotFound
	}
	meta := &store.Meta{}
	if err := json.Unmarshal(metaJSON, meta); err != nil {
		return "", nil, fmt.Errorf("couldn't unmarshal meta: %v", err)
	}
	return url, meta, nil
}

// GetID retrives short url from a long url if it exists else send ErrNotFound.
func (s *Store) GetID(url string) (string, error) {
	conn := s.pool.Get()
	defer conn.Close()
	v, err := redis.String(conn.Do("GET", s.urlKey(url)))
	if err == redis.ErrNil {
		return "", store.ErrNotFound
	} else if err != nil {
		return "", err
	}
	return v, nil
}

// Set stores full url against short url (stores reverse map if needed).
func (s *Store) Set(id string, url string, meta *store.Meta) error {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal meta: %v", err)
	}
	conn := s.pool.Get()
	defer conn.Close()
	conn.Send("MULTI")
	// Store full url and meta in hashmap.
	conn.Send("HMSET", s.idKey(id), "url", url, "meta", metaJSON)
	// Store a reverse map of full url and id.
	conn.Send("SET", s.urlKey(url), id)
	rep, err := redis.Values(conn.Do("EXEC"))
	// Check if there are any errors.
	for _, r := range rep {
		if _, ok := r.(redis.Error); ok {
			return err
		}
	}
	return nil
}

// Del removes shorturl and fullurl map.
func (s *Store) Del(id string) error {
	url, _, err := s.Get(id)
	if err != nil {
		return err
	}
	conn := s.pool.Get()
	defer conn.Close()
	conn.Send("MULTI")
	conn.Send("DEL", s.idKey(id))
	conn.Send("DEL", s.urlKey(url))
	rep, err := redis.Values(conn.Do("EXEC"))
	// Check if there are any errors.
	for _, r := range rep {
		if _, ok := r.(redis.Error); ok {
			return err
		}
	}
	return nil
}
