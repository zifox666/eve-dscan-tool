package dscan

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zifox666/eve-dscan-tool/internal/service/sde"
)

var systemNamePattern = regexp.MustCompile(`^(.+?)( [XIV]+)? - `)

var capitalGroupIDs = map[int64]struct{}{
	30: {}, 485: {}, 547: {}, 659: {}, 883: {}, 1538: {}, 4594: {},
}

var structureCategoryIDs = map[int64]struct{}{
	41: {}, 46: {}, 65: {}, 66: {},
}

type SDEClient interface {
	TypeInfo(ctx context.Context, typeID int64, language string) (*sde.TypeInfo, error)
	SystemIDByName(ctx context.Context, name string, language string) (int64, error)
	SystemInfoByID(ctx context.Context, systemID int64) (*sde.SystemInfo, error)
}

func ParseShipDScan(rawData string) []ShipItem {
	lines := strings.Split(rawData, "\n")
	items := make([]ShipItem, 0, len(lines))

	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), "\t")
		if len(parts) < 2 {
			continue
		}

		typeID, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			continue
		}

		item := ShipItem{
			TypeID: typeID,
			Name:   strings.TrimSpace(parts[1]),
		}
		if len(parts) >= 3 {
			value := strings.TrimSpace(parts[2])
			item.TypeName = &value
		}
		if len(parts) >= 4 {
			value := strings.TrimSpace(parts[3])
			item.Distance = &value
		}

		items = append(items, item)
	}

	return items
}

func AnalyzeShip(ctx context.Context, sdeClient SDEClient, rawData string, language string, filterDistance bool) (*ShipResult, error) {
	return OrganizeShipDScanData(ctx, sdeClient, ParseShipDScan(rawData), language, filterDistance)
}

func OrganizeShipDScanData(ctx context.Context, sdeClient SDEClient, shipItems []ShipItem, language string, filterDistance bool) (*ShipResult, error) {
	result := &ShipResult{
		ShipTypes:         map[string][]TypeCount{},
		CapitalTypes:      map[string][]TypeCount{},
		StructureTypes:    map[string][]TypeCount{},
		MiscTypes:         map[string][]TypeCount{},
		CapitalGroupIDs:   []int64{30, 485, 547, 659, 883, 1538, 4594},
		FilterDistance:    filterDistance,
		internalShip:      map[string]map[string]internalTypeEntry{},
		internalCapital:   map[string]map[string]internalTypeEntry{},
		internalStructure: map[string]map[string]internalTypeEntry{},
		internalMisc:      map[string]map[string]internalTypeEntry{},
	}
	result.Stats.TotalCount = len(shipItems)

	systemCandidates := map[string]int{}

	for _, item := range shipItems {
		info, err := sdeClient.TypeInfo(ctx, item.TypeID, language)
		if err != nil {
			if err == sde.ErrNotFound {
				continue
			}
			return nil, err
		}

		if info.CategoryID == 65 || info.CategoryID == 2 {
			if match := systemNamePattern.FindStringSubmatch(item.Name); len(match) > 1 {
				systemCandidates[strings.TrimSpace(match[1])]++
			}
		}

		if filterDistance && (item.Distance == nil || *item.Distance == "" || *item.Distance == "-") {
			continue
		}

		switch {
		case info.CategoryID == 6:
			incrementType(result.internalShip, info.GroupName, info.Name, item.TypeID)
			result.Stats.ShipCount++
			if _, ok := capitalGroupIDs[info.GroupID]; ok {
				incrementType(result.internalCapital, info.GroupName, info.Name, item.TypeID)
				result.Stats.CapitalCount++
			}
		case hasID(structureCategoryIDs, info.CategoryID):
			incrementType(result.internalStructure, info.GroupName, info.Name, item.TypeID)
			result.Stats.StructureCount++
		default:
			incrementType(result.internalMisc, info.GroupName, info.Name, item.TypeID)
			result.Stats.MiscCount++
		}
	}

	result.ShipTypes = sortedTypeMap(result.internalShip)
	result.CapitalTypes = sortedTypeMap(result.internalCapital)
	result.StructureTypes = sortedTypeMap(result.internalStructure)
	result.MiscTypes = sortedTypeMap(result.internalMisc)
	result.clearInternal()

	if len(systemCandidates) > 0 {
		systemName := mostCommon(systemCandidates)
		result.SystemInfo.Name = systemName

		systemID, err := sdeClient.SystemIDByName(ctx, systemName, language)
		if err == nil && systemID != 0 {
			result.SystemInfo.ID = &systemID
			if info, err := sdeClient.SystemInfoByID(ctx, systemID); err == nil {
				result.SystemInfo.Region = &info.RegionID
				result.SystemInfo.RegionName = info.RegionName
				result.SystemInfo.Constellation = &info.ConstellationID
				result.SystemInfo.ConstellationName = info.ConstellationName
				result.SystemInfo.Security = &info.Security
			}
		}
	}

	return result, nil
}

