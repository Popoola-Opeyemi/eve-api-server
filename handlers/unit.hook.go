package handlers

import (
	"encoding/json"
	"errors"
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

func BeforeSaveUnit(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	dbc := utils.Env.Db
	// log := utils.Env.Log

	Unit := model.Unit{}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		return true, err
	}

	siteID := ses.String("admin_site_id")

	unitNo := struct {
		UnitNumber string `json:"unit_number"`
	}{}

	form := frm.(*model.Unit)

	if err := json.Unmarshal(form.Attr, &unitNo); err != nil {
		return true, err
	}

	if _, err := strconv.Atoi(unitNo.UnitNumber); err != nil {
		err = errors.New("unit no must be an integer")
		return true, err
	}

	labelQuery := ""
	if form.Label != "" {
		labelQuery = fmt.Sprintf("and label like '%s'", form.Label)
	} else {
		labelQuery = ""
	}

	// checking to ensure unit information has not been filled before
	Query := fmt.Sprintf("select * from unit where street_id = '%s' and type = %d and site_id = '%s' and attr->>'unit_number' = '%s' %s", form.StreetID, form.Type, siteID, unitNo.UnitNumber, labelQuery)

	_, err = dbc.Query(&Unit, Query)

	if err != nil && err != pg.ErrNoRows {
		return false, err
	}

	if Unit.ID != "" && Unit.SiteID != "" && form.ID == "" {
		return true, errors.New("Unit address taken, please select another")
	}

	return false, nil
}

// DeleteUnit ...
func DeleteUnit(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (stop bool, err error) {
	dbc := utils.Env.Db
	log := utils.Env.Log
	oid := c.Param("id")

	// if site has streets registered refuse to delete site
	count, err := dbc.Model((*model.Residency)(nil)).Where("unit_id = ?", oid).Count()
	if err != nil {
		log.Debug(err)
		return
	}
	if count > 0 {
		err = fmt.Errorf("can't delete: unit is assigned to a resident")

		resp.APIError(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	return false, nil
}
