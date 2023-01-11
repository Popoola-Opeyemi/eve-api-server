package form

import (
	"encoding/json"
	"eve/service/model"
	"eve/utils"

	"github.com/shopspring/decimal"
)

// ResidentProfileSave ...
type ResidentProfileSave struct {
	tableName struct{}        `sql:"resident"`
	ID        string          `json:"id,omitempty"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Email     string          `json:"email"`
	Phone     string          `json:"phone"`
	Attr      json.RawMessage `json:"attr"`
}

// PaymentForm ...
type PaymentForm struct {
	tableName  struct{}        `sql:"payment"`
	ID         string          `json:"id"`
	ResidentID string          `json:"resident_id"`
	DateTrx    utils.DateTime  `json:"date_trx"`
	Narration  string          `json:"narration"`
	Amount     decimal.Decimal `json:"amount"`
	Dues       json.RawMessage `json:"dues"`
	Provider   model.PayEntity `json:"provider"`
	PayMode    int             `json:"pay_mode"`
	Attr       json.RawMessage `json:"attr"`
}

// Registration ...
type Registration struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	AdminFirstName string `json:"admin_first_name"`
	AdminLastName  string `json:"admin_last_name"`
	AdminEmail     string `json:"admin_email"`
	AdminPhone     string `json:"admin_phone"`
	AdminPassword  string `json:"admin_password"`
	SpamTrap       string `json:"admin_access_level"`
}

// RegistrationForm ...
type RegistrationForm struct {
	SiteID    string          `json:"site_id"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Email     string          `json:"email"`
	Address   string          `json:"address"`
	Phone     string          `json:"phone"`
	UnitID    string          `json:"unit_id"`
	Attr      json.RawMessage `json:"attr"`
}

// AcceptNewRegistration ...
type AcceptNewRegistration struct {
	ID     string `json:"id"`
	UnitID string `json:"unit_id"`
}

// PaymentAttributes ...
type PaymentAttributes struct {
	Reference   string `json:"reference"`
	Transaction string `json:"transaction"`
	Provider    string `json:"provider"`
}

// PayProviderResponse ...
type PayProviderResponse struct {
	ProviderName  string            `json:"provider_name"`
	ProviderID    model.PayProvider `json:"provider_id"`
	TransactionID string            `json:"transaction_id"`
	ReferenceID   string            `json:"reference_id"`
}

type VerifyEmail struct {
	Email  string `json:"email"`
	SiteID string `json:"site_id"`
}

type SupportAccountForm struct {
	SiteID string `json:"site_id"`
}

type ResidentAlertForm struct {
	Status int             `json:"status"`
	Attr   json.RawMessage `json:"attr"`
}

type UpdatePushNotification struct {
	UserType int    `json:"user_type"`
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
}
