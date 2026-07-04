package postgres

import (
	"context"

	"github.com/zifox666/eve-dscan-tool/internal/model"
	"gorm.io/gorm"
)

type RequestLogRepository struct {
	db *gorm.DB
}

func NewRequestLogRepository(db *gorm.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

func (r *RequestLogRepository) Create(ctx context.Context, input CreateRequestLogInput) error {
	record := &model.RequestLog{
		ClientIP:      input.ClientIP,
		RequestPath:   input.RequestPath,
		RequestMethod: input.RequestMethod,
		ProcessTimeMS: input.ProcessTimeMS,
		StatusCode:    input.StatusCode,
	}

	return r.db.WithContext(ctx).Create(record).Error
}

type CreateRequestLogInput struct {
	ClientIP      string
	RequestPath   string
	RequestMethod string
	ProcessTimeMS int64
	StatusCode    int
}
