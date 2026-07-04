package postgres

import (
	"database/sql"
	"time"

	"github.com/zifox666/eve-dscan-tool/internal/config"
	"github.com/zifox666/eve-dscan-tool/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	return open(cfg.DatabaseURL, cfg.DatabaseMaxConns, cfg.DatabaseMinConns)
}

func OpenSDE(cfg config.Config) (*gorm.DB, error) {
	return open(cfg.SDEURL, cfg.DatabaseMaxConns, cfg.DatabaseMinConns)
}

func open(databaseURL string, maxConns int32, minConns int32) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Warn),
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(int(maxConns))
	sqlDB.SetMaxIdleConns(int(minConns))
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.LocalDScan{},
		&model.ShipDScan{},
		&model.RequestLog{},
	)
}

func SQLDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}
