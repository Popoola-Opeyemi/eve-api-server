package handlers

import (
	"eve/service/model"
	"eve/utils"
	"fmt"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

func (s *Controller) GetAssociationByCode(c echo.Context) error {
	dbc := s.env.Dbc
	response := utils.Response{}

	siteCode := c.Param("id")
	res := &model.Site{}

	if err := dbc.Model(res).Where("site_code = ?", siteCode).Select(); err != nil && err != pg.ErrNoRows {
		return err
	}

	if res.ID == "" {
		response.APIError(fmt.Errorf("record not found"))
		return c.JSON(http.StatusNotFound, response)

	}

	response.Set("record", res)

	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil

}
