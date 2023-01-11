package handlers

import (
	"encoding/json"
	"net/http"

	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

// BeforeBillSave ...
func BeforeBillSave(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	log := utils.Env.Log
	siteID := getSiteID(c)
	oid := c.Param("id")

	record := frm.(*model.Bill)

	// there can only be one active bill for each unit type
	// so we disable all other bill records for this unit_type on create
	if len(oid) == 0 || oid == "new" {
		_, err := tx.Exec(`
			update bill
				set status = 0
			where
				site_id = ?
				and unit_type = ?
		`, siteID, record.UnitType)
		if err != nil {
			log.Debug(err)
			return false, err
		}
		// if bill.Status == 1 then set all other bills of this unit_type to status = 0
	} else if record.Status == 1 {
		_, err := tx.Exec(`
			update bill
				set status = 0
			where
				site_id = ?
				and unit_type = ?
				and id != ?
		`, siteID, record.UnitType, oid)
		if err != nil {
			log.Debug(err)
			return false, err
		}
	}

	return false, nil
}

// AfterReadBill ...
func AfterReadBill(c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (err error) {
	svc := utils.CRUDServiceInstance
	// log := utils.Env.Log

	bill := frm.(*model.Bill)

	opts := utils.Options{}
	opts["bill_id"] = bill.ID
	records, err := svc.List("BillItemList", opts, "")
	if err != nil {
		et.APIError(c, err, http.StatusInternalServerError)
		return err
	}

	bill.Items, err = json.Marshal(records)
	if err != nil {
		et.APIError(c, err, http.StatusInternalServerError)
		return err
	}

	return nil
}

// AfterSaveBill ...
func AfterSaveBill(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (stop bool, err error) {
	svc := utils.CRUDServiceInstance
	// log := utils.Env.Log

	oid := c.Param("id")
	siteID := getSiteID(c)

	bill := frm.(*model.Bill)

	if len(oid) > 0 && oid != "new" {
		_, err = utils.Env.Db.Exec("delete from bill_item where site_id = ? and bill_id=?", siteID, bill.ID)
		if err != nil {
			et.APIError(c, err, http.StatusInternalServerError)
			return false, err
		}
	}

	billItems := []model.BillItem{}
	if err := json.Unmarshal(bill.Items, &billItems); err != nil {
		et.APIError(c, err, http.StatusInternalServerError)
		return false, err
	}

	modelType := "BillItem"
	for _, i := range billItems {
		i.BillID = bill.ID
		i.SiteID = siteID

		if err = svc.Create(tx, modelType, &i, true); err != nil {
			et.APIError(c, err, http.StatusInternalServerError)
			return false, err
		}
	}

	return false, nil
}
