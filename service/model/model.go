package model

import (
	"encoding/json"
	"eve/utils"
	"fmt"

	"github.com/shopspring/decimal"
)

// Status ...
type Status int

// MarshalJSON ...
func (s Status) MarshalJSON() ([]byte, error) {
	retv := fmt.Sprintf("%d", s)
	return []byte(retv), nil
}

// UnmarshalJSON ...
func (s *Status) UnmarshalJSON(val []byte) error {
	retv := utils.Atoi(string(val))
	*s = Status(retv)
	return nil
}

// Site ...
type Site struct {
	ID             string          `json:"id,omitempty"`
	Subdomain      string          `json:"subdomain"`
	Name           string          `json:"name"`
	Status         int             `json:"status" sql:",notnull"`
	SiteCode       string          `json:"site_code"`
	DateRegistered utils.DateTime  `json:"date_registered"`
	Attr           json.RawMessage `json:"attr"`
	Platform       bool            `json:"platform" sql:"-,notnull"`
}

// User ...
type User struct {
	ID             string          `json:"id,omitempty"`
	SiteID         string          `json:"site_id,omitempty"`
	Status         Status          `json:"status" sql:",notnull"`
	ActiveStatus   int             `json:"active_status" sql:"-"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          string          `json:"email"`
	Password       string          `json:"password"`
	Phone          string          `json:"phone"`
	Attr           json.RawMessage `json:"attr"`
	SupportAccount bool            `json:"support_account"`
	Role           int             `json:"role" sql:",notnull"`
	// 1: service, 2: security, 3: resident, 4: official, 5: admin, 6: platform
	Type int `json:"type" sql:",notnull"`

	// used to indicate that this user is a subtype of a main type eg secondary residents
	SubType int `json:"sub_type" sql:"-"`

	// indicates that this is the default user and cannot be deleted
	IsSiteUser bool `json:"is_site_user"`

	Address     string `json:"Address" sql:"-"`
	ResidencyID string `json:"residency_id" sql:"-"`

	// push notification token
	PushToken string `json:"json_token" sql:"-"`
}

// UserType ...
type UserType struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
}

// Street ...
type Street struct {
	ID     string `json:"id"`
	SiteID string `json:"site_id"`
	Name   string `json:"name"`
}

// UnitType ...
type UnitType struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Unit ...
type Unit struct {
	ID       string          `json:"id"`
	SiteID   string          `json:"site_id"`
	Type     int             `json:"type"`
	StreetID string          `json:"street_id"`
	Label    string          `json:"label" sql:",notnull"`
	Attr     json.RawMessage `json:"attr"`
}

// Resident ...
type Resident struct {
	ID           string          `json:"id,omitempty"`
	CanLogin     bool            `json:"can_login"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Email        string          `json:"email"`
	Password     string          `json:"password"`
	Phone        string          `json:"phone"`
	Attr         json.RawMessage `json:"attr"`
	Type         int             `json:"type"`
	Status       Status          `json:"status" sql:",notnull"`
	PrimaryID    string          `json:"primary_id" sql:",notnull"`
	ResidencyID  string          `json:"residency_id" sql:",notnull"`
	Unit         string          `json:"unit" sql:"-"`
	UnitID       string          `json:"unit_id" sql:"-"`
	SiteID       string          `json:"site_id" sql:"-"`
	PushToken    string          `json:"json_token" sql:"-"`
	ActiveStatus int             `json:"active_status" sql:"-"`
}

// Residency ...
type Residency struct {
	ID             string         `json:"id,omitempty"`
	SiteID         string         `json:"site_id,omitempty"`
	UnitID         string         `json:"unit_id"`
	PreviousUnitID string         `json:"previous_unit_id"`
	Type           int            `json:"type" sql:"-"`
	ActiveStatus   int            `json:"active_status" sql:",notnull"`
	DateStart      utils.DateTime `json:"date_start"`
	DateExit       utils.DateTime `json:"date_exit"`
}

// Due ...
type Due struct {
	ID          string          `json:"id"`
	SiteID      string          `json:"site_id,omitempty"`
	DateCreated utils.DateTime  `json:"date_created"`
	Name        string          `json:"name"`
	Description string          `json:"description" sql:",notnull"`
	Amount      decimal.Decimal `json:"amount" sql:",notnull"`
	Status      Status          `json:"status" sql:",notnull"`
	Attr        json.RawMessage `json:"attr"`
}

// Bill ...
type Bill struct {
	ID          string         `json:"id"`
	SiteID      string         `json:"site_id,omitempty"`
	UnitType    int            `json:"unit_type"`
	DateCreated utils.DateTime `json:"date_created"`
	Name        string         `json:"name"`
	Note        string         `json:"note" sql:",notnull"`
	// there can only be one active (status == 1) bill for each unit type
	Status int             `json:"status" sql:",notnull"`
	Total  decimal.Decimal `json:"total" sql:",notnull"`
	Items  json.RawMessage `json:"items" sql:"-"`
	Attr   json.RawMessage `json:"attr"`
}

