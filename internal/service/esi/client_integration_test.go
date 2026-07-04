package esi_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/zifox666/eve-dscan-tool/internal/config"
	"github.com/zifox666/eve-dscan-tool/internal/service/esi"
)

func TestIntegrationESIClient(t *testing.T) {
	if os.Getenv("RUN_ESI_INTEGRATION") != "1" {
		t.Skip("set RUN_ESI_INTEGRATION=1 to run")
	}

	cfg := config.Load()
	client := esi.NewClient(&http.Client{Timeout: cfg.HTTPTimeout}, nil, esi.ClientConfig{
		BaseURL:      cfg.ESIBaseURL,
		Datasource:   cfg.ESIDatasource,
		Language:     cfg.ESILanguage,
		MaxRetries:   cfg.ESIMaxRetries,
		RetryDelay:   time.Second,
		IDsBatchSize: cfg.ESIIDsBatchSize,
		BatchSize:    cfg.ESIBatchSize,
	})

	ids, err := client.CharacterIDs(t.Context(), []string{"Chribba"})
	if err != nil {
		t.Fatalf("character ids: %v", err)
	}
	characterID := ids["Chribba"]
	if characterID == 0 {
		t.Fatalf("expected Chribba id")
	}

	affiliations, err := client.CharacterAffiliations(t.Context(), []int64{characterID})
	if err != nil {
		t.Fatalf("affiliations: %v", err)
	}
	if len(affiliations) != 1 || affiliations[0].CharacterID != characterID {
		t.Fatalf("unexpected affiliations %#v", affiliations)
	}

	names, err := client.EntityNames(t.Context(), []int64{characterID, affiliations[0].CorporationID})
	if err != nil {
		t.Fatalf("entity names: %v", err)
	}
	if names[characterID].Name == "" {
		t.Fatalf("expected character name for %d", characterID)
	}
}
