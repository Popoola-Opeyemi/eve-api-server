package handlers

import (
	"errors"
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

func BeforeResidentSaveAlerts(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {

	dbc := utils.Env.Db
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return false, err
	}

	data := frm.(*model.ResidentAlerts)

	usrID := ses.String("admin_id")
	siteID := ses.String("admin_site_id")

	residentAlert := []model.ResidentAlerts{}

	err = dbc.Model(&residentAlert).
		Where("resident_id = ? and site_id = ? and status = ?", usrID, siteID, model.AlertSent).Select()
	if err != nil && err != pg.ErrNoRows {
		log.Debug("========= err ", err)
		return false, err
	}

	if len(residentAlert) > 0 {
		err := errors.New("Alert has been raised already")
		return true, err
	}

	if len(residentAlert) == 0 {
		data.Status = model.AlertSent
	}

	return false, nil
}
