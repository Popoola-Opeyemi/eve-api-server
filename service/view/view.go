package view

import (
	"encoding/json"
	"eve/utils"

	"github.com/shopspring/decimal"
)

// AssociationView ...
type AssociationView struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Subdomain      string          `json:"subdomain"`
	DateRegistered utils.DateTime  `json:"date_registered"`
	Status         int             `json:"status"`
	SiteCode       string          `json:"site_code"`
	Attr           json.RawMessage `json:"attr"`

	AdminID        string `json:"admin_id"`
	AdminFirstName string `json:"admin_first_name"`
	AdminLastName  string `json:"admin_last_name"`
	AdminEmail     string `json:"admin_email"`
	AdminPassword  string `json:"admin_password" sql:"-"`
	AdminPhone     string `json:"admin_phone"`
	AdminStatus    int    `json:"admin_status"`

	StatsPaidResidents   int `json:"stats_paid_residents"`
	StatsIndebtResidents int `json:"stats_indebt_residents"`

	StatsActiveResidents    int `json:"stats_active_residents"`
	StatsPrimaryResidents   int `json:"stats_primary_residents"`
	StatsSecondaryResidents int `json:"stats_secondary_residents"`
	StatsResidents          int `json:"stats_residents"`

	UnitsOccupied int `json:"units_occupied"`
	StatsUsers    int `json:"stats_users"`
	StatsUnits    int `json:"stats_units"`
	StatsStreets  int `json:"stats_streets"`

	RegistrationID string `json:"registration_id" sql:"-"`
}

// AssociationList ...
type AssociationList struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Subdomain      string         `json:"subdomain"`
	DateRegistered utils.DateTime `json:"date_registered"`
	Status         int            `json:"status"`
}

// UserView ...
type UserView struct {
	tableName  struct{}        `sql:"user"`
	ID         string          `json:"id,omitempty"`
	Status     int             `json:"status" sql:",notnull"`
	FirstName  string          `json:"first_name"`
	LastName   string          `json:"last_name"`
	Email      string          `json:"email"`
	Phone      string          `json:"phone"`
	Attr       json.RawMessage `json:"attr,omitempty"`
	Role       int             `json:"role" sql:",notnull"`
	Type       int             `json:"type" sql:",notnull"`
	IsSiteUser bool            `json:"is_site_user"`
}

// UnitStreetView ...
type UnitStreetView struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// ResidentView ...
type ResidentView struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	UnitType    string          `json:"unit_type"`
	Attr        json.RawMessage `json:"attr"`
	Type        int             `json:"type"`
	UnitID      string          `json:"unit_id"`
	Unit        string          `json:"unit"`
	UnitName    string          `json:"unit_name"`
	StreetName  string          `json:"street_name"`
	Status      int             `json:"status"`
	DateStart   utils.DateTime  `json:"date_start"`
	ResidencyID string          `json:"residency_id"`
	DateExit    string          `json:"date_exit"`
}

// PlatformResidentView alias of ResidentView created as a work around the site_id requirements
// for manager access
type PlatformResidentView struct {
	tableName   struct{}        `sql:"resident_view"`
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Attr        json.RawMessage `json:"attr"`
	Type        int             `json:"type"`
	UnitID      string          `json:"unit_id"`
	Unit        string          `json:"unit"`
	Status      int             `json:"status"`
	Subdomain   string          `json:"subdomain"`
	Association string          `json:"association"`
}

// BillList ...
type BillList struct {
	ID           string          `json:"id"`
	SiteID       string          `json:"site_id,omitempty"`
	UnitType     int             `json:"unit_type"`
	DateCreated  utils.DateTime  `json:"date_created"`
	Name         string          `json:"name"`
	Note         string          `json:"note" sql:",notnull"`
	Status       int             `json:"status" sql:",notnull"`
	UnitTypeName string          `json:"unit_type_name"`
	Total        decimal.Decimal `json:"total" sql:",notnull"`
}

