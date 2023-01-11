package shared

import (
	"bytes"
	"encoding/json"
	"eve/service/model"
	"eve/service/view"
	"eve/utils"
	"fmt"
	"reflect"
	"time"

	"github.com/CloudyKit/jet/v3"
	"github.com/go-pg/pg"
	"github.com/leekchan/accounting"
	"github.com/shopspring/decimal"
	"github.com/vanng822/go-premailer/premailer"
	"jaytaylor.com/html2text"
)

// Months ...
var Months = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

var templates = jet.NewHTMLSet("./templates")
var moneyFormat = accounting.Accounting{Symbol: "₦ ", Precision: 2}

// EMailMsg ...
type EMailMsg struct {
	To      string `json:"to,omitempty"`
	Subject string `json:"subject,omitempty"`
	HTML    string `json:"html,omitempty"`
	Text    string `json:"text,omitempty"`
}

func fmtMoney(args jet.Arguments) reflect.Value {
	args.RequireNumOfArguments("fmtMoney", 1, 1)
	val, ok := args.Get(0).Interface().(decimal.Decimal)
	if !ok {
		zero := "0.0"
		return reflect.ValueOf(zero)
	}

	return reflect.ValueOf(moneyFormat.FormatMoneyDecimal(val))
}

// HTMLToEMail converts html to email compatible html and text format
func HTMLToEMail(b []byte) (*EMailMsg, error) {
	prem, err := premailer.NewPremailerFromBytes(b, premailer.NewOptions())
	if err != nil {
		return nil, err
	}

	eml := EMailMsg{}

	// inline css
	eml.HTML, err = prem.Transform()
	if err != nil {
		return nil, err
	}

	// convert html to text
	eml.Text, err = html2text.FromString(eml.HTML, html2text.Options{PrettyTables: false})
	if err != nil {
		return nil, err
	}

	return &eml, nil
}

type invDetail struct {
	Due    string          `json:"due"`
	Amount decimal.Decimal `json:"amount"`
}

// MakeInvoice ...
func MakeInvoice(tx *pg.Tx, invID string) (*EMailMsg, error) {
	dbc := tx
	log := utils.Env.Log
	// decimal.DivisionPrecision = 2

	// get invoice record
	record := view.InvoiceList{}
	_, err := dbc.QueryOne(&record, "select * from invoice_list where id=?", invID)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	// get site record
	site := model.Site{}
	_, err = dbc.QueryOne(&site, "select * from site where id=?", record.SiteID)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	// unmarshal records.Dues to invDetail{}
	details := []invDetail{}
	if err := json.Unmarshal(record.Dues, &details); err != nil {
		log.Debug(err)
		return nil, err
	}
	// divide amount by 12
	twv, _ := decimal.NewFromString("12.00")
	for i := range details {
		details[i].Amount = details[i].Amount.DivRound(twv, 2)
	}

	templates.SetDevelopmentMode(true)
	templates.AddGlobalFunc("fmtMoney", fmtMoney)

	t, err := templates.GetTemplate("invoice.jet.html")
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	vars := make(jet.VarMap)
	vars.Set("invoice", record)
	vars.Set("invDate", record.DateCreated.Format("January 2 2006"))
	vars.Set("invNumber", fmt.Sprintf("%04d", record.InvoiceNumber))
	vars.Set("invDetails", details)
	vars.Set("association", site.Name)

	var w bytes.Buffer
	if err = t.Execute(&w, vars, nil); err != nil {
		log.Debug(err)
		return nil, err
	}

	eml, err := HTMLToEMail(w.Bytes())
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	eml.Subject = fmt.Sprintf("%s %s Invoice", site.Name, record.DateCreated.Format("January 2 2006"))

	return eml, nil
}

type payDetail struct {
	Name   string          `json:"name"`
	Amount decimal.Decimal `json:"amount"`
}

// MakeReceipt ...
func MakeReceipt(tx *pg.Tx, payID string) (*EMailMsg, error) {
	dbc := tx
	log := utils.Env.Log
	// decimal.DivisionPrecision = 2

	// get payment record
	record := view.PaymentList{}
	_, err := dbc.QueryOne(&record, "select * from payment_list where id=?", payID)

	if err != nil {
		log.Debug(err)
		return nil, err
	}

	// get site record
	site := model.Site{}
	_, err = dbc.QueryOne(&site, "select * from site where id=?", record.SiteID)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	// get resident record
	resident := new(model.Resident)
	_, err = dbc.QueryOne(resident, "select email from resident where id=?", record.ResidentID)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	// unmarshal records.Dues to payDetail{}
	details := []payDetail{}
	if err := json.Unmarshal(record.Dues, &details); err != nil {
		log.Debug(err)
		return nil, err
	}

	templates.SetDevelopmentMode(true)
	templates.AddGlobalFunc("fmtMoney", fmtMoney)

	t, err := templates.GetTemplate("receipt.jet.html")
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	vars := make(jet.VarMap)
	vars.Set("payment", record)
	vars.Set("payDate", record.DateTrx.Format("January 2 2006"))
	vars.Set("payDetails", details)
	vars.Set("association", site.Name)

	var w bytes.Buffer
	if err = t.Execute(&w, vars, nil); err != nil {
		log.Debug(err)
		return nil, err
	}

	eml, err := HTMLToEMail(w.Bytes())
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	eml.Subject = fmt.Sprintf("%s Receipt", site.Name)
	eml.To = resident.Email

	return eml, nil
}

// MakeRegConfirmation ...
func MakeRegConfirmation(tx *pg.Tx, site *model.Site, user *model.User) (*EMailMsg, error) {
	log := utils.Env.Log
	// decimal.DivisionPrecision = 2

	templates.SetDevelopmentMode(true)
	templates.AddGlobalFunc("fmtMoney", fmtMoney)

	t, err := templates.GetTemplate("reconfirm.jet.html")
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	details := []struct {
		Field string
		Value interface{}
	}{
		{"Website", fmt.Sprintf("https://%s.eveng.com", site.Subdomain)},
		{"Admin Login", user.Email},
		{"Admin Password", "*** (use the password you supplied durring registration)"},
	}

	vars := make(jet.VarMap)
	vars.Set("user", user)
	vars.Set("site", site)
	vars.Set("date", time.Now().Format("January 2 2006"))
	vars.Set("association", site.Name)
	vars.Set("details", details)

	var w bytes.Buffer
	if err = t.Execute(&w, vars, nil); err != nil {
		log.Debug(err)
		return nil, err
	}

	eml, err := HTMLToEMail(w.Bytes())
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	eml.Subject = fmt.Sprintf("eve: %s", site.Name)
	eml.To = user.Email

	return eml, nil
}

// NewResidents ...
func NewResidents(tx *pg.Tx, resident *model.Resident, password string) (*EMailMsg, error) {
	// logger for the app
	log := utils.Env.Log

	templates.SetDevelopmentMode(true)

	t, err := templates.GetTemplate("residents_email.jet.html")

	if err != nil {
		log.Debug(err)
		return nil, err
	}

	// variables to set
	vars := make(jet.VarMap)

	fullname := fmt.Sprintf("%s %s", resident.FirstName, resident.LastName)
	vars.Set("fullname", fullname)
	vars.Set("email", resident.Email)
	vars.Set("password", password)

	var w bytes.Buffer

	if err = t.Execute(&w, vars, nil); err != nil {
		log.Debug(err)
		return nil, err
	}

	eml, err := HTMLToEMail(w.Bytes())
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	eml.Subject = fmt.Sprintf("eve: User Onboarding")
	eml.To = resident.Email

	return eml, nil
}
