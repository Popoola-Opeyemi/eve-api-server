package operations

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Unit struct {
	ID       string          `json:"id"`
	SiteID   string          `json:"site_id"`
	Type     int             `json:"type"`
	StreetID string          `json:"street_id"`
	Label    string          `json:"label" sql:",notnull"`
	Attr     json.RawMessage `json:"attr"`
}

func GetAllUnit(db *gorm.DB) ([]Unit, error) {
	model := []Unit{}
	err := db.Find(&model).Error

	if err != nil {
		return model, err
	}
	return model, nil
}

// Create ...
func (s *Unit) Create(db *gorm.DB) error {

	return db.Create(s).Error
}

func MigrateUnit(oldeve, neweve *gorm.DB, logger *zap.SugaredLogger) error {
	res, err := GetAllUnit(oldeve)
	if err != nil {
		logger.Debug("error")
		return err
	}

	// logger.Debug(res)

	for _, unit := range res {
		if err := unit.Create(neweve); err != nil {
			logger.Debug("error", err)
			return err
		}
	}
	return nil
}