// BillDetailList ...
type BillDetailList struct {
	ID           string          `json:"id"`
	SiteID       string          `json:"site_id,omitempty"`
	UnitType     int             `json:"unit_type"`
	DateCreated  utils.DateTime  `json:"date_created"`
	Name         string          `json:"name"`
	Note         string          `json:"note" sql:",notnull"`
	Status       int             `json:"status" sql:",notnull"`
	UnitTypeName string          `json:"unit_type_name"`
	Total        decimal.Decimal `json:"total" sql:",notnull"`
	Items        json.RawMessage `json:"items"`
}

// BillItemList ...
type BillItemList struct {
	ID     string          `json:"id"`
	SiteID string          `json:"site_id,omitempty"`
	BillID string          `json:"bill_id"`
	DueID  string          `json:"due_id"`
	Due    string          `json:"due"`
	Amount decimal.Decimal `json:"amount" sql:",notnull"`
}

// UnitList ...
type UnitList struct {
	ID             string          `json:"id"`
	SiteID         string          `json:"site_id,omitempty"`
	Type           int             `json:"type"`
	StreetID       string          `json:"street_id"`
	Label          string          `json:"label"`
	Attr           json.RawMessage `json:"attr"`
	Street         string          `json:"street"`
	UnitType       string          `json:"unit_type"`
	OccupiedStatus string          `json:"occupied_status"`
}

// AvailableUnitsList ...
type AvailableUnitsList struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// ResidentID string `json:"resident_id"`
}

// GatePassList ...
type GatePassList struct {
	ID           string          `json:"id"`
	ResidentID   string          `json:"resident_id"`
	DateCreated  utils.DateTime  `json:"date_created"`
	Token        string          `json:"token"`
	Status       int             `json:"status"`
	Attr         json.RawMessage `json:"attr"`
	Visitor      string          `json:"visitor"`
	Resident     string          `json:"resident"`
	ResidentType int             `json:"resident_type"`
	ResidencyID  string          `json:"residency_id"`
	PlateNumber  string          `json:"plate_number"`
}

// ActiveNotice ...
type ActiveNotice struct {
	ID          string          `json:"id"`
	SiteID      string          `json:"site_id,omitempty"`
	DateCreated utils.DateTime  `json:"date_created"`
	DateExpiry  utils.DateTime  `json:"date_expiry"`
	Title       string          `json:"title"`
	Message     string          `json:"message"`
	Attr        json.RawMessage `json:"attr"`
}

// ExpiredNotice ...
type ExpiredNotice struct {
	ID          string          `json:"id"`
	SiteID      string          `json:"site_id,omitempty"`
	DateCreated utils.DateTime  `json:"date_created"`
	DateExpiry  utils.DateTime  `json:"date_expiry"`
	Title       string          `json:"title"`
	Message     string          `json:"message"`
	Attr        json.RawMessage `json:"attr"`
}

// VisitorList ...
type VisitorList struct {
	ID               string          `json:"id"`
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
	Resident         string          `json:"resident"`
	Security         string          `json:"security"`
	Attr             json.RawMessage `json:"attr"`
}

// ResidentList ...
type ResidentList struct {
	ID       string          `json:"id"`
	SiteID   string          `json:"site_id"`
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	Type     int             `json:"type"`
	UnitType int             `json:"unit_type"`
	Attr     json.RawMessage `json:"attr"`
}

// ResidentFamilyList ...
type ResidentFamilyList struct {
	ID        string          `json:"id"`
	SiteID    string          `json:"site_id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Phone     string          `json:"phone"`
	Type      int             `json:"type"`
	PrimaryID string          `json:"primary_id"`
	UnitID    string          `json:"unit_id"`
	Attr      json.RawMessage `json:"attr"`
}

// SecondaryResidentList ...
type SecondaryResidentList struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Phone     string          `json:"phone"`
	PrimaryID string          `json:"primary_id"`
	UnitID    string          `json:"unit_id"`
	Attr      json.RawMessage `json:"attr"`
}

// SecurityResidentList ...
type SecurityResidentList struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Unit   string `json:"unit"`
	Type   int    `json:"type"`
	InDebt bool   `json:"in_debt" sql:"-"`
}

// InvoiceMasterList ...
type InvoiceMasterList struct {
	ID          string          `json:"id"`
	SiteID      string          `json:"site_id"`
	Month       int             `json:"month"`
	Year        int             `json:"year"`
	Bills       json.RawMessage `json:"bills"`
	DateCreated utils.DateTime  `json:"date_created"`
	Description string          `json:"description"`
	Residents   int             `json:"residents"`
	Attr        json.RawMessage `json:"attr"`
}

// InvoiceList ...
type InvoiceList struct {
	ID            string          `json:"id"`
	SiteID        string          `json:"site_id"`
	ResidentID    string          `json:"resident_id"`
	Resident      string          `json:"resident"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	Address       string          `json:"address"`
	InvoiceNumber int64           `json:"invoice_number"`
	Month         int             `json:"month"`
	Year          int             `json:"year"`
	DateCreated   utils.DateTime  `json:"date_created"`
	BillID        string          `json:"bill_id"`
	UnitType      int             `json:"unit_type"`
	UnitTypeLabel string          `json:"unit_type_label"`
	Description   string          `json:"description"`
	Amount        decimal.Decimal `json:"amount"`
	Dues          json.RawMessage `json:"dues"`
}

