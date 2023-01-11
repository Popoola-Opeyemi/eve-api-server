package handlers

import (
	"encoding/json"
	"errors"
	"eve/service/form"
	"eve/shared"
	"fmt"

	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// InvDue ...
type InvDue struct {
	DueID  string          `json:"due_id"`
	Name   string          `json:"name"`
	Amount decimal.Decimal `json:"amount"`
}

// SavePayment ...
func SavePayment(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrType := ses.Int("admin_type")

	// do not allow all lower tier except residents
	if usrType < model.OfficialUser && usrType != model.ResidentUser {
		err := fmt.Errorf("Access denied")
		return true, err
	}

	// do not allow secondary residents to make payment
	if usrType == model.SecondaryResident {
		err := fmt.Errorf("Access denied")
		return true, err
	}

	siteID := getSiteID(c)
	apiForm := frm.(*form.PaymentForm)

	oid := c.Param("id")
	payment := &model.Payment{}

	if err := copier.Copy(payment, apiForm); err != nil {
		log.Debug(err)
		return false, err
	}

	// allow only payments by residents online to go through
	if usrType == model.ResidentUser && apiForm.PayMode == model.ManualPayment {
		err := fmt.Errorf("Access denied, pay online")
		return true, err
	}

	// unmarshal attributes
	attrEntity := form.PayProviderResponse{}
	if err := json.Unmarshal(apiForm.Attr, &attrEntity); err != nil {
		return false, err
	}

	invDues := []InvDue{}
	if err := json.Unmarshal(apiForm.Dues, &invDues); err != nil {
		log.Debug(err)
		return false, err
	}

	// store for verification data received from api
	var metadata interface{}

	if apiForm.PayMode == model.BankTransaction && usrType == model.ResidentUser {
		log.Debug("Here 1")

		if err := handleBankTransaction(*apiForm, siteID, ses, tx); err != nil {
			return true, err
		}

		return true, nil
	}

	/* Beginning of payment verification (online payment)
	- ensure the online payment can only occur for residents
	- get the provider details
	- verify the payment made
	*/
	if payment.PayMode == model.OnlinePayment && usrType == model.ResidentUser {
		log.Debug("Here 2")

		var ref string
		provider := model.GetProvider(attrEntity.ProviderID)
		log.Debug("provider", provider)
		attrEntity.ProviderName = provider.Name
		field := fmt.Sprintf("%s_secret", provider.Name)

		secretKey, err := utils.Getkey("testpayments", field)
		if err != nil {
			return false, err
		}

		if provider.Value == model.ProviderPaystack {
			ref = attrEntity.ReferenceID
		}
		if provider.Value == model.ProviderFlutterwave {
			ref = attrEntity.TransactionID
		}

		url, err := provider.GetURL(ref)
		if err != nil {
			return false, err
		}

		response, err := utils.VerifyPayment(url, secretKey)
		if err != nil {
			return false, err
		}

		json.Unmarshal(response, &metadata)

		if !handleResponse(response, provider.Value) {
			return true, errors.New("could not verify transaction")
		}

		payment.Metadata, _ = json.Marshal(metadata)
		payment.Attr, _ = json.Marshal(attrEntity)

		if err = handleValidPayment(ses, oid, siteID, invDues, *apiForm, payment, tx, log); err != nil {
			return true, err
		}
	}

	// all payments except online transaction
	if payment.PayMode != model.OnlinePayment {
		log.Debug("Here 3")

		payment.Attr = apiForm.Attr
		if err = handleValidPayment(ses, oid, siteID, invDues, *apiForm, payment, tx, log); err != nil {
			return true, err
		}
	}

	return true, nil
}

// BeforeListPayment ...
func BeforeListPayment(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
	dbc := utils.Env.Db
	log := utils.Env.Log
	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrID := ses.String("admin_id")
	usrType := ses.Int("admin_type")
	usrSubType := ses.Int("admin_subtype")

	// if this user is a secondary resident
	if usrType == 3 && usrSubType == 1 {
		// get resident record
		res := model.Resident{}
		err := dbc.Model(&res).Where("id = ?", usrID).Select()
		if err != nil {
			log.Debug(err)
			return false, err
		}

		(*filter)["resident_id"] = res.PrimaryID
	} else if usrType == 3 {
		(*filter)["resident_id"] = usrID
	}

	return false, nil
}

// DeletePayment ...
func DeletePayment(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (bool, error) {

	log := utils.Env.Log
	dbc := utils.Env.Db

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrType := ses.Int("admin_type")
	if usrType < 5 {
		err := fmt.Errorf("Access denied")
		return true, err
	}

	siteID := getSiteID(c)
	oid := c.Param("id")

	payment := model.Payment{}

	if err := dbc.Model(&payment).Where("id = ? ", oid).Select(); err != nil {
		return true, errors.New("invalid transaction")
	}

	res := model.Resident{}
	if err := tx.Model(&res).Where("id = ?", payment.ResidentID).Select(); err != nil {
		return false, err
	}

	payLog := &model.PaymentLog{
		ID:        xid.New().String(),
		SiteID:    siteID,
		Operation: string(model.PaymentDelete),
		Amount:    payment.Amount,

		InitiatedBY: model.UserDetails{
			UserID:   ses.String("admin_id"),
			UserType: ses.Int("admin_type"),
			Name:     ses.String("admin_name"),
		},

		InitiatedFor: model.UserDetails{
			UserID:   res.ID,
			UserType: model.ResidentUser,
			Name:     fmt.Sprintf("%s %s", res.FirstName, res.LastName),
		},
		Narration: "Deleted Payment Record",
	}

	_, err = tx.Model(payLog).Insert()
	if err != nil {
		return false, err
	}

	// ensure cannot delete online transactions
	if payment.PayMode == model.OnlinePayment {
		err := errors.New("Cannot delete online payments")
		return true, err
	}

	// delete related transaction records
	_, err = tx.Exec("delete from transaction where site_id = ? and payment_id = ? and type=1", siteID, oid)
	if err != nil {
		log.Debug(err)
		return true, err
	}

	return false, nil

}

func handleResponse(response []byte, provider model.PayProvider) bool {
	log := utils.Env.Log

	switch provider {
	case model.ProviderPaystack:
		value := model.PaystackResponse{}
		json.Unmarshal(response, &value)
		if value.Status {
			return true
		}

	case model.ProviderFlutterwave:
		value := model.FlutterwaveResponse{}
		json.Unmarshal(response, &value)
		log.Debugf("%+v", value)
		if value.Status == "success" {
			return true
		}
	}
	return false
}

func handleValidPayment(ses *et.SessionMgr, oid, siteID string, invDues []InvDue, apiForm form.PaymentForm, payment *model.Payment, tx *pg.Tx, log *zap.SugaredLogger) error {
	payment.SiteID = siteID

	if len(oid) == 0 || oid == "new" {

		// insert payment record
		payment.ID = xid.New().String()
		_, err := tx.Model(payment).Insert()

		if err != nil {
			log.Debug(err)
			return err
		}

	} else {
		_, err := tx.Model(payment).
			Column("date_trx", "amount", "attr").
			WherePK().
			Update()
		if err != nil {
			log.Debug(err)
			return err
		}

		// delete related transaction records
		_, err = tx.Exec("delete from transaction where site_id = ? and payment_id = ? and type=1", siteID, payment.ID)
		if err != nil {
			log.Debug(err)
			return err
		}
	}

	for _, i := range invDues {
		trx := &model.Transaction{
			ID:         xid.New().String(),
			SiteID:     siteID,
			Type:       1,
			ResidentID: payment.ResidentID,
			PaymentID:  payment.ID,
			DateTrx:    apiForm.DateTrx,
			DueID:      i.DueID,
			Amount:     i.Amount,
		}
		_, err := tx.Model(trx).Insert()
		if err != nil {
			log.Debug(err)
			return err
		}
	}

	res := model.Resident{}
	if err := tx.Model(&res).Where("id = ?", payment.ResidentID).Select(); err != nil {
		return err
	}

	payLog := &model.PaymentLog{
		ID:        xid.New().String(),
		SiteID:    siteID,
		Operation: model.PaymentInsert,
		Amount:    payment.Amount,

		InitiatedBY: model.UserDetails{
			UserID:   ses.String("admin_id"),
			Name:     ses.String("admin_name"),
			UserType: ses.Int("admin_type"),
		},

		InitiatedFor: model.UserDetails{
			UserID:   res.ID,
			UserType: model.ResidentUser,
			Name:     fmt.Sprintf("%s %s", res.FirstName, res.LastName),
		},
		Narration: "Created new payment",
	}

	_, err := tx.Model(payLog).Insert()

	if err != nil {
		log.Debug(err)
		return err
	}

	// email receipt
	eml, err := shared.MakeReceipt(tx, payment.ID)
	if err != nil {
		log.Debug(err)
		return err
	}
	_, err = tx.Exec(`
	insert into task_queue (site_id, type, data)
		values(?, 1, ?)
	`, siteID, &eml)
	if err != nil {
		log.Debug(err)
		return err
	}

	return nil
}

func handleBankTransaction(form form.PaymentForm, siteID string, ses *et.SessionMgr, tx *pg.Tx) error {

	log := utils.Env.Log
	res := model.Resident{}
	if err := tx.Model(&res).Where("id = ?", form.ResidentID).Select(); err != nil {
		return err
	}

	pendingPayment := &model.PaymentPending{
		ID:           xid.New().String(),
		SiteID:       siteID,
		ResidentName: fmt.Sprintf("%s %s", res.FirstName, res.LastName),
		ResidentID:   form.ResidentID,
		Narration:    "Bank Transaction",
		Amount:       form.Amount,
		PayMode:      model.BankTransaction,
		Dues:         form.Dues,
		Attr:         form.Attr,
	}

	_, err := tx.Model(pendingPayment).Insert()
	if err != nil {
		log.Debug(err)
		return err
	}

	payLog := &model.PaymentLog{
		ID:        xid.New().String(),
		SiteID:    siteID,
		Operation: model.PaymentInsert,
		Amount:    form.Amount,

		InitiatedBY: model.UserDetails{
			UserID:   ses.String("admin_id"),
			UserType: ses.Int("admin_type"),
			Name:     ses.String("admin_name"),
		},

		InitiatedFor: model.UserDetails{
			UserID:   res.ID,
			UserType: model.ResidentUser,
			Name:     fmt.Sprintf("%s %s", res.FirstName, res.LastName),
		},
		Narration: "Created new pending payment Bank Transaction",
	}

	_, err = tx.Model(payLog).Insert()

	if err != nil {
		log.Debug(err)
		return err
	}

	return nil
}
