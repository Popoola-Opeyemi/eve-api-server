package handlers

import (
	"errors"
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

// SaveResidencyProfile ...
func SaveResidencyProfile(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {

	log := utils.Env.Log
	dbc := utils.Env.Db
	svc := utils.CRUDServiceInstance

	filter := utils.Options{}

	// getting the form
	form := frm.(*model.Residency)

	// log.Debugf("%+v", form)

	// validating if the person is a resident
	result, err := svc.GetBy("Resident", "id", form.ID, filter, "")

	if err != nil {
		return true, err
	}

	resident, _ := result.(*model.Resident)

	/* Ending a resident
	- ensure the resident is a primary resident
	- ensure the active status is set to 0
	- modify the end date of the residency
	*/

	if resident.Type == model.SecondaryResident {
		msg := ""
		switch resident.ActiveStatus {
		case model.ResidencyActive:
			msg = "end primary resident first"
		case model.ResidencyEnded:
			msg = "reinstate primary resident first"
		}
		dError := errors.New(msg)
		return true, dError
	}

	if form.ActiveStatus == model.ResidencyEnded {

		result, err := svc.GetBy("Residency", "id", resident.ResidencyID, filter, "")

		if err != nil {
			return true, err
		}

		res, _ := result.(*model.Residency)

		residency := model.Residency{
			ID:             resident.ResidencyID,
			PreviousUnitID: res.UnitID,
			UnitID:         "",
			DateExit:       utils.DateTime{}.Now(),
			ActiveStatus:   model.ResidencyEnded,
		}

		_, err = dbc.Model(&residency).
			Set("date_exit =?date_exit, unit_id = ?unit_id, active_status = ?active_status, previous_unit_id= ?previous_unit_id").
			Where("id = ?id").Update()

		if err != nil {
			return false, err
		}
	}

	/* Resinstating a resident
	- query resident using the id and get the
	- ensure the resident is a primary resident
	- ensure the active status is set to 1
	- confirm if resident previous unit hasn't been assigned to another person
	- reinstate resident based on result
	*/

	if form.ActiveStatus == model.ResidencyActive && form.PreviousUnitID != "" {

		residencyQuery, err := svc.GetBy("Residency", "previous_unit_id", form.PreviousUnitID, filter, "")
		if err != nil {
			if err == pg.ErrNoRows {
				err = errors.New("Cannot reinstate invalid unit")
			}
			return true, err
		}

		residency := residencyQuery.(*model.Residency)
		log.Debug("Residency Found is")
		log.Debugf("%+v", residency)

		unitActiveQuery, err := svc.GetBy("Residency", "unit_id", residency.PreviousUnitID, filter, "")

		if err != nil && err != pg.ErrNoRows {
			return true, err
		}

		unitActive := unitActiveQuery.(*model.Residency)
		log.Debug("Unit active found is ")
		log.Debugf("%+v", unitActive)

		if unitActive.ID != "" && unitActive.ActiveStatus == model.ResidencyActive {
			nError := errors.New("active resident present on previous unit")

			return true, nError
		}

		// if unit is not active
		if err == pg.ErrNoRows && unitActive.ID == "" {
			_ = unitActive

			form.ID = residency.ID
			form.UnitID = residency.PreviousUnitID
			form.DateStart = utils.DateTime{}.Now()
			form.DateExit = utils.DateTime{}
			form.ActiveStatus = model.ResidencyActive
			form.PreviousUnitID = ""

			log.Debug("sending value ...")
			log.Debugf("%+v", form)
		}
	}

	return false, nil
}

// AfterSaveResidency ...
func AfterSaveResidency(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (stop bool, err error) {
	dbc := utils.Env.Db
	log := utils.Env.Log

	record := frm.(*model.Residency)

	log.Debugf("%+v", record)

	_ = dbc

	return false, nil
}

// func reinstateResident() error {

// 	return nil
// }
