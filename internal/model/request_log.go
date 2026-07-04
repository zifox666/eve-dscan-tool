package model

import "time"

type RequestLog struct {
	ID            uint64    `gorm:"primaryKey"`
	ClientIP      string    `gorm:"type:inet;index;not null"`
	RequestPath   string    `gorm:"type:text;not null"`
	RequestMethod string    `gorm:"size:10;not null"`
	ProcessTimeMS int64     `gorm:"not null"`
	StatusCode    int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null;autoCreateTime"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}
