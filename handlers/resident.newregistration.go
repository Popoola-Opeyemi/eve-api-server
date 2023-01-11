package handlers

import (
	"errors"
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

// BeforeListNewRegistration ...
func BeforeListNewRegistration(c echo.Context, filter *utils.Options, resp *utils.Response) (stop bool, err error) {
	log := utils.Env.Log
	ses, err := et.NewSessionMgr(c, "")

	if err != nil {
		log.Debug(err)
		return
	}

	siteID := ses.String("admin_site_id")

	(*filter)["site_id"] = siteID
	delete(*filter, "status")

	return false, nil
}

// BeforeSaveNewRegistration ...
func BeforeSaveNewRegistration(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {

	log := utils.Env.Log
	dbc := utils.Env.Db

	form := frm.(*model.NewResidentRegistrations)

	if form.ID == "" {
		err := errors.New("invalid information supplied")
		return true, err
	}

	form.Name = fmt.Sprintf("%s %s", form.FirstName, form.LastName)

	_, err := dbc.Model(form).
		Set("email =?email,first_name =?first_name,last_name =?last_name,phone =?phone,address =?address,name =?name").
		Where("id = ?id").Update()

	if err != nil {
		log.Debug(err)
		return true, err
	}

	if err == nil {
		return true, nil
	}

	return false, nil
}
