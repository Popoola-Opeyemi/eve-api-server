package handlers

import (
	"eve/service/view"
	"eve/utils"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

// AfterListGatePass ...
func AfterListSecurityResidents(c echo.Context, data interface{}, resp *utils.Response) (err error) {

	log := utils.Env.Log
	dbc := utils.Env.Db

	securityList := data.(*[]view.SecurityResidentList)

	for idx := 0; idx < len(*securityList); idx++ {
		record := &(*securityList)[idx]
		accountSummary := view.ResidentBillingSummary{}
		err = dbc.Model(&accountSummary).
			Where("id = ?", record.ID).Select()
		if err != nil && err != pg.ErrNoRows {
			log.Debug("============ error ", err)
			return err
		}

		if accountSummary.Balance.IsNegative() {
			record.InDebt = true
		}
	}

	log.Debug("========== in hook ... ", securityList)

	return nil
}