type internalTypeEntry struct {
	ID    int64
	Count int
}

func incrementType(target map[string]map[string]internalTypeEntry, group string, name string, typeID int64) {
	if group == "" {
		group = "Unknown"
	}
	if name == "" {
		name = "Unknown"
	}
	if _, ok := target[group]; !ok {
		target[group] = map[string]internalTypeEntry{}
	}
	entry := target[group][name]
	entry.Count++
	entry.ID = typeID
	target[group][name] = entry
}

func sortedTypeMap(source map[string]map[string]internalTypeEntry) map[string][]TypeCount {
	groupNames := make([]string, 0, len(source))
	for group := range source {
		groupNames = append(groupNames, group)
	}
	sort.Slice(groupNames, func(i, j int) bool {
		return groupTotal(source[groupNames[i]]) > groupTotal(source[groupNames[j]])
	})

	result := make(map[string][]TypeCount, len(source))
	for _, group := range groupNames {
		names := make([]string, 0, len(source[group]))
		for name := range source[group] {
			names = append(names, name)
		}
		sort.Slice(names, func(i, j int) bool {
			return source[group][names[i]].Count > source[group][names[j]].Count
		})

		counts := make([]TypeCount, 0, len(names))
		for _, name := range names {
			entry := source[group][name]
			counts = append(counts, TypeCount{ID: entry.ID, Name: name, Count: entry.Count})
		}
		result[group] = counts
	}

	return result
}

func groupTotal(types map[string]internalTypeEntry) int {
	total := 0
	for _, entry := range types {
		total += entry.Count
	}
	return total
}

func mostCommon(values map[string]int) string {
	bestName := ""
	bestCount := -1
	for name, count := range values {
		if count > bestCount {
			bestName = name
			bestCount = count
		}
	}
	return bestName
}

func hasID(values map[int64]struct{}, id int64) bool {
	_, ok := values[id]
	return ok
}

type ShipItem struct {
	TypeID   int64   `json:"type_id"`
	Name     string  `json:"name"`
	TypeName *string `json:"type_name"`
	Distance *string `json:"distance"`
}

type ShipResult struct {
	ShipTypes       map[string][]TypeCount `json:"ship_types"`
	CapitalTypes    map[string][]TypeCount `json:"capital_types"`
	StructureTypes  map[string][]TypeCount `json:"structure_types"`
	MiscTypes       map[string][]TypeCount `json:"misc_types"`
	CapitalGroupIDs []int64                `json:"capital_group_ids"`
	Stats           ShipStats              `json:"stats"`
	SystemInfo      ShipSystemInfo         `json:"system_info"`
	FilterDistance  bool                   `json:"filter_distance"`

	internalShip      map[string]map[string]internalTypeEntry
	internalCapital   map[string]map[string]internalTypeEntry
	internalStructure map[string]map[string]internalTypeEntry
	internalMisc      map[string]map[string]internalTypeEntry
}

func (r *ShipResult) clearInternal() {
	r.internalShip = nil
	r.internalCapital = nil
	r.internalStructure = nil
	r.internalMisc = nil
}

type TypeCount struct {
	ID    int64  `json:"id,omitempty"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type ShipStats struct {
	TotalCount     int `json:"total_count"`
	ShipCount      int `json:"ship_count"`
	CapitalCount   int `json:"capital_count"`
	StructureCount int `json:"structure_count"`
	MiscCount      int `json:"misc_count"`
}

type ShipSystemInfo struct {
	Name              string   `json:"name"`
	ID                *int64   `json:"id"`
	Region            *int64   `json:"region"`
	RegionName        string   `json:"regionName"`
	Constellation     *int64   `json:"constellation"`
	ConstellationName string   `json:"constellationName"`
	Security          *float64 `json:"security"`
}
