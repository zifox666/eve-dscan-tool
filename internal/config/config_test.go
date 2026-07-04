package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DISABLE_DOTENV", "1")
	t.Setenv("APP_NAME", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("REDIS_DB", "")

	cfg := Load()

	if cfg.AppName != defaultAppName {
		t.Fatalf("expected default app name %q, got %q", defaultAppName, cfg.AppName)
	}
	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("expected default http addr %q, got %q", defaultHTTPAddr, cfg.HTTPAddr)
	}
	if cfg.RedisDB != 0 {
		t.Fatalf("expected default redis db 0, got %d", cfg.RedisDB)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("DISABLE_DOTENV", "1")
	t.Setenv("APP_NAME", "Test DScan")
	t.Setenv("HTTP_ADDR", ":9000")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("SHORT_LINK_LENGTH", "12")
	t.Setenv("DSCAN_CACHE_TTL", "2h")
	t.Setenv("ESI_AFFILIATION_CACHE_TTL", "168h")

	cfg := Load()

	if cfg.AppName != "Test DScan" {
		t.Fatalf("expected overridden app name, got %q", cfg.AppName)
	}
	if cfg.HTTPAddr != ":9000" {
		t.Fatalf("expected overridden http addr, got %q", cfg.HTTPAddr)
	}
	if cfg.RedisDB != 2 {
		t.Fatalf("expected overridden redis db 2, got %d", cfg.RedisDB)
	}
	if cfg.ShortLinkLength != 12 {
		t.Fatalf("expected overridden short link length 12, got %d", cfg.ShortLinkLength)
	}
	if cfg.DScanCacheTTL != 2*time.Hour {
		t.Fatalf("expected dscan cache ttl 2h, got %s", cfg.DScanCacheTTL)
	}
	if cfg.ESIAffiliationTTL != 7*24*time.Hour {
		t.Fatalf("expected esi affiliation ttl 168h, got %s", cfg.ESIAffiliationTTL)
	}
}