// PaymentList ...
type PaymentList struct {
	ID            string          `json:"id"`
	SiteID        string          `json:"site_id"`
	ResidentID    string          `json:"resident_id"`
	Resident      string          `json:"resident"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	DateTrx       utils.DateTime  `json:"date_trx"`
	Narration     string          `json:"narration"`
	Amount        decimal.Decimal `json:"amount"`
	Dues          json.RawMessage `json:"dues"`
	UnitType      string          `json:"unit_type"`
	UnitTypeLabel string          `json:"unit_type_label"`
	PayMode       int             `json:"pay_mode"`
	Attr          json.RawMessage `json:"attr"`
	ReferenceID   string          `json:"reference_id"`
}

// ResidentDue is not a view, used as a mount point to query for data
type ResidentDue struct {
	DueID  string          `json:"due_id"`
	Due    string          `json:"due"`
	Amount decimal.Decimal `json:"amount"`
}

// ResidentBillingSummary ...
type ResidentBillingSummary struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	UnitID    string          `json:"unit_id"`
	Invoices  decimal.Decimal `json:"invoices"`
	Payments  decimal.Decimal `json:"payments"`
	Balance   decimal.Decimal `json:"balance"`
}

// InvoiceSummary ...
type InvoiceSummary struct {
	ID            string          `json:"id"`
	Month         int             `json:"month"`
	Year          int             `json:"year"`
	InvoiceNumber int             `json:"invoice_number"`
	ResidentID    string          `json:"resident_id"`
	Title         string          `json:"title"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	UnitID        string          `json:"unit_id"`
	Invoices      decimal.Decimal `json:"invoices"`
	Payments      decimal.Decimal `json:"payments"`
	Balance       decimal.Decimal `json:"balance"`
}

// ResidentAccountStatus ...
type ResidentAccountStatus struct {
	ID         string          `json:"id"`
	SiteID     string          `json:"site_id"`
	FirstName  string          `json:"first_name"`
	LastName   string          `json:"last_name"`
	Name       string          `json:"name"`
	Email      string          `json:"email"`
	Phone      string          `json:"phone"`
	UnitID     string          `json:"unit_id"`
	UnitTypeID int             `json:"unit_type_id"`
	UnitType   string          `json:"unit_type"`
	Attr       json.RawMessage `json:"attr"`
	Invoices   decimal.Decimal `json:"invoices"`
	Payments   decimal.Decimal `json:"payments"`
	Balance    decimal.Decimal `json:"balance"`
}

