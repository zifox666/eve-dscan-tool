package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	cache      *Cache
	config     ClientConfig
}

type ClientConfig struct {
	BaseURL      string
	Datasource   string
	Language     string
	MaxRetries   int
	RetryDelay   time.Duration
	IDsBatchSize int
	BatchSize    int
}

func NewClient(httpClient *http.Client, cache *Cache, config ClientConfig) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 0
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = time.Second
	}
	if config.IDsBatchSize <= 0 {
		config.IDsBatchSize = 500
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 1000
	}
	if config.Datasource == "" {
		config.Datasource = "tranquility"
	}
	if config.Language == "" {
		config.Language = "en"
	}

	config.BaseURL = strings.TrimRight(config.BaseURL, "/")

	return &Client{
		httpClient: httpClient,
		cache:      cache,
		config:     config,
	}
}

func (c *Client) CharacterIDs(ctx context.Context, names []string) (map[string]int64, error) {
	result := make(map[string]int64)
	uniqueNames := uniqueStrings(names)
	var uncached []string

	for _, name := range uniqueNames {
		if c.cache != nil {
			id, found, err := c.cache.GetCharacterID(ctx, name)
			if err == nil && found {
				result[name] = id
				continue
			}
		}
		uncached = append(uncached, name)
	}

	for _, batch := range chunkStrings(uncached, c.config.IDsBatchSize) {
		var response universeIDsResponse
		if err := c.post(ctx, "/universe/ids/", queryValues{
			"datasource": c.config.Datasource,
			"language":   c.config.Language,
		}, batch, &response); err != nil {
			return result, err
		}

		for _, character := range response.Characters {
			result[character.Name] = character.ID
			if c.cache != nil {
				_ = c.cache.SetCharacterID(ctx, character.Name, character.ID)
			}
		}
	}

	return result, nil
}

func (c *Client) CharacterAffiliations(ctx context.Context, characterIDs []int64) ([]Affiliation, error) {
	var result []Affiliation
	var uncached []int64

	for _, characterID := range uniqueInt64s(characterIDs) {
		if c.cache != nil {
			affiliation, found, err := c.cache.GetAffiliation(ctx, characterID)
			if err == nil && found {
				result = append(result, *affiliation)
				continue
			}
		}
		uncached = append(uncached, characterID)
	}

	for _, batch := range chunkInt64s(uncached, c.config.BatchSize) {
		var affiliations []Affiliation
		if err := c.post(ctx, "/characters/affiliation/", queryValues{
			"datasource": c.config.Datasource,
		}, batch, &affiliations); err != nil {
			return result, err
		}

		for _, affiliation := range affiliations {
			result = append(result, affiliation)
			if c.cache != nil {
				_ = c.cache.SetAffiliation(ctx, affiliation)
			}
		}
	}

	return result, nil
}

func (c *Client) EntityNames(ctx context.Context, entityIDs []int64) (map[int64]EntityName, error) {
	result := make(map[int64]EntityName)
	var uncached []int64

	for _, entityID := range uniqueInt64s(entityIDs) {
		if c.cache != nil {
			entity, found, err := c.cache.GetEntityName(ctx, entityID)
			if err == nil && found {
				result[entityID] = *entity
				continue
			}
		}
		uncached = append(uncached, entityID)
	}

	for _, batch := range chunkInt64s(uncached, c.config.BatchSize) {
		var names []EntityName
		if err := c.post(ctx, "/universe/names/", queryValues{
			"datasource": c.config.Datasource,
		}, batch, &names); err != nil {
			return result, err
		}

		for _, entity := range names {
			result[entity.ID] = entity
			if c.cache != nil {
				_ = c.cache.SetEntityName(ctx, entity)
			}
		}
	}

	return result, nil
}

func (c *Client) post(ctx context.Context, path string, query queryValues, payload any, dest any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			timer := time.NewTimer(c.config.RetryDelay * time.Duration(attempt))
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(path, query), bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Language", c.config.Language)
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		err = decodeESIResponse(resp, dest)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return lastErr
}

func (c *Client) url(path string, query queryValues) string {
	values := url.Values{}
	for key, value := range query {
		if value != "" {
			values.Set(key, value)
		}
	}

	encoded := values.Encode()
	if encoded == "" {
		return c.config.BaseURL + path
	}

	return c.config.BaseURL + path + "?" + encoded
}

func decodeESIResponse(resp *http.Response, dest any) error {
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("esi status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

type queryValues map[string]string

type universeIDsResponse struct {
	Characters []universeIDCharacter `json:"characters"`
}

type universeIDCharacter struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}

	return result
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))

	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func chunkStrings(values []string, size int) [][]string {
	if size <= 0 {
		size = len(values)
	}

	var result [][]string
	for start := 0; start < len(values); start += size {
		end := start + size
		if end > len(values) {
			end = len(values)
		}
		result = append(result, values[start:end])
	}

	return result
}

func chunkInt64s(values []int64, size int) [][]int64 {
	if size <= 0 {
		size = len(values)
	}

	var result [][]int64
	for start := 0; start < len(values); start += size {
		end := start + size
		if end > len(values) {
			end = len(values)
		}
		result = append(result, values[start:end])
	}

	return result
}
