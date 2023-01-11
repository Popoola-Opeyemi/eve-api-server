package operations

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Residency struct {
	ID             string    `json:"id,omitempty"`
	SiteID         string    `json:"site_id,omitempty"`
	UnitID         string    `json:"unit_id"`
	PreviousUnitID string    `json:"previous_unit_id"`
	ActiveStatus   int       `json:"active_status"`
	DateStart      time.Time `json:"date_start"`
}

// Create ...
func (s *Residency) Create(db *gorm.DB) error {
	return db.Create(s).Error
}

func GetAllResidency(db *gorm.DB) ([]Residency, error) {
	model := []Residency{}
	err := db.Find(&model).Error

	if err != nil {
		return model, err
	}
	return model, nil
}

func CreateResidency(oldeve, neweve *gorm.DB, logger *zap.SugaredLogger) {

	residentList, err := GetAllResident(oldeve)

	if err != nil {
		logger.Debug("error => ", err)
	}

	for _, resident := range residentList {

		residency := Residency{
			ID:           xid.New().String(),
			SiteID:       resident.SiteID,
			UnitID:       resident.UnitID,
			ActiveStatus: resident.Status,
			DateStart:    time.Now(),
		}
		err := residency.Create(neweve)
		if err != nil {
			logger.Debug("error => ", err)
		}
	}
}