// ResidentDueStatus ...
type ResidentDueStatus struct {
	ID        string          `json:"id"`
	SiteID    string          `json:"site_id"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Name      string          `json:"name"`
	UnitID    string          `json:"unit_id"`
	UnitType  int             `json:"unit_type"`
	Attr      json.RawMessage `json:"attr"`
	DueID     string          `json:"due_id"`
	Due       string          `json:"due"`
	Invoices  decimal.Decimal `json:"invoices"`
	Payments  decimal.Decimal `json:"payments"`
	Balance   decimal.Decimal `json:"balance"`
}

// ServiceDueStatus ...
type ServiceDueStatus struct {
	tableName struct{}        `pg:",discard_unknown_columns"`
	ID        string          `json:"id"`
	Resident  string          `json:"resident"`
	Balance   decimal.Decimal `json:"balance"`
}

// AccountHistory ...
type AccountHistory struct {
	ResidentID    string          `json:"resident_id"`
	DocumentID    string          `json:"document_id"`
	DateTrx       utils.DateTime  `json:"date_trx"`
	Type          int             `json:"type"`
	Amount        decimal.Decimal `json:"amount"`
	Balance       decimal.Decimal `json:"balance"`
	InvoiceNumber string          `json:"invoice_number"`
}

// OldUnitsResidents ...
type OldUnitsResidents struct {
	ID             string          `json:"id,omitempty"`
	Name           string          `json:"name"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	Attr           json.RawMessage `json:"attr"`
	Type           int             `json:"type"`
	PreviousUnitID string          `json:"previous_unit_id"`
	Unit           string          `json:"unit"`
	DateStart      utils.DateTime  `json:"date_start"`
	DateExit       utils.DateTime  `json:"date_exit"`
}

type PaymentDetails struct {
	ID                    string          `json:"id"`
	SiteID                string          `json:"site_id,omitempty"`
	Amount                decimal.Decimal `json:"amount"`
	PayMode               string          `json:"pay_mode"`
	TransactionID         string          `json:"transaction_id"`
	ReferenceID           string          `json:"reference_id"`
	Name                  string          `json:"name"`
	Email                 string          `json:"email"`
	Provider              string          `json:"provider"`
	ProviderTransactionID string          `json:"provider_transaction_id"`
	ProviderReferenceID   string          `json:"provider_reference_id"`
	Dues                  json.RawMessage `json:"dues"`
	Date                  utils.DateTime  `json:"date"`
}

type ReportingResidents struct {
	ID             string          `json:"id"`
	SiteID         string          `json:"site_id"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          string          `json:"email"`
	Label          string          `json:"label"`
	Phone          string          `json:"phone"`
	Unit           string          `json:"unit"`
	Street         string          `json:"street"`
	UnitType       string          `json:"unit_type"`
	Attr           json.RawMessage `json:"attr,omitempty"`
	ActiveStatus   int             `json:"active_status" sql:",notnull"`
	OccupiedStatus string          `json:"occupied_status" sql:",notnull"`
	ResidentType   string          `json:"resident_type" sql:",notnull"`
}

type ReportingPayments struct {
	Amount        decimal.Decimal `json:"amount" sql:",notnull"`
	DateTrx       utils.DateTime  `json:"date_trx"`
	TransactionID string          `json:"transaction_id"`
	Name          string          `json:"name"`
	Dues          json.RawMessage `json:"dues"`
	PayMode       string          `json:"pay_mode"`
	SiteID        string          `json:"site_id"`
	Street        string          `json:"street"`
	UnitType      string          `json:"unit_type"`
	Attr          json.RawMessage `json:"attr,omitempty"`
	Label         string          `json:"label"`
}

type ReportingBill struct {
	SiteID      string          `json:"site_id"`
	UnitLabel   string          `json:"unit_label"`
	DateCreated utils.DateTime  `json:"date_created"`
	BillName    string          `json:"bill_name"`
	Total       decimal.Decimal `json:"total" sql:",notnull"`
}

type ReportingUnit struct {
	SiteID     string `json:"site_id"`
	UnitLabel  string `json:"unit_label"`
	Label      string `json:"label"`
	StreetName string `json:"street_name"`
	UnitNo     string `json:"unit_no"`
}

type ReportingInvoice struct {
	ID            string          `json:"id"`
	SiteID        string          `json:"site_id"`
	ResidentID    string          `json:"resident_id"`
	Resident      string          `json:"resident"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	Address       string          `json:"address"`
	InvoiceNumber string          `json:"invoice_number"`
	Month         int             `json:"month"`
	Year          int             `json:"year"`
	DateCreated   utils.DateTime  `json:"date_created"`
	BillID        string          `json:"bill_id"`
	UnitType      int             `json:"unit_type"`
	Description   string          `json:"description"`
	Amount        decimal.Decimal `json:"amount"`
	Dues          json.RawMessage `json:"dues"`
}