// BillItem ...
type BillItem struct {
	ID     string          `json:"id"`
	SiteID string          `json:"site_id,omitempty"`
	BillID string          `json:"bill_id"`
	DueID  string          `json:"due_id"`
	Amount decimal.Decimal `json:"amount" sql:",notnull"`
}

// BillGenerate ...
type BillGenerate struct {
	ID          string         `json:"id"`
	SiteID      string         `json:"site_id"`
	BillID      string         `json:"bill_id"`
	DateCreated utils.DateTime `json:"date_created"`
	Month       int            `json:"month"`
	Year        int            `json:"year"`
	UserID      string         `json:"user_id"`
}

// InvoiceMaster is deprecated
type InvoiceMaster struct {
	ID          string          `json:"id"`
	SiteID      string          `json:"site_id"`
	Month       int             `json:"month"`
	Year        int             `json:"year"`
	Bills       json.RawMessage `json:"bills"`
	DateCreated utils.DateTime  `json:"date_created"`
	Description string          `json:"description"`
	Attr        json.RawMessage `json:"attr"`
}

// Transaction ...
type Transaction struct {
	ID         string          `json:"id"`
	SiteID     string          `json:"site_id"`
	ResidentID string          `json:"resident_id"`
	Type       int             `json:"type"`
	DateTrx    utils.DateTime  `json:"date_trx"`
	InvoiceID  string          `json:"invoice_id"`
	PaymentID  string          `json:"payment_id"`
	DueID      string          `json:"due_id"`
	Amount     decimal.Decimal `json:"amount" sql:",notnull"`
}

// NoticeBoard ...
type NoticeBoard struct {
	ID          string          `json:"id"`
	SiteID      string          `json:"site_id,omitempty"`
	DateCreated utils.DateTime  `json:"date_created"`
	DateExpiry  utils.DateTime  `json:"date_expiry"`
	Title       string          `json:"title"`
	Message     string          `json:"message"`
	Attr        json.RawMessage `json:"attr"`
}

// GatePass ...
type GatePass struct {
	ID          string          `json:"id"`
	SiteID      string          `json:"site_id,omitempty"`
	ResidentID  string          `json:"resident_id"`
	DateCreated utils.DateTime  `json:"date_created"`
	Token       string          `json:"token"`
	PlateNumber string          `json:"plate_number"`
	Type        int             `json:"type" sql:",notnull"`
	Status      int             `json:"status" sql:",notnull"`
	Resident    string          `json:"resident" sql:"-"`
	Attr        json.RawMessage `json:"attr"`
}

// Visitor ...
type Visitor struct {
	ID               string          `json:"id"`
	SiteID           string          `json:"site_id"`
	DateCreated      utils.DateTime  `json:"date_created"`
	DateArrival      utils.DateTime  `json:"date_arrival"`
	Name             string          `json:"name"`
	VehicleNumber    string          `json:"vehicle_number"`
	ArrivalTime      string          `json:"arrival_time"`
	DepartureTime    string          `json:"departure_time"`
	RegistrationType int             `json:"registration_type"`
	Status           int             `json:"status"`
	ResidentID       string          `json:"resident_id"`
	UnitID           string          `json:"unit_id"`
	SecurityID       string          `json:"security_id"`
	Resident         string          `json:"resident" sql:"-"`
	Security         string          `json:"security" sql:"-"`
	Attr             json.RawMessage `json:"attr"`
}

