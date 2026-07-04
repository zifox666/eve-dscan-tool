package dscan

import (
	"context"
	"errors"

	"github.com/zifox666/eve-dscan-tool/internal/model"
	"github.com/zifox666/eve-dscan-tool/internal/service/shortid"
	"github.com/zifox666/eve-dscan-tool/internal/store/postgres"
	"gorm.io/gorm"
)

const defaultCreateRetries = 5

type Service struct {
	repo            *postgres.DScanRepository
	shortLinkLength int
	createRetries   int
}

func NewService(repo *postgres.DScanRepository, shortLinkLength int) *Service {
	return &Service{
		repo:            repo,
		shortLinkLength: shortLinkLength,
		createRetries:   defaultCreateRetries,
	}
}

func (s *Service) Repository() *postgres.DScanRepository {
	return s.repo
}

func (s *Service) CreateLocal(ctx context.Context, input CreateInput) (*model.LocalDScan, error) {
	var lastErr error

	for i := 0; i < s.createRetries; i++ {
		generatedID, err := shortid.Generate(s.shortLinkLength)
		if err != nil {
			return nil, err
		}

		record, err := s.repo.CreateLocal(ctx, postgres.CreateDScanInput{
			ShortID:       generatedID,
			RawData:       input.RawData,
			ProcessedData: input.ProcessedData,
			ClientIP:      input.ClientIP,
		})
		if err == nil {
			return record, nil
		}
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

func (s *Service) CreateShip(ctx context.Context, input CreateInput) (*model.ShipDScan, error) {
	var lastErr error

	for i := 0; i < s.createRetries; i++ {
		generatedID, err := shortid.Generate(s.shortLinkLength)
		if err != nil {
			return nil, err
		}

		record, err := s.repo.CreateShip(ctx, postgres.CreateDScanInput{
			ShortID:       generatedID,
			RawData:       input.RawData,
			ProcessedData: input.ProcessedData,
			ClientIP:      input.ClientIP,
		})
		if err == nil {
			return record, nil
		}
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

type CreateInput struct {
	RawData       string
	ProcessedData []byte
	ClientIP      string
}
