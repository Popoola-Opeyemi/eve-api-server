package handlers

import (
	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

// DeleteStreet ...
func DeleteStreet(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (stop bool, err error) {
	dbc := utils.Env.Db
	log := utils.Env.Log
	oid := c.Param("id")

	// if site has streets registered refuse to delete site
	count, err := dbc.Model((*model.Unit)(nil)).Where("street_id = ?", oid).Count()
	if err != nil {
		log.Debug(err)
		return
	}
	if count > 0 {
		err = fmt.Errorf("can't delete: this street is associated with one or more units")

		resp.APIError(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	return false, nil
}