// Invoice ...
type Invoice struct {
	ID            string          `json:"id"`
	SiteID        string          `json:"site_id"`
	ResidentID    string          `json:"resident_id"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	Address       string          `json:"address"`
	Month         int             `json:"month"`
	Year          int             `json:"year"`
	DateCreated   utils.DateTime  `json:"date_created"`
	BillID        string          `json:"bill_id"`
	UnitType      int             `json:"unit_type"`
	Description   string          `json:"description"`
	Amount        decimal.Decimal `json:"amount"`
	Dues          json.RawMessage `json:"dues"`
	InvoiceNumber int64           `json:"invoice_number"`
	UnitTypeLabel string          `json:"unit_type_label" sql:"-"`
	Resident      string          `json:"resident" sql:"-"`
}

// Payment ...
type Payment struct {
	ID         string          `json:"id"`
	SiteID     string          `json:"site_id"`
	ResidentID string          `json:"resident_id"`
	DateTrx    utils.DateTime  `json:"date_trx"`
	Narration  string          `json:"narration" sql:",notnull"`
	Amount     decimal.Decimal `json:"amount"`
	Dues       json.RawMessage `json:"dues"`
	PayMode    int             `json:"pay_mode"`
	Attr       json.RawMessage `json:"attr"`
	Metadata   json.RawMessage `json:"metadata"`
}

type PaymentPending struct {
	ID           string          `json:"id"`
	SiteID       string          `json:"site_id"`
	ResidentID   string          `json:"resident_id"`
	ResidentName string          `json:"resident_name"`
	DateTrx      utils.DateTime  `json:"date_trx"`
	Narration    string          `json:"narration" sql:",notnull"`
	Amount       decimal.Decimal `json:"amount"`
	PayMode      int             `json:"pay_mode"`
	Dues         json.RawMessage `json:"dues"`
	Attr         json.RawMessage `json:"attr"`
}

// LoginForm ...
type LoginForm struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsResident bool   `json:"is_resident"`
	IsMobile   bool   `json:"is_mobile"`
}

// TaskQueue ...
type TaskQueue struct {
	ID          int             `json:"id"`
	SiteID      string          `json:"site_id"`
	Type        int             `json:"type"`
	DateCreated utils.DateTime  `json:"date_created"`
	Data        json.RawMessage `json:"data"`
	Status      int             `json:"status"`
}

// Registration ...
type Registration struct {
	ID          string          `json:"id"`
	DateCreated utils.DateTime  `json:"date_created"`
	Data        json.RawMessage `json:"data"`
	Status      int             `json:"status"`
}

// Content ...
type Content struct {
	ID     string          `json:"id,omitempty"`
	SiteID string          `json:"site_id"`
	Data   json.RawMessage `json:"data"`
}

// NewResidentRegistrations ...
type NewResidentRegistrations struct {
	ID             string          `json:"id,omitempty"`
	SiteID         string          `json:"site_id"`
	Name           string          `json:"name"`
	Attr           json.RawMessage `json:"attr"`
	SiteName       string          `json:"site_name"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          string          `json:"email"`
	Address        string          `json:"address"`
	Phone          string          `json:"phone"`
	DateRegistered utils.DateTime  `json:"date_registered"`
}

type PaymentProviders struct {
	ID    string      `json:"id"`
	Value PayProvider `json:"value"`
	Name  string      `json:"name"`
}

type InitializePaystackRequest struct {
	Email  string `json:"email"`
	Amount string `json:"amount"`
}

type PayResponse struct {
	AuthorizationUrl string `json:"authorization_url"`
	AccessCode       string `json:"access_code"`
	Reference        string `json:"reference"`
}

type InitializePaystackResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    PayResponse `json:"data"`
}

type PaymentLog struct {
	ID           string          `json:"id"`
	SiteID       string          `json:"site_id"`
	Operation    string          `json:"operation"`
	Amount       decimal.Decimal `json:"amount"`
	InitiatedBY  UserDetails     `json:"initiated_by"`
	InitiatedFor UserDetails     `json:"initiated_for"`
	Narration    string          `json:"narration"`
	Date         utils.DateTime  `json:"date"`
}

type ResidentDue struct {
	DueID    string          `json:"due_id"`
	Due      string          `json:"due"`
	Invoices decimal.Decimal `json:"invoices"`
	Payments decimal.Decimal `json:"payments"`
	Balance  decimal.Decimal `json:"balance"`
}

type SubResident struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Type      int    `json:"type"`
}

type ResidentDashboard struct {
	SiteID         string          `json:"site_id"`
	ResidentID     string          `json:"resident_id"`
	Name           string          `json:"name"`
	ResidentDues   ResidentDues    `json:"resident_dues"`
	GatePassList   GatePassList    `json:"gate_pass_list"`
	SubResidents   SubResidents    `json:"sub_residents"`
	Notification   NoticeBoards    `json:"notification"`
	InDebt         bool            `json:"in_debt"`
	AccountBalance decimal.Decimal `json:"account_balance"`
}

type ResidentAlerts struct {
	ID            string          `json:"id"`
	SiteID        string          `json:"site_id"`
	ResidentID    string          `json:"resident_id"`
	Name          string          `json:"name"`
	PhoneNumber   string          `json:"phone_number"`
	Address       string          `json:"Address"`
	Status        ResidentAlert   `json:"status"`
	Attr          json.RawMessage `json:"attr"`
	TimeLogged    utils.DateTime  `json:"time_logged"`
	TimeResponded utils.DateTime  `json:"time_responded"`
}

type (
	ResidentDues struct {
		List  []ResidentDue `json:"list"`
		Count int           `json:"count"`
	}

	GatePassList struct {
		List  []GatePass `json:"list"`
		Count int        `json:"count"`
	}

	SubResidents struct {
		List  []SubResident `json:"list"`
		Count int           `json:"count"`
	}

	NoticeBoards struct {
		List  []NoticeBoard `json:"list"`
		Count int           `json:"count"`
	}
)
