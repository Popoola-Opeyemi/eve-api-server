package handlers

import (
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

// InvBillItems ...
type InvBillItems struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	UnitType      int    `json:"unit_type"`
	UnitTypeLabel string `json:"unit_type_label"`
}

// InvResidentInfo ...
type InvResidentInfo struct {
	ID        string
	FirstName string
	LastName  string
	Address   string
	UnitType  int
}

// BeforeSaveInvoiceMaster ...
func BeforeSaveInvoiceMaster(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	// log := utils.Env.Log
	// siteID := getSiteID(c)
	// record := frm.(*model.InvoiceMaster)

	// count := 0
	// _, err := tx.Query(&count, "select count(id) from invoice_master where site_id = ? and month = ? and year = ?", siteID, record.Month, record.Year)
	// if err != nil {
	// 	log.Debug(err)
	// 	return true, err
	// }

	// if count > 0 {
	// 	err := fmt.Errorf("Invoices have already been created for this period")
	// 	resp.APIError(err)
	// 	et.APIError(c, err, http.StatusBadRequest)
	// 	return true, err
	// }

	return false, nil
}

// BeforeSaveInvoice ...
func BeforeSaveInvoice(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	log := utils.Env.Log
	siteID := getSiteID(c)
	record := frm.(*model.Invoice)

	//1: get residents record

	resident := struct {
		ID        string
		BillID    string
		FirstName string
		LastName  string
		Email     string
		Address   string
		UnitType  int
	}{}
	_, err := tx.Query(&resident, `
		select
			r.id, r.first_name, r.last_name, r.email,
			u.type as unit_type,
			concat(
				(case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end)
				, s.name, ', '||u.label
			) as "address",
			b.id as "bill_id"

		from
			resident as r
		
		left join residency as rs 
			on rs.id = r.residency_id

		left join unit as u
			on u.id = rs.unit_id

		left join "street" as s
			on s.id = u.street_id

		left join bill as b
			on b.unit_type = u.type and b.status = 1

		where
			rs.site_id = ? and r.type = 1 and r.id = ?
	`, siteID, record.ResidentID)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	//2: get bill_items matching bll
	billDues := gjson.GetBytes(record.Dues, "#[*]#")

	//3: create invoice record
	invoice := &model.Invoice{
		ID:          xid.New().String(),
		SiteID:      siteID,
		ResidentID:  record.ResidentID,
		FirstName:   resident.FirstName,
		LastName:    resident.LastName,
		Address:     resident.Address,
		Month:       record.Month,
		Year:        record.Year,
		DateCreated: record.DateCreated,
		BillID:      resident.BillID,
		UnitType:    resident.UnitType,
		Description: record.Description,
		Amount:      record.Amount,
		Dues:        record.Dues,
	}
	if _, err := tx.Model(invoice).Insert(); err != nil {
		log.Debug(err)
		return false, err
	}

	//4: create transaction entries for each due in the invoice
	nve, _ := decimal.NewFromString("-1")
	for _, i := range billDues.Array() {

		// bill (amount/12) * -1
		amt, _ := decimal.NewFromString(i.Get("amount").String())
		amt = amt.Mul(nve)

		trx := &model.Transaction{
			ID:         xid.New().String(),
			SiteID:     siteID,
			ResidentID: resident.ID,
			Type:       2,
			DateTrx:    utils.DateTime{}.Now(),
			InvoiceID:  invoice.ID,
			DueID:      i.Get("due_id").String(),
			Amount:     amt,
		}
		if _, err := tx.Model(trx).Insert(); err != nil {
			log.Debug(err)
			return false, err
		}
	}

	return true, nil
}

// BeforeReadInvoice ...
func BeforeReadInvoice(c echo.Context, mi *et.ModelInfo, field, value string, filter *utils.Options, resp *utils.Response) (stop bool, err error) {
	dbc := utils.Env.Db
	log := utils.Env.Log

	siteID := c.Get("siteID").(string)

	record := &model.Invoice{}
	_, err = dbc.QueryOne(record,
		`select * from invoice_list where id = ? and site_id = ?`,
		value, siteID)
	if err != nil {
		log.Debug(err)

		resp.APIError(err)
		c.JSON(http.StatusInternalServerError, resp)
		return true, err
	}

	resp.Set("record", record)

	return true, nil
}

// // AfterSaveInvoiceMaster ...
// func AfterSaveInvoiceMaster(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {

