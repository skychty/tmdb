package service

import (
	"context"
	"encoding/json"
	"time"

	"tmdb/internal/cache"
)

type cacheEnvelope struct {
	FreshUntil time.Time       `json:"fresh_until"`
	Payload    json.RawMessage `json:"payload"`
}

type CacheStore struct {
	cache    cache.Cache
	freshTTL time.Duration
	staleTTL time.Duration
}

func NewCacheStore(c cache.Cache, freshTTL, staleTTL time.Duration) *CacheStore {
	return &CacheStore{
		cache:    c,
		freshTTL: freshTTL,
		staleTTL: staleTTL,
	}
}

func (s *CacheStore) GetFresh(ctx context.Context, key string) ([]byte, bool, error) {
	data, ok, err := s.cache.Get(ctx, key)
	if err != nil || !ok {
		return nil, false, err
	}

	payload, fresh, err := parseEnvelope(data)
	if err != nil {
		return nil, false, err
	}
	if !fresh {
		return nil, false, nil
	}
	return payload, true, nil
}

func (s *CacheStore) GetStale(ctx context.Context, key string) ([]byte, bool, error) {
	data, ok, err := s.cache.Get(ctx, key)
	if err != nil || !ok {
		return nil, false, err
	}

	payload, _, err := parseEnvelope(data)
	if err != nil {
		return data, true, nil
	}
	if len(payload) == 0 {
		return nil, false, nil
	}
	return payload, true, nil
}

func (s *CacheStore) Set(ctx context.Context, key string, payload []byte) error {
	entry := cacheEnvelope{
		FreshUntil: time.Now().UTC().Add(s.freshTTL),
		Payload:    payload,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, data, s.staleTTL)
}

func parseEnvelope(data []byte) (payload []byte, fresh bool, err error) {
	var entry cacheEnvelope
	if err := json.Unmarshal(data, &entry); err != nil {
		return data, true, nil
	}
	if len(entry.Payload) == 0 {
		return nil, false, nil
	}
	return entry.Payload, time.Now().UTC().Before(entry.FreshUntil), nil
}
