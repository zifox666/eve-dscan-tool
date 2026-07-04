package model

import (
	"time"

	"gorm.io/datatypes"
)

type LocalDScan struct {
	ID            uint64         `gorm:"primaryKey"`
	ShortID       string         `gorm:"size:16;uniqueIndex;not null"`
	RawData       string         `gorm:"type:text;not null"`
	ProcessedData datatypes.JSON `gorm:"type:jsonb;not null"`
	ClientIP      string         `gorm:"type:inet;not null"`
	ViewCount     int64          `gorm:"not null;default:0"`
	CreatedAt     time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"not null;autoUpdateTime"`
}

func (LocalDScan) TableName() string {
	return "local_dscans"
}

type ShipDScan struct {
	ID            uint64         `gorm:"primaryKey"`
	ShortID       string         `gorm:"size:16;uniqueIndex;not null"`
	RawData       string         `gorm:"type:text;not null"`
	ProcessedData datatypes.JSON `gorm:"type:jsonb;not null"`
	ClientIP      string         `gorm:"type:inet;not null"`
	ViewCount     int64          `gorm:"not null;default:0"`
	CreatedAt     time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"not null;autoUpdateTime"`
}

func (ShipDScan) TableName() string {
	return "ship_dscans"
}
