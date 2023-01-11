package operations

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Street struct {
	ID     string `json:"id"`
	SiteID string `json:"site_id"`
	Name   string `json:"name"`
}

func GetAllStreet(db *gorm.DB) ([]Street, error) {
	model := []Street{}
	err := db.Find(&model).Error

	if err != nil {
		return model, err
	}
	return model, nil
}

// Create ...
func (s *Street) Create(db *gorm.DB) error {

	return db.Create(s).Error
}

func MigrateStreet(oldeve, neweve *gorm.DB, logger *zap.SugaredLogger) error {
	res, err := GetAllStreet(oldeve)
	if err != nil {
		// logger.Debug("error")
		return err
	}

	// logger.Debug(res)

	for _, street := range res {
		if err := street.Create(neweve); err != nil {
			// logger.Debug("error", err)
			return err
		}
	}
	return nil
}
