package handlers

import (
	"eve/utils"

	"github.com/labstack/echo/v4"
)

func BeforeUnitList(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {

	// dbc := utils.Env.Db
	log := utils.Env.Log

	log.Debug("hello there")

	// ses, err := et.NewSessionMgr(c, "")
	// if err != nil {
	// 	log.Debug(err)
	// 	return false, err
	// }

	// records := []view.UnitList{}

	// res, err := dbc.Query(&records,
	// `select ul.site_id, ul.type, ul.street_id, ul.attr, ul.label, ul.street, ul.unit_type, `)

	return false, nil
}

// select * from unit_list as ul join residency as rs on rs.site_id = ul.site_id;