// 	log := utils.Env.Log
// 	siteID := getSiteID(c)
// 	record := frm.(*model.InvoiceMaster)
// 	log.Debug("record.id --->", record.ID)

// 	// unmarshal InvoiceMaster.Bills into a struct
// 	invBills := []InvBillItems{}
// 	if err := json.Unmarshal(record.Bills, &invBills); err != nil {
// 		log.Debug(err)
// 		return false, err
// 	}

// 	// get a list of bill ids
// 	idList := []string{}
// 	for _, i := range invBills {
// 		idList = append(idList, i.ID)
// 	}

// 	// create a bill cache
// 	billCache := map[int]view.BillDetailList{}
// 	bills := []view.BillDetailList{}
// 	err := tx.Model(&bills).
// 		Where("site_id = ?", siteID).
// 		WhereIn("id in (?)", idList).
// 		Select()
// 	if err != nil {
// 		log.Debug(err)
// 		return false, err
// 	}

// 	for _, i := range bills {
// 		billCache[i.UnitType] = i
// 	}

// 	// get residents for this site
// 	residents := []InvResidentInfo{}
// 	_, err = tx.Query(&residents, `
// 		select
// 			r.id, r.first_name, r.last_name,
// 			u.type as unit_type,
// 			concat(
// 				(case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end)
// 				, s.name, ', '||u.label
// 			) as "address"

// 		from
// 			resident as r

// 		left join unit as u
// 			on u.id = r.unit_id

// 		left join "street" as s
// 			on s.id = u.street_id

// 		where
// 			r.type = 1 and r.site_id = ?
// 	`, siteID)
// 	if err != nil {
// 		log.Debug(err)
// 		return false, err
// 	}

// 	// prepreare invoice
// 	invoice := []model.Invoice{}
// 	transaction := []model.Transaction{}

// 	for _, r := range residents {
// 		inv := model.Invoice{
// 			ID:              xid.New().String(),
// 			SiteID:          siteID,
// 			ResidentID:      r.ID,
// 			FirstName:       r.FirstName,
// 			LastName:        r.LastName,
// 			Address:         r.Address,
// 			Month:           record.Month,
// 			Year:            record.Year,
// 			DateCreated:     utils.DateTime{}.Now(),
// 			BillID:          billCache[r.UnitType].ID,
// 			InvoiceMasterID: record.ID,
// 			UnitType:        r.UnitType,
// 			Description:     record.Description,
// 			Amount:          billCache[r.UnitType].Total,
// 			Dues:            billCache[r.UnitType].Items,
// 		}
// 		invoice = append(invoice, inv)

// 		ntv, _ := decimal.NewFromString("-1")
// 		amt := billCache[r.UnitType].Total.Mul(ntv)
// 		trx := model.Transaction{
// 			ID:         xid.New().String(),
// 			SiteID:     siteID,
// 			ResidentID: r.ID,
// 			Type:       1,
// 			DateTrx:    utils.DateTime{}.Now(),
// 			InvoiceID:  inv.ID,
// 			Amount:     amt,
// 		}
// 		transaction = append(transaction, trx)
// 	}

// 	for _, i := range invoice {
// 		if _, err := tx.Model(&i).Insert(); err != nil {
// 			log.Debug(err)
// 			return false, err
// 		}
// 	}

// 	for _, i := range transaction {
// 		if _, err := tx.Model(&i).Insert(); err != nil {
// 			log.Debug(err)
// 			return false, err
// 		}
// 	}

// 	return true, nil
// }

// BeforeInvoiceSummaryList ...
func BeforeInvoiceSummaryList(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
	dbc := utils.Env.Db
	log := utils.Env.Log
	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	oid := c.Param("id")

	usrID := ses.String("admin_id")
	usrType := ses.Int("admin_type")
	usrSubType := ses.Int("admin_subtype")

	// if this user is a secondary resident
	if usrType == 3 && usrSubType == 1 {
		// get resident record
		res := model.Resident{}
		err := dbc.Model(&res).Where("id = ?", oid).Select()
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

// DeleteInvoiceSummary ...
func DeleteInvoiceSummary(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (bool, error) {
	// prevent delete
	return true, nil
}

// BeforeInvoiceList ...
func BeforeInvoiceList(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
	dbc := utils.Env.Db
	log := utils.Env.Log
	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	// oid := c.Param("id")

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
