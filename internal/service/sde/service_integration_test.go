package sde_test

import (
	"os"
	"testing"

	"github.com/zifox666/eve-dscan-tool/internal/config"
	"github.com/zifox666/eve-dscan-tool/internal/service/sde"
	"github.com/zifox666/eve-dscan-tool/internal/store/postgres"
)

func TestIntegrationSDEQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if run := os.Getenv("RUN_SDE_INTEGRATION"); run != "1" {
		t.Skip("set RUN_SDE_INTEGRATION=1 to run")
	}

	cfg := config.Load()
	db, err := postgres.OpenSDE(cfg)
	if err != nil {
		t.Fatalf("open sde: %v", err)
	}
	sqlDB, err := postgres.SQLDB(db)
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	defer sqlDB.Close()

	service := sde.NewService(db)

	typeInfo, err := service.TypeInfo(t.Context(), 34, "en")
	if err != nil {
		t.Fatalf("type info: %v", err)
	}
	if typeInfo.TypeID != 34 {
		t.Fatalf("expected type id 34, got %d", typeInfo.TypeID)
	}

	systemID, err := service.SystemIDByName(t.Context(), "Jita", "en")
	if err != nil {
		t.Fatalf("system id: %v", err)
	}
	if systemID != 30000142 {
		t.Fatalf("expected Jita system id 30000142, got %d", systemID)
	}
}
