package model

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PayProvider ...
type PayProvider int

type PaymentOPerations string

const (
	// IsDisabled ...
	IsDisabled Status = iota
	// IsEnabled ...
	IsEnabled
	// IsExResident ...
	IsExResident
)

type SessionName string

const (
	AdminSession    SessionName = "AdminSession"
	ResidentSession             = "ResidentSession"
)

// const (
// 	CashTransaction ManualPaymentOption = iota + 10
// 	BankTransaction
// )

// ResidencyEnded ...
const (
	ResidencyEnded int = iota
	ResidencyActive
)

// PrimaryResident && SecondaryResident
const (
	PrimaryResident int = iota + 1
	SecondaryResident
)

// 1: service, 2: security, 3: resident, 4: official, 5: admin, 6: platform
const (
	ServiceUser int = iota + 1
	SecurityUser
	ResidentUser
	OfficialUser
	AdminUser
	PlatformUser
	SupportUser
)

// Payment modes providers ...
const (
	ProviderManual PayProvider = iota
	ProviderPaystack
	ProviderRemita
	ProviderFlutterwave
)

type ResidentAlert int

const (
	AlertSent ResidentAlert = iota + 1
	AlertResponded
	AlertResolved
	AlertCompleted
)

// PaymentProviders ...
var paymentProviders = map[PayProvider]PayEntity{
	ProviderPaystack:    {Name: "paystack", Value: ProviderPaystack, URL: "https://api.paystack.co/transaction/verify"},
	ProviderRemita:      {Name: "remita", Value: ProviderRemita},
	ProviderFlutterwave: {Name: "flutterwave", Value: ProviderFlutterwave, URL: "https://api.flutterwave.com/v3/transactions"},
}

// payment types
const (
	ManualPayment int = iota
	OnlinePayment
	BankTransaction
)

const (
	PaymentDelete PaymentOPerations = "Delete"
	PaymentInsert                   = "Insert"
)

// PayEntity ...
type PayEntity struct {
	Name  string
	Value PayProvider
	URL   string
}

// PaymentProviderResponse ...
type PaymentProviderResponse struct {
	Status  bool            `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// PaystackResponse ...
type PaystackResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type UserDetails struct {
	UserID   string `json:"user_id"`
	UserType int    `json:"user_type"`
	Name     string `json:"name"`
}

// FlutterwaveResponse ...
type FlutterwaveResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// GetProvider ...
func GetProvider(id PayProvider) PayEntity {
	return paymentProviders[id]
}

// GetURL ...
func (p *PayEntity) GetURL(txRef string) (url string, err error) {

	if p.Value == ProviderPaystack {
		url = fmt.Sprintf("%s/%s", p.URL, txRef)
	}

	if p.Value == ProviderFlutterwave {
		url = fmt.Sprintf("%s/%s/verify", p.URL, txRef)
	}

	if url == "" {
		return url, errors.New("empty url")
	}

	return
}

func GetSession(usr int) string {
	switch usr {

	case AdminUser:
		return string(AdminSession)

	case ResidentUser:
		return string(ResidentSession)
	}
	return ""
}
