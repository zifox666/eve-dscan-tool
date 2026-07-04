package dscan

import (
	"context"
	"encoding/json"
	"time"
)

type CacheStore interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type ResultCache struct {
	store CacheStore
	ttl   time.Duration
}

func NewResultCache(store CacheStore, ttl time.Duration) *ResultCache {
	return &ResultCache{
		store: store,
		ttl:   ttl,
	}
}

func (c *ResultCache) GetLocal(ctx context.Context, shortID string) (*CachedResult, bool, error) {
	var result CachedResult
	found, err := c.store.Get(ctx, LocalCacheKey(shortID), &result)
	if err != nil || !found {
		return nil, found, err
	}

	return &result, true, nil
}

func (c *ResultCache) SetLocal(ctx context.Context, result CachedResult) error {
	return c.store.Set(ctx, LocalCacheKey(result.ShortID), result, c.ttl)
}

func (c *ResultCache) DeleteLocal(ctx context.Context, shortID string) error {
	return c.store.Delete(ctx, LocalCacheKey(shortID))
}

func (c *ResultCache) GetShip(ctx context.Context, lang string, shortID string) (*CachedResult, bool, error) {
	var result CachedResult
	found, err := c.store.Get(ctx, ShipCacheKey(lang, shortID), &result)
	if err != nil || !found {
		return nil, found, err
	}

	return &result, true, nil
}

func (c *ResultCache) SetShip(ctx context.Context, lang string, result CachedResult) error {
	return c.store.Set(ctx, ShipCacheKey(lang, result.ShortID), result, c.ttl)
}

func (c *ResultCache) DeleteShip(ctx context.Context, shortID string) error {
	return c.store.Delete(ctx, ShipCacheKey("zh", shortID), ShipCacheKey("en", shortID))
}

func LocalCacheKey(shortID string) string {
	return "dscan:local:" + shortID
}

func ShipCacheKey(lang string, shortID string) string {
	if lang != "en" {
		lang = "zh"
	}

	return "dscan:ship:" + lang + ":" + shortID
}

type CachedResult struct {
	ID            uint64          `json:"id"`
	ShortID       string          `json:"short_id"`
	ProcessedData json.RawMessage `json:"processed_data"`
	ViewCount     int64           `json:"view_count"`
	CreatedAt     time.Time       `json:"created_at"`
	TimeAgo       string          `json:"time_ago,omitempty"`
	Lang          string          `json:"lang,omitempty"`
}
