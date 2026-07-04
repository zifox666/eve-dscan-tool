package esi

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type CacheStore interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type Cache struct {
	store          CacheStore
	characterTTL   time.Duration
	affiliationTTL time.Duration
	entityNameTTL  time.Duration
}

func NewCache(store CacheStore, characterTTL time.Duration, affiliationTTL time.Duration, entityNameTTL time.Duration) *Cache {
	return &Cache{
		store:          store,
		characterTTL:   characterTTL,
		affiliationTTL: affiliationTTL,
		entityNameTTL:  entityNameTTL,
	}
}

func (c *Cache) GetCharacterID(ctx context.Context, name string) (int64, bool, error) {
	var payload CharacterIDCacheItem
	found, err := c.store.Get(ctx, CharacterIDKey(name), &payload)
	if err != nil || !found {
		return 0, found, err
	}

	return payload.ID, true, nil
}

func (c *Cache) SetCharacterID(ctx context.Context, name string, id int64) error {
	return c.store.Set(ctx, CharacterIDKey(name), CharacterIDCacheItem{ID: id}, c.characterTTL)
}

func (c *Cache) GetAffiliation(ctx context.Context, characterID int64) (*Affiliation, bool, error) {
	var payload Affiliation
	found, err := c.store.Get(ctx, AffiliationKey(characterID), &payload)
	if err != nil || !found {
		return nil, found, err
	}

	return &payload, true, nil
}

func (c *Cache) SetAffiliation(ctx context.Context, affiliation Affiliation) error {
	return c.store.Set(ctx, AffiliationKey(affiliation.CharacterID), affiliation, c.affiliationTTL)
}

func (c *Cache) GetEntityName(ctx context.Context, entityID int64) (*EntityName, bool, error) {
	var payload EntityName
	found, err := c.store.Get(ctx, EntityNameKey(entityID), &payload)
	if err != nil || !found {
		return nil, found, err
	}

	return &payload, true, nil
}

func (c *Cache) SetEntityName(ctx context.Context, entity EntityName) error {
	return c.store.Set(ctx, EntityNameKey(entity.ID), entity, c.entityNameTTL)
}

func CharacterIDKey(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	sum := sha1.Sum([]byte(normalized))
	return "esi:character:id:" + hex.EncodeToString(sum[:])
}

func AffiliationKey(characterID int64) string {
	return "esi:character:affiliation:" + strconv.FormatInt(characterID, 10)
}

func EntityNameKey(entityID int64) string {
	return "esi:entity:name:" + strconv.FormatInt(entityID, 10)
}

type CharacterIDCacheItem struct {
	ID int64 `json:"id"`
}

type Affiliation struct {
	CharacterID   int64  `json:"character_id"`
	CorporationID int64  `json:"corporation_id"`
	AllianceID    *int64 `json:"alliance_id,omitempty"`
}

type EntityName struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}
