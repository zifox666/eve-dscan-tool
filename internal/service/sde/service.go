package sde

import (
	"context"
	"errors"
	"math"
	"strings"

	"gorm.io/gorm"
)

const (
	LanguageZH = "zh"
	LanguageEN = "en"

	translationType     = 8
	translationGroup    = 7
	translationCategory = 6
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}

func (s *Service) TypeInfo(ctx context.Context, typeID int64, language string) (*TypeInfo, error) {
	language = normalizeLanguage(language)

	var row typeInfoRow
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			t."typeID" AS type_id,
			t."groupID" AS group_id,
			t."typeName" AS default_name,
			g."groupName" AS default_group_name,
			g."categoryID" AS category_id,
			c."groupName" AS default_category_name
		FROM "invTypes" t
		LEFT JOIN "invGroups" g ON t."groupID" = g."groupID"
		LEFT JOIN "invGroups" c ON g."categoryID" = c."groupID"
		WHERE t."typeID" = ?
	`, typeID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.TypeID == 0 {
		return nil, ErrNotFound
	}

	info := &TypeInfo{
		TypeID:       row.TypeID,
		GroupID:      row.GroupID,
		CategoryID:   row.CategoryID,
		DefaultName:  row.DefaultName,
		Name:         row.DefaultName,
		GroupName:    row.DefaultGroupName,
		CategoryName: row.DefaultCategoryName,
	}

	if translated, err := s.translation(ctx, typeID, translationType, language); err == nil && translated != "" {
		info.Name = translated
	}
	if translated, err := s.translation(ctx, row.GroupID, translationGroup, language); err == nil && translated != "" {
		info.GroupName = translated
	}
	if translated, err := s.translation(ctx, row.CategoryID, translationCategory, language); err == nil && translated != "" {
		info.CategoryName = translated
	}

	return info, nil
}

func (s *Service) TypeInfos(ctx context.Context, typeIDs []int64, language string) (map[int64]TypeInfo, error) {
	result := make(map[int64]TypeInfo)
	for _, typeID := range uniqueInt64s(typeIDs) {
		info, err := s.TypeInfo(ctx, typeID, language)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}
			return result, err
		}
		result[typeID] = *info
	}

	return result, nil
}

func (s *Service) SystemIDByName(ctx context.Context, name string, language string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, ErrNotFound
	}

	var direct struct {
		ItemID int64 `gorm:"column:item_id"`
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT "itemID" AS item_id
		FROM "mapDenormalize"
		WHERE "itemName" = ?
		LIMIT 1
	`, name).Scan(&direct).Error
	if err != nil {
		return 0, err
	}
	if direct.ItemID != 0 {
		return direct.ItemID, nil
	}

	language = normalizeLanguage(language)
	var translated struct {
		ItemID int64 `gorm:"column:item_id"`
	}
	err = s.db.WithContext(ctx).Raw(`
		SELECT d."itemID" AS item_id
		FROM "mapDenormalize" d
		INNER JOIN "trnTranslations" tr ON tr."keyID" = d."itemID"
		WHERE tr."languageID" = ? AND tr."text" = ?
		LIMIT 1
	`, language, name).Scan(&translated).Error
	if err != nil {
		return 0, err
	}
	if translated.ItemID == 0 {
		return 0, ErrNotFound
	}

	return translated.ItemID, nil
}

func (s *Service) SystemInfoByID(ctx context.Context, systemID int64) (*SystemInfo, error) {
	var row systemInfoRow
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			d."itemID" AS item_id,
			d."constellationID" AS constellation_id,
			d."regionID" AS region_id,
			d."itemName" AS item_name,
			d."security" AS security,
			co."constellationName" AS constellation_name,
			r."regionName" AS region_name
		FROM "mapDenormalize" d
		LEFT JOIN "mapConstellations" co ON d."constellationID" = co."constellationID"
		LEFT JOIN "mapRegions" r ON d."regionID" = r."regionID"
		WHERE d."itemID" = ?
		LIMIT 1
	`, systemID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ItemID == 0 {
		return nil, ErrNotFound
	}

	security := math.Floor(row.Security*100) / 100
	return &SystemInfo{
		ID:                row.ItemID,
		Name:              row.ItemName,
		ConstellationID:   row.ConstellationID,
		ConstellationName: row.ConstellationName,
		RegionID:          row.RegionID,
		RegionName:        row.RegionName,
		Security:          security,
	}, nil
}

func (s *Service) translation(ctx context.Context, keyID int64, tcID int, language string) (string, error) {
	var row struct {
		Text string `gorm:"column:text"`
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT "text"
		FROM "trnTranslations"
		WHERE "keyID" = ? AND "tcID" = ? AND "languageID" = ?
		LIMIT 1
	`, keyID, tcID, language).Scan(&row).Error
	if err != nil {
		return "", err
	}

	return row.Text, nil
}

func normalizeLanguage(language string) string {
	if language == LanguageEN {
		return LanguageEN
	}
	return LanguageZH
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

var ErrNotFound = errors.New("sde record not found")

type TypeInfo struct {
	TypeID       int64  `json:"typeID"`
	GroupID      int64  `json:"groupID"`
	CategoryID   int64  `json:"categoryID"`
	DefaultName  string `json:"default_name"`
	Name         string `json:"name"`
	GroupName    string `json:"group_name"`
	CategoryName string `json:"category_name"`
}

type SystemInfo struct {
	ID                int64   `json:"id"`
	Name              string  `json:"name"`
	ConstellationID   int64   `json:"constellation"`
	ConstellationName string  `json:"constellationName"`
	RegionID          int64   `json:"region"`
	RegionName        string  `json:"regionName"`
	Security          float64 `json:"security"`
}

type typeInfoRow struct {
	TypeID              int64  `gorm:"column:type_id"`
	GroupID             int64  `gorm:"column:group_id"`
	DefaultName         string `gorm:"column:default_name"`
	DefaultGroupName    string `gorm:"column:default_group_name"`
	CategoryID          int64  `gorm:"column:category_id"`
	DefaultCategoryName string `gorm:"column:default_category_name"`
}

type systemInfoRow struct {
	ItemID            int64   `gorm:"column:item_id"`
	ConstellationID   int64   `gorm:"column:constellation_id"`
	RegionID          int64   `gorm:"column:region_id"`
	ItemName          string  `gorm:"column:item_name"`
	Security          float64 `gorm:"column:security"`
	ConstellationName string  `gorm:"column:constellation_name"`
	RegionName        string  `gorm:"column:region_name"`
}
