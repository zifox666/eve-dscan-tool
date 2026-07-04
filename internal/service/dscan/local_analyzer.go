package dscan

import (
	"context"
	"strconv"
	"strings"

	"github.com/zifox666/eve-dscan-tool/internal/service/esi"
)

type ESIClient interface {
	CharacterIDs(ctx context.Context, names []string) (map[string]int64, error)
	CharacterAffiliations(ctx context.Context, characterIDs []int64) ([]esi.Affiliation, error)
	EntityNames(ctx context.Context, entityIDs []int64) (map[int64]esi.EntityName, error)
}

func AnalyzeLocal(ctx context.Context, client ESIClient, rawData string) (*LocalResult, error) {
	characterNames := ParseLocalDScan(rawData)

	characterIDsByName, err := client.CharacterIDs(ctx, characterNames)
	if err != nil {
		return nil, err
	}

	characterIDs := make([]int64, 0, len(characterIDsByName))
	for _, id := range characterIDsByName {
		characterIDs = append(characterIDs, id)
	}

	affiliations, err := client.CharacterAffiliations(ctx, characterIDs)
	if err != nil {
		return nil, err
	}

	entityIDs := make([]int64, 0, len(affiliations)*3)
	for _, affiliation := range affiliations {
		entityIDs = append(entityIDs, affiliation.CharacterID, affiliation.CorporationID)
		if affiliation.AllianceID != nil {
			entityIDs = append(entityIDs, *affiliation.AllianceID)
		}
	}

	names, err := client.EntityNames(ctx, entityIDs)
	if err != nil {
		return nil, err
	}

	return organizeLocalDScanData(affiliations, names), nil
}

func ParseLocalDScan(rawData string) []string {
	lines := strings.Split(rawData, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func organizeLocalDScanData(affiliations []esi.Affiliation, names map[int64]esi.EntityName) *LocalResult {
	result := &LocalResult{
		Alliances:    map[string]LocalAlliance{},
		Corporations: map[string]LocalCorporation{},
		Characters:   map[string]LocalCharacter{},
	}

	for _, affiliation := range affiliations {
		characterID := affiliation.CharacterID
		corporationID := affiliation.CorporationID
		allianceID := affiliation.AllianceID

		if entity, ok := names[characterID]; ok {
			result.Characters[idKey(characterID)] = LocalCharacter{
				ID:            characterID,
				Name:          stripTicker(entity.Name),
				CorporationID: corporationID,
				AllianceID:    allianceID,
			}
		}

		if _, exists := result.Corporations[idKey(corporationID)]; !exists {
			if entity, ok := names[corporationID]; ok {
				name, ticker := splitTicker(entity.Name)
				result.Corporations[idKey(corporationID)] = LocalCorporation{
					ID:             corporationID,
					Name:           name,
					Ticker:         ticker,
					AllianceID:     allianceID,
					CharacterCount: 0,
				}
			}
		}

		if corporation, ok := result.Corporations[idKey(corporationID)]; ok {
			corporation.CharacterCount++
			result.Corporations[idKey(corporationID)] = corporation
		}

		if allianceID != nil {
			if _, exists := result.Alliances[idKey(*allianceID)]; !exists {
				if entity, ok := names[*allianceID]; ok {
					name, ticker := splitTicker(entity.Name)
					result.Alliances[idKey(*allianceID)] = LocalAlliance{
						ID:               *allianceID,
						Name:             name,
						Ticker:           ticker,
						CorporationCount: 0,
						CharacterCount:   0,
					}
				}
			}

			if alliance, ok := result.Alliances[idKey(*allianceID)]; ok {
				alliance.CharacterCount++
				result.Alliances[idKey(*allianceID)] = alliance
			}
		}
	}

	seenAllianceCorps := map[string]map[int64]struct{}{}
	for _, corporation := range result.Corporations {
		if corporation.AllianceID == nil {
			continue
		}
		allianceKey := idKey(*corporation.AllianceID)
		if _, ok := seenAllianceCorps[allianceKey]; !ok {
			seenAllianceCorps[allianceKey] = map[int64]struct{}{}
		}
		if _, ok := seenAllianceCorps[allianceKey][corporation.ID]; ok {
			continue
		}
		seenAllianceCorps[allianceKey][corporation.ID] = struct{}{}

		alliance := result.Alliances[allianceKey]
		alliance.CorporationCount++
		result.Alliances[allianceKey] = alliance
	}

	result.Stats.AllianceCount = len(result.Alliances)
	result.Stats.CorporationCount = len(result.Corporations)
	result.Stats.CharacterCount = len(result.Characters)

	return result
}

func splitTicker(value string) (string, string) {
	open := strings.LastIndex(value, "[")
	close := strings.LastIndex(value, "]")
	if open >= 0 && close > open {
		return strings.TrimSpace(value[:open]), strings.TrimSpace(value[open+1 : close])
	}

	return strings.TrimSpace(value), ""
}

func stripTicker(value string) string {
	name, _ := splitTicker(value)
	return name
}

func idKey(id int64) string {
	return strconv.FormatInt(id, 10)
}

type LocalResult struct {
	Alliances    map[string]LocalAlliance    `json:"alliances"`
	Corporations map[string]LocalCorporation `json:"corporations"`
	Characters   map[string]LocalCharacter   `json:"characters"`
	Stats        LocalStats                  `json:"stats"`
}

type LocalAlliance struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Ticker           string `json:"ticker"`
	CorporationCount int    `json:"corporation_count"`
	CharacterCount   int    `json:"character_count"`
}

type LocalCorporation struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Ticker         string `json:"ticker"`
	AllianceID     *int64 `json:"alliance_id"`
	CharacterCount int    `json:"character_count"`
}

type LocalCharacter struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	CorporationID int64  `json:"corporation_id"`
	AllianceID    *int64 `json:"alliance_id"`
}

type LocalStats struct {
	AllianceCount    int `json:"alliance_count"`
	CorporationCount int `json:"corporation_count"`
	CharacterCount   int `json:"character_count"`
}
