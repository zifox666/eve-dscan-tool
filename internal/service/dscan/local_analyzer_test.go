package dscan

import (
	"context"
	"testing"

	"github.com/zifox666/eve-dscan-tool/internal/service/esi"
)

func TestAnalyzeLocal(t *testing.T) {
	allianceID := int64(300)
	client := fakeESIClient{
		characterIDs: map[string]int64{
			"Pilot One": 100,
			"Pilot Two": 101,
		},
		affiliations: []esi.Affiliation{
			{CharacterID: 100, CorporationID: 200, AllianceID: &allianceID},
			{CharacterID: 101, CorporationID: 200, AllianceID: &allianceID},
		},
		names: map[int64]esi.EntityName{
			100: {ID: 100, Name: "Pilot One", Category: "character"},
			101: {ID: 101, Name: "Pilot Two", Category: "character"},
			200: {ID: 200, Name: "Corp Name [CORP]", Category: "corporation"},
			300: {ID: 300, Name: "Alliance Name [ALLY]", Category: "alliance"},
		},
	}

	result, err := AnalyzeLocal(context.Background(), client, "Pilot One\nPilot Two\n")
	if err != nil {
		t.Fatalf("analyze local: %v", err)
	}

	if result.Stats.CharacterCount != 2 {
		t.Fatalf("expected 2 characters, got %d", result.Stats.CharacterCount)
	}
	if result.Corporations["200"].Ticker != "CORP" {
		t.Fatalf("expected corp ticker CORP, got %q", result.Corporations["200"].Ticker)
	}
	if result.Alliances["300"].CorporationCount != 1 {
		t.Fatalf("expected 1 alliance corporation, got %d", result.Alliances["300"].CorporationCount)
	}
}

type fakeESIClient struct {
	characterIDs map[string]int64
	affiliations []esi.Affiliation
	names        map[int64]esi.EntityName
}

func (f fakeESIClient) CharacterIDs(ctx context.Context, names []string) (map[string]int64, error) {
	return f.characterIDs, nil
}

func (f fakeESIClient) CharacterAffiliations(ctx context.Context, characterIDs []int64) ([]esi.Affiliation, error) {
	return f.affiliations, nil
}

func (f fakeESIClient) EntityNames(ctx context.Context, entityIDs []int64) (map[int64]esi.EntityName, error) {
	return f.names, nil
}
