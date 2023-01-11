package operations

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Content struct {
	ID     string          `json:"id,omitempty"`
	SiteID string          `json:"site_id"`
	Data   json.RawMessage `json:"data"`
}

func GetAllContent(db *gorm.DB) ([]Content, error) {
	model := []Content{}
	err := db.Find(&model).Error

	if err != nil {
		return model, err
	}
	return model, nil
}

// Create ...
func (s *Content) Create(db *gorm.DB) error {

	return db.Create(s).Error
}

func MigrateContent(oldeve, neweve *gorm.DB, logger *zap.SugaredLogger) error {
	res, err := GetAllContent(oldeve)
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
