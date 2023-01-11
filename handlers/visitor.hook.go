package handlers

import (
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

// BeforeReadVisitor ...
func BeforeReadVisitor(c echo.Context, mi *et.ModelInfo, field, value string, filter *utils.Options, resp *utils.Response) (stop bool, err error) {
	dbc := utils.Env.Db
	log := utils.Env.Log

	siteID := c.Get("siteID").(string)

	record := &model.Visitor{}
	_, err = dbc.QueryOne(record,
		`select * from visitor_list where id = ? and site_id = ?`,
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

// BeforeSaveVisitor ...
func BeforeSaveVisitor(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (stop bool, err error) {
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")
	userID := ses.String("admin_id")

	record := frm.(*model.Visitor)
	// if its a resident creating this visitor record
	if len(record.ID) == 0 && record.RegistrationType == 1 && usrType == 3 {
		record.ResidentID = userID
	}

	// if its security creating this visitor record
	if len(record.ID) == 0 && len(record.SecurityID) == 0 && usrType == 2 {
		record.SecurityID = userID
	}

	// get the residents unit id
	_, err = tx.QueryOne(&record.UnitID, "select rs.unit_id as unit_id from resident as r left join residency as rs on rs.id = r.residency_id where r.id = ?", record.ResidentID)
	if err != nil {
		log.Debug(err)
		return
	}

	return false, err
}

// BeforeVisitorList ...
func BeforeVisitorList(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
	dbc := utils.Env.Db
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	usrType := ses.Int("admin_type")
	userID := ses.String("admin_id")
	usrSubtype := ses.Int("admin_subtype")

	if usrType == model.ResidentUser {
		// this is a secondary user get its primary_id and use that to filter
		unitID := ""
		_, err := dbc.QueryOne(&unitID, "select rs.unit_id as unit_id from resident as r left join residency as rs on rs.id = r.residency_id where r.id = ? ", userID)

		if err != nil {
			log.Debug(err)
			return false, err
		}

		(*filter)["unit_id"] = unitID

		// secondary resident, filter using resident id
		if usrSubtype == 1 {
			(*filter)["resident_id"] = userID
		}

	}

	return false, nil
}
