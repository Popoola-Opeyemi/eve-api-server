package handlers

import (
	"eve/service/form"
	"eve/service/view"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
)

var colName = make(map[string]string)
var filter = utils.Options{}

// BeforeSaveResident ...
func BeforeSaveResident(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {

	log := utils.Env.Log
	dbc := utils.Env.Db
	svc := utils.CRUDServiceInstance

	// form received from client
	form := frm.(*model.Resident)
	form.Email = strings.ToLower(form.Email)

	// determine the column name to use for query
	switch form.Type {
	case model.PrimaryResident:
		colName["column"] = "id"
		colName["value"] = form.ID
		break
	case model.SecondaryResident:
		colName["column"] = "primary_id"
		colName["value"] = form.PrimaryID
	}

	// using the id of the resident to search
	if form.Type == model.SecondaryResident && form.ID != "" {
		filter[colName["column"]] = form.ID
	}

	// Disabling a Primary resident
	if form.ID != "" && form.PrimaryID == "" && form.Type == model.PrimaryResident {
		if err := disablePrimResident(form, c); err != nil {
			return true, err
		}
	}

	// Disabling a Secondary resident
	if form.ID != "" && form.PrimaryID != "" && form.Type == model.SecondaryResident {
		if err := disableSecResident(form, c); err != nil {
			return true, err
		}
	}

	/* creating a new resident
	 	- ensure the person id is null
		- ensure the resident is a primary resident
		- ensure the person primary_id is null
	*/

	if form.ID == "" &&
		form.PrimaryID == "" &&
		form.Type == model.PrimaryResident {

		log.Debug("<<<<<<<<<<<<<<< fn At create Resident >>>>>>>>>>>>>>>")

		residencyID := xid.New().String()
		residency := model.Residency{
			ID:           residencyID,
			UnitID:       form.UnitID,
			DateStart:    utils.DateTime{}.Now(),
			SiteID:       getSiteID(c),
			ActiveStatus: model.ResidencyActive,
		}
		_, err := dbc.Model(&residency).Insert()

		if err != nil {
			log.Debug(err)
			return false, err
		}

		form.ResidencyID = residencyID
		filter = utils.Options{}

	}

	/* creating a new secondary resident
	 	- ensure the resident id is null
		- ensure the resident is a secondary resident
		- ensure the person primary_id is not null
	*/

	if form.Type == model.SecondaryResident &&
		form.PrimaryID != "" && form.ID == "" {
		log.Debug("<<<<<<<<<<<<<<< fn Create a Secondary Resident >>>>>>>>>>>>>>>")

		result, err := svc.GetBy("Resident", "id", form.PrimaryID, filter, "")

		if err != nil && err != pg.ErrNoRows {
			et.APIError(c, err, http.StatusInternalServerError)
			return true, err
		}

		resident, _ := result.(*model.Resident)

		form.ResidencyID = resident.ResidencyID
		form.Status = resident.Status
		filter = utils.Options{}

	}
	log.Debug("i came and i saw")

	return false, nil

}

// DeleteResident ...
func DeleteResident(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (stop bool, err error) {

	// log := utils.Env.Log
	dbc := utils.Env.Db

	rID := c.Param("id")

	resident := model.Resident{}
	residency := model.Residency{}

	if err := dbc.Model(&resident).Where("id = ?", rID).Select(); err != nil {
		return true, err
	}

	if resident.Type == model.PrimaryResident {
		_, err = dbc.Model(&resident).Where("id = ?", resident.ID).Delete()

		if err != nil {
			return false, err
		}

		_, err = dbc.Model(&residency).Where("id = ?", resident.ResidencyID).Delete()

		if err != nil {
			return false, err
		}

		// operation completed successfully no need to return to the query handler
		if err == nil {
			return true, nil
		}
	}

	return
}

// ReadResident ...
func ReadResident(c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (err error) {
	svc := utils.CRUDServiceInstance
	siteID := getSiteID(c)

	// log := utils.Env.Log
	record := frm.(*model.Resident)

	filter := utils.Options{}
	if len(siteID) > 0 {
		filter["site_id"] = siteID
	}
	retv, err := svc.GetBy("ResidentView", "id", record.ID, filter, "")
	if err != nil {
		et.APIError(c, err, http.StatusInternalServerError)
		return err
	}

	rv, _ := retv.(*view.ResidentView)
	record.Unit = rv.Unit

	return nil
}

// SaveResidentProfile ...
func SaveResidentProfile(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {

	log := utils.Env.Log

	record := frm.(*form.ResidentProfileSave)
	_, err := tx.Model(record).
		WherePK().
		Update()

	if err != nil {
		log.Debug(err)

		et.APIError(c, err, http.StatusInternalServerError)
		return false, err
	}

	return true, nil
}

// BeforeReadResidentBilling ...
func BeforeReadResidentBilling(c echo.Context, mi *et.ModelInfo, field, value string, filter *utils.Options, resp *utils.Response) (bool, error) {
	dbc := utils.Env.Db
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrType := ses.Int("admin_type")
	usrSubType := ses.Int("admin_subtype")

	// if this user is a secondary resident
	if usrType == 3 && usrSubType == 1 {
		// get resident record
		res := model.Resident{}
		err := dbc.Model(&res).Where("id = ?", value).Select()
		if err != nil {
			log.Debug(err)
			return false, err
		}

		// get billing summary based on primary_id
		resbill := view.ResidentBillingSummary{}
		err = dbc.Model(&resbill).Where("id = ?", res.PrimaryID).Select()
		if err != nil {
			log.Debug(err)
			return false, err
		}

		resp.Set("record", resbill)
		return true, nil
	}

	return false, nil
}

// BeforeResidentFamilyList ...
func BeforeResidentFamilyList(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
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

	res := model.Resident{}
	err = dbc.Model(&res).Where("id = ?", usrID).Select()
	if err != nil {
		log.Debug(err)
		return false, err
	}
	// if this user is a secondary resident
	if usrType == 3 && usrSubType == 1 {
		// get resident record

		(*filter)["primary_id"] = res.PrimaryID
	} else if usrType == 3 {
		(*filter)["residency_id"] = res.ResidencyID
		(*filter)["$order"] = "type"
		delete(*filter, "primary_id")
	}

	return false, nil
}

// DeleteResidentFamilyMember ...
func DeleteResidentFamilyMember(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (bool, error) {
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrType := ses.Int("admin_type")
	usrSubType := ses.Int("admin_subtype")

	if usrType != 3 && usrSubType == 0 {
		err := fmt.Errorf("Access denied")
		return true, err
	}

	return false, nil
}

// BeforeAccountHistoryList ...
func BeforeAccountHistoryList(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
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
	residentID := ""
	if usrType == 3 && usrSubType == 1 {
		// get resident record
		res := model.Resident{}
		err := dbc.Model(&res).Where("id = ?", usrID).Select()
		if err != nil {
			log.Debug(err)
			return false, err
		}

		residentID = res.PrimaryID
	} else if usrType == 3 {
		residentID = usrID
	}

	records := []view.AccountHistory{}
	res, err := dbc.Query(&records, `
		select
			h.resident_id,  h.document_id, h.date_trx, h.type,h.invoice_number,
			h.amount, sum(j.amount) as balance
		from account_history as h
		left join account_history as j on
			(j.resident_id = h.resident_id and j.invoice_number = h.invoice_number and h.date_trx >= j.date_trx)
		where
			h.resident_id = ?
		group by
			h.resident_id,  h.document_id, h.date_trx, h.type, h.amount, h.invoice_number
		order by
			h.date_trx desc
	`, residentID)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	resp.Set("list", records)
	resp.Set("count", res.RowsReturned())

	return true, nil
}

func disablePrimResident(form *model.Resident, c echo.Context) error {

	log := utils.Env.Log
	dbc := utils.Env.Db
	svc := utils.CRUDServiceInstance

	log.Debug("<<<<<<<<<<<<<<< fn Disable Primary Resident >>>>>>>>>>>>>>>")

	result, err := svc.GetBy("Resident", colName["column"], colName["value"], filter, "")

	if err != nil {
		et.APIError(c, err, http.StatusInternalServerError)
		return err
	}

	resident, _ := result.(*model.Resident)

	form.ResidencyID = resident.ResidencyID

	/* Disabling all secondary resident associated with primary resident
	 	- get all the resident associated with primary
		- set status for those resident to be disabled
	*/

	if form.Status == model.IsDisabled {
		opt := utils.Options{}
		opt["primary_id"] = form.ID
		opt["status"] = strconv.Itoa(int(model.IsEnabled))
		res, count, err := svc.ListAndCount("Resident", opt, "Resident")

		if err != nil && err != pg.ErrNoRows {
			et.APIError(c, err, http.StatusInternalServerError)
			return err
		}

		secondaryResident, _ := res.(*[]model.Resident)

		if count > 0 {
			for _, secResident := range *secondaryResident {
				err = utils.Transact(dbc, log, func(tx *pg.Tx) error {

					secResident.Status = model.IsDisabled

					_, err = dbc.Model(&secResident).Set("status =?status").Where("id = ?id").Update()
					if err != nil {
						return err
					}

					return nil
				})
			}
		}

	}

	filter = utils.Options{}

	return nil
}

func disableSecResident(form *model.Resident, c echo.Context) error {

	log := utils.Env.Log
	// dbc := utils.Env.Db
	svc := utils.CRUDServiceInstance
	log.Debug("<<<<<<<<<<<<<<< fn Disable Secondary Resident >>>>>>>>>>>>>>>")

	// using the id of the resident to search
	filter[colName["column"]] = form.PrimaryID
	result, err := svc.GetBy("Resident", "id", form.ID, filter, "")

	if err != nil {
		et.APIError(c, err, http.StatusInternalServerError)
		return err
	}

	resident, _ := result.(*model.Resident)

	form.ResidencyID = resident.ResidencyID

	filter = utils.Options{}

	return nil

}
