package operations

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Resident struct {
	ID           string          `json:"id,omitempty"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Email        string          `json:"email"`
	Password     string          `json:"password"`
	Phone        string          `json:"phone"`
	Attr         json.RawMessage `json:"attr"`
	Type         int             `json:"type"`
	Status       int             `json:"status"`
	PrimaryID    string          `json:"primary_id"`
	ResidencyID  string          `json:"residency_id"`
	Unit         string          `json:"unit"`
	UnitID       string          `json:"unit_id"`
	SiteID       string          `json:"site_id"`
	ActiveStatus int             `json:"active_status"`
	DateStart    time.Time       `json:"date_start"`
}

type NewResident struct {
	ID          string          `json:"id,omitempty"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	Password    string          `json:"password"`
	Phone       string          `json:"phone"`
	Attr        json.RawMessage `json:"attr"`
	Type        int             `json:"type"`
	Status      int             `json:"status"`
	PrimaryID   string          `json:"primary_id"`
	ResidencyID string          `json:"residency_id"`
}

// Create ...
func (s *NewResident) Create(db *gorm.DB) error {
	return db.Table("resident").Create(s).Error
}

func GetAllResident(db *gorm.DB) ([]Resident, error) {
	model := []Resident{}
	err := db.Find(&model, "type = ?", 1).Error

	if err != nil {
		return model, err
	}
	return model, nil
}

// GetByID ...
func (s *Resident) GetByID(db *gorm.DB, id int64) error {
	return db.Take(s, "id = ?", id).Error
}

// GetByID ...
func (s *Resident) GetByUnitID(db *gorm.DB) error {
	return db.Take(s, "site_id = ? and unit_id = ?", s.SiteID, s.UnitID).Error
}

func CreateResident(oldeve, neweve *gorm.DB, logger *zap.SugaredLogger) {

	residencyList, err := GetAllResidency(neweve)

	if err != nil {
		logger.Debug("error => ", err)
	}

	for _, residency := range residencyList {

		r := Resident{
			UnitID: residency.UnitID,
			SiteID: residency.SiteID,
		}

		err = r.GetByUnitID(oldeve)
		if err != nil {
			logger.Debug("error => ", err)
		}

		if r.ID != "" {
			if r.Type == 1 {
				newResident := NewResident{
					ID:          xid.New().String(),
					FirstName:   r.FirstName,
					LastName:    r.LastName,
					Email:       r.Email,
					Password:    r.Password,
					Phone:       r.Phone,
					Attr:        r.Attr,
					Type:        r.Type,
					Status:      r.Status,
					PrimaryID:   r.PrimaryID,
					ResidencyID: residency.ID,
				}

				err := newResident.Create(neweve)
				if err != nil {
					logger.Debug("error => ", err)
				}

			}
		}
	}
}
