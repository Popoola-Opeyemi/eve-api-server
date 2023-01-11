package handlers

import (
	"encoding/json"
	"eve/service/model"
	"eve/service/view"
	"eve/shared"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"net/http"
	"time"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
)

// BeforeSaveBillGenerate ...
func BeforeSaveBillGenerate(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	log := utils.Env.Log
	siteID := getSiteID(c)
	record := frm.(*model.BillGenerate)

	// get number of unit types
	utCount := 0
	_, err := tx.Query(
		&utCount,
		"select count(id) from unit_type",
	)
	if err != nil {
		log.Debug(err)
		return true, err
	}

	// get list of active bills
	activeBills := []string{}
	_, err = tx.Query(
		&activeBills,
		"select id from bill where site_id = ? and status = 1",
		siteID,
	)
	if err != nil {
		log.Debug(err)
		return true, err
	}

	if utCount > len(activeBills) {
		resp.Set("statusErr", "You must create bills for all unit types for bill generation to proceed")
		return true, nil
	}

	// check form bills that have been generated for this period and remove them from the list
	bills := []string{}
	for _, id := range activeBills {
		count := 0
		_, err := tx.Query(
			&count,
			"select count(id) from bill_generate where bill_id = ? and month = ? and year = ?",
			id, record.Month, record.Year,
		)
		if err != nil {
			log.Debug(err)
			return true, err
		}
		if count == 0 {
			// bills have been not generated for this bill, add to list
			bills = append(bills, id)
		}
	}

	if len(bills) == 0 {
		resp.Set("statusErr", "Residents have already been billed for this period")
		return true, nil
	}

	// generate bills
	for _, b := range bills {
		if stop, err := generateBill(b, tx, c, mi, frm, resp); err != nil {
			rerr := fmt.Errorf("an error occurred bill generating bills for this period")
			resp.APIError(rerr)
			et.APIError(c, err, http.StatusInternalServerError)
			return stop, err
		}
	}

	return true, nil
}

type billableResident struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Address   string
	UnitType  int
	DateStart time.Time
}

func generateBill(id string, tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (stop bool, err error) {
	// svc := utils.CRUDServiceInstance
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	siteID := getSiteID(c)
	record := frm.(*model.BillGenerate)

	// create a bill_generated record
	record.ID = xid.New().String()
	record.SiteID = siteID
	record.BillID = id
	record.UserID = ses.String("admin_id")

	err = tx.Insert(record)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	// get the bill to be generated
	bill := model.Bill{}
	err = tx.Model(&bill).
		Where("id = ?", record.BillID).
		Select()
	if err != nil {
		log.Debug(err)
		return false, err
	}

	// get bill items
	billItems := []view.BillItemList{}
	err = tx.Model(&billItems).
		Where("bill_id = ?", record.BillID).
		Select()
	if err != nil {
		log.Debug(err)
		return false, err
	}
	bill.Items, err = json.Marshal(billItems)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	// get residents matching bills unit_type
	residents := []billableResident{}
	_, err = tx.Query(&residents, `
		select
			r.id, r.first_name, r.last_name, r.email, rs.date_start,
			u.type as unit_type,
			concat(
				(case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end)
				, s.name, ', '||u.label
			) as "address"

		from
			resident as r
		left join residency as rs
			on rs.id = r.residency_id 

		left join unit as u
			on u.id = rs.unit_id

		left join "street" as s
			on s.id = u.street_id

		where
			r.type = 1 and rs.site_id = ? and u.type = ?
	`, siteID, bill.UnitType)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	// billMonth := gostradamus.NewDateTime(record.Year, record.Month, 1, 0, 0, 0, 0, gostradamus.UTC).
	// 	CeilMonth().Time()

	// for each resident create an invoice record and a debit transaction record for each due in the bill
	nve, _ := decimal.NewFromString("-1")
	portion, _ := decimal.NewFromString("12")

	for _, r := range residents {
		// do not bill resident if start_date > billMonth
		// if r.DateStart.Before(billMonth) == false {
		// 	log.Debug("r.date_start:", r.DateStart, "| billMonth:", billMonth)
		// 	continue
		// }

		// create invoice record
		invoice := &model.Invoice{
			ID:          xid.New().String(),
			SiteID:      siteID,
			ResidentID:  r.ID,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			Address:     r.Address,
			Month:       record.Month,
			Year:        record.Year,
			DateCreated: utils.DateTime{}.Now(),
			BillID:      record.BillID,
			UnitType:    bill.UnitType,
			Description: bill.Note,
			Amount:      bill.Total.Div(portion),
			Dues:        bill.Items,
		}
		if _, err := tx.Model(invoice).Insert(); err != nil {
			log.Debug(err)
			return false, err
		}

		// create debit transaction records for each due in bil
		for _, bi := range billItems {

			// bill (amount/12) * -1
			amt := bi.Amount.Div(portion)
			amt = amt.Mul(nve)

			transaction := &model.Transaction{
				ID:         xid.New().String(),
				SiteID:     siteID,
				ResidentID: r.ID,
				Type:       2,
				DateTrx:    utils.DateTime{}.Now(),
				InvoiceID:  invoice.ID,
				DueID:      bi.DueID,
				Amount:     amt,
			}
			if _, err := tx.Model(transaction).Insert(); err != nil {
				log.Debug(err)
				return false, err
			}
		}

		// email Invoice
		invEml, err := shared.MakeInvoice(tx, invoice.ID)
		if err != nil {
			log.Debug(err)
			return false, err
		}
		invEml.To = r.Email
		// invEml.Text = ""
		_, err = tx.Exec(`
		insert into task_queue (site_id, type, data)
			values(?, 1, ?)
		`, siteID, &invEml)
		if err != nil {
			log.Debug(err)
			return false, err
		}
	}

	return false, nil
}
