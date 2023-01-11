package handlers

import (
	"encoding/json"
	"eve/service/model"
	"eve/shared"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
)

func DeletePendingPayment(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (stop bool, err error) {
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrType := ses.Int("admin_type")

	if usrType < model.OfficialUser {
		err := fmt.Errorf("Access denied")
		return true, err
	}

	return false, nil
}

func ListPendingPayment(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}
	usrID := ses.String("admin_id")
	usrType := ses.Int("admin_type")

	if usrType == model.ResidentUser {
		(*filter)["resident_id"] = usrID
	}
	return false, nil
}

func SavePendingPayment(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrType := ses.Int("admin_type")

	if usrType < model.OfficialUser {
		err := fmt.Errorf("Access denied")
		return true, err
	}
	// // do not allow all lower tier except residents
	// if usrType < model.OfficialUser && usrType != model.ResidentUser {
	// 	err := fmt.Errorf("Access denied")
	// 	return true, err
	// }

	APIform := frm.(*model.PaymentPending)

	pendingPay := &model.PaymentPending{}

	if err := tx.Model(pendingPay).Where("id = ?", APIform.ID).Select(); err != nil {
		return false, err
	}

	payment := &model.Payment{
		ID:         xid.New().String(),
		ResidentID: pendingPay.ResidentID,
		SiteID:     pendingPay.SiteID,
		Narration:  fmt.Sprintf("approved %s", pendingPay.Narration),
		Amount:     pendingPay.Amount,
		Attr:       pendingPay.Attr,
		Dues:       pendingPay.Dues,
		PayMode:    model.BankTransaction,
	}
	_, err = tx.Model(payment).Insert()

	if err != nil {
		log.Debug(err)
		return true, err
	}

	invDues := []InvDue{}
	if err := json.Unmarshal(pendingPay.Dues, &invDues); err != nil {
		log.Debug(err)
		return false, err
	}
	log.Debugf("%+v", invDues)

	// insert transaction records
	for _, i := range invDues {
		trx := &model.Transaction{
			ID:         xid.New().String(),
			SiteID:     pendingPay.SiteID,
			Type:       1,
			ResidentID: payment.ResidentID,
			PaymentID:  payment.ID,
			DateTrx:    utils.DateTime{}.Now(),
			DueID:      i.DueID,
			Amount:     i.Amount,
		}
		_, err := tx.Model(trx).Insert()
		if err != nil {
			log.Debug(err)
			return true, err
		}
	}
	log.Debug("AA", pendingPay)

	payLog := &model.PaymentLog{
		ID:        xid.New().String(),
		SiteID:    pendingPay.SiteID,
		Operation: model.PaymentInsert,
		Amount:    pendingPay.Amount,

		InitiatedBY: model.UserDetails{
			UserID:   ses.String("admin_id"),
			UserType: ses.Int("admin_type"),
		},

		InitiatedFor: model.UserDetails{
			UserID:   pendingPay.ResidentID,
			UserType: model.ResidentUser,
		},

		Narration: "Approved Bank Transaction Payment",
	}

	_, err = tx.Model(payLog).Insert()

	if err != nil {
		log.Debug(err)
		return false, err
	}
	// deleting item from the pending payment table
	_, err = tx.Model(pendingPay).Where("id = ?", pendingPay.ID).Delete()

	if err != nil {
		log.Debug(err)
		return true, err
	}

	// email receipt
	eml, err := shared.MakeReceipt(tx, payment.ID)
	if err != nil {
		log.Debug(err)
		return false, err
	}
	// eml.Text = ""
	_, err = tx.Exec(`
	insert into task_queue (site_id, type, data)
		values(?, 1, ?)
	`, pendingPay.SiteID, &eml)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	return true, nil
}
