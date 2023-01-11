package operations

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Site struct {
	ID             string          `json:"id"`
	Subdomain      string          `json:"subdomain"`
	Name           string          `json:"name"`
	Status         int             `json:"status"`
	DateRegistered time.Time       `json:"date_registered"`
	Attr           json.RawMessage `json:"attr"`
	Platform       bool            `json:"platform"`
}

func GetAll(db *gorm.DB) ([]Site, error) {
	model := []Site{}
	err := db.Find(&model).Error

	if err != nil {
		return model, err
	}
	return model, nil
}

// Update ...
func (s *Site) Update(db *gorm.DB) error {

	return db.Model(&s).Update(*s).Error
}

// Create ...
func (s *Site) Create(db *gorm.DB) error {

	return db.Create(s).Error
}

func MigrateSite(oldeve, neweve *gorm.DB, logger *zap.SugaredLogger) error {
	res, err := GetAll(oldeve)

	if err != nil {
		logger.Debug(err)
		return err
	}

	for _, site := range res {
		if err := site.Create(neweve); err != nil {
			return err

		}
	}
	logger.Debug(res)
	return nil
}
