package operations

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type User struct {
	ID        string          `json:"id"`
	SiteID    string          `json:"site_id"`
	Status    int             `json:"status"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	Phone     string          `json:"phone"`
	Attr      json.RawMessage `json:"attr"`
	Role      int             `json:"role"`
	// 1: service, 2: security, 3: resident, 4: official, 5: admin, 6: platform
	Type int `json:"type"`
	// indicates that this is the default user and cannot be deleted
	IsSiteUser bool `json:"is_site_user"`
}

func GetAllUsers(db *gorm.DB) ([]User, error) {
	model := []User{}
	err := db.Find(&model).Error

	if err != nil {
		return model, err
	}
	return model, nil
}

// Create ...
func (s *User) Create(db *gorm.DB) error {

	return db.Create(s).Error
}

func MigrateUser(oldeve, neweve *gorm.DB, logger *zap.SugaredLogger) error {
	res, err := GetAllUsers(oldeve)
	if err != nil {
		logger.Debug("error")
		return err
	}

	// logger.Debug(res)

	for _, users := range res {
		if err := users.Create(neweve); err != nil {
			logger.Debug("error", err)
			return err
		}
	}
	return nil
}

func CreateSupportAccount(db *gorm.DB) {
	UserList := []User{}
	err := db.Find(&UserList).Where("type = ?", 4).Error

	for _, user := range UserList {

	}
}
