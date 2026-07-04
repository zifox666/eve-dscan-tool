package dscan

import (
	"context"
	"testing"

	"github.com/zifox666/eve-dscan-tool/internal/service/sde"
)

func TestAnalyzeShip(t *testing.T) {
	client := fakeSDEClient{
		types: map[int64]sde.TypeInfo{
			123: {TypeID: 123, GroupID: 25, CategoryID: 6, Name: "Caracal", GroupName: "Cruiser"},
			456: {TypeID: 456, GroupID: 30, CategoryID: 6, Name: "Revelation", GroupName: "Dreadnought"},
			789: {TypeID: 789, GroupID: 0, CategoryID: 65, Name: "Astrahus", GroupName: "Citadel"},
		},
		systemIDs: map[string]int64{"Jita": 30000142},
		systems: map[int64]sde.SystemInfo{
			30000142: {ID: 30000142, Name: "Jita", RegionID: 10000002, RegionName: "The Forge", ConstellationID: 20000020, ConstellationName: "Kimotoro", Security: 0.9},
		},
	}

	raw := "123\tShip One\tCaracal\t10 km\n456\tShip Two\tRevelation\t-\n789\tJita - Structure\tAstrahus\t1 AU"
	result, err := AnalyzeShip(context.Background(), client, raw, "en", true)
	if err != nil {
		t.Fatalf("analyze ship: %v", err)
	}

	if result.Stats.TotalCount != 3 {
		t.Fatalf("expected total 3, got %d", result.Stats.TotalCount)
	}
	if result.Stats.ShipCount != 1 {
		t.Fatalf("expected filtered ship count 1, got %d", result.Stats.ShipCount)
	}
	if result.Stats.CapitalCount != 0 {
		t.Fatalf("expected filtered capital count 0, got %d", result.Stats.CapitalCount)
	}
	if result.Stats.StructureCount != 1 {
		t.Fatalf("expected structure count 1, got %d", result.Stats.StructureCount)
	}
	if result.SystemInfo.ID == nil || *result.SystemInfo.ID != 30000142 {
		t.Fatalf("expected system id 30000142, got %#v", result.SystemInfo.ID)
	}
}

type fakeSDEClient struct {
	types     map[int64]sde.TypeInfo
	systemIDs map[string]int64
	systems   map[int64]sde.SystemInfo
}

func (f fakeSDEClient) TypeInfo(ctx context.Context, typeID int64, language string) (*sde.TypeInfo, error) {
	value, ok := f.types[typeID]
	if !ok {
		return nil, sde.ErrNotFound
	}
	return &value, nil
}

func (f fakeSDEClient) SystemIDByName(ctx context.Context, name string, language string) (int64, error) {
	value, ok := f.systemIDs[name]
	if !ok {
		return 0, sde.ErrNotFound
	}
	return value, nil
}

func (f fakeSDEClient) SystemInfoByID(ctx context.Context, systemID int64) (*sde.SystemInfo, error) {
	value, ok := f.systems[systemID]
	if !ok {
		return nil, sde.ErrNotFound
	}
	return &value, nil
}
