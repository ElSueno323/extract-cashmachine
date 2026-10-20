package files

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Item struct {
	Name        string
	ContentType string
	Data        []byte
	Created     time.Time
}

type Store struct {
	mu    sync.Mutex
	items map[string]Item
	ttl   time.Duration
}

func New() *Store {
	store := &Store{
		items: map[string]Item{},
		ttl:   2 * time.Hour,
	}
	go store.cleanup()
	return store
}

func (s *Store) Put(name, contentType string, data []byte) string {
	token := newToken()
	s.mu.Lock()
	s.items[token] = Item{
		Name:        name,
		ContentType: contentType,
		Data:        data,
		Created:     time.Now(),
	}
	s.mu.Unlock()
	return token
}

func (s *Store) Get(token string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[token]
	if !ok || time.Since(item.Created) > s.ttl {
		delete(s.items, token)
		return Item{}, false
	}
	return item, true
}

func (s *Store) cleanup() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for token, item := range s.items {
			if time.Since(item.Created) > s.ttl {
				delete(s.items, token)
			}
		}
		s.mu.Unlock()
	}
}

func newToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buf)
}
