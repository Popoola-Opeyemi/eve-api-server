package handlers

import (
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

func BeforeSaveContent(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {

	dbc := utils.Env.Db

	form := frm.(*model.Content)

	_, err := dbc.Model(form).Set("data =?data").Where("id = ?id").Update()
	if err != nil {
		return true, err
	}

	return true, nil
}
