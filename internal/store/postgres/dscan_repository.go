package postgres

import (
	"context"
	"errors"

	"github.com/zifox666/eve-dscan-tool/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("record not found")

type DScanRepository struct {
	db *gorm.DB
}

func NewDScanRepository(db *gorm.DB) *DScanRepository {
	return &DScanRepository{db: db}
}

func (r *DScanRepository) CreateLocal(ctx context.Context, input CreateDScanInput) (*model.LocalDScan, error) {
	record := &model.LocalDScan{
		ShortID:       input.ShortID,
		RawData:       input.RawData,
		ProcessedData: datatypes.JSON(input.ProcessedData),
		ClientIP:      input.ClientIP,
	}

	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

func (r *DScanRepository) CreateShip(ctx context.Context, input CreateDScanInput) (*model.ShipDScan, error) {
	record := &model.ShipDScan{
		ShortID:       input.ShortID,
		RawData:       input.RawData,
		ProcessedData: datatypes.JSON(input.ProcessedData),
		ClientIP:      input.ClientIP,
	}

	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

func (r *DScanRepository) FindLocalByShortID(ctx context.Context, shortID string) (*model.LocalDScan, error) {
	var record model.LocalDScan
	if err := r.db.WithContext(ctx).Where("short_id = ?", shortID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &record, nil
}

func (r *DScanRepository) FindShipByShortID(ctx context.Context, shortID string) (*model.ShipDScan, error) {
	var record model.ShipDScan
	if err := r.db.WithContext(ctx).Where("short_id = ?", shortID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &record, nil
}

func (r *DScanRepository) IncrementLocalViewCount(ctx context.Context, shortID string) (*model.LocalDScan, error) {
	var record model.LocalDScan
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("short_id = ?", shortID).First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		return tx.Model(&record).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
	})
	if err != nil {
		return nil, err
	}

	record.ViewCount++
	return &record, nil
}

func (r *DScanRepository) IncrementShipViewCount(ctx context.Context, shortID string) (*model.ShipDScan, error) {
	var record model.ShipDScan
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("short_id = ?", shortID).First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		return tx.Model(&record).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
	})
	if err != nil {
		return nil, err
	}

	record.ViewCount++
	return &record, nil
}

type CreateDScanInput struct {
	ShortID       string
	RawData       string
	ProcessedData []byte
	ClientIP      string
}
