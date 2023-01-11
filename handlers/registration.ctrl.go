package handlers

import (
	"encoding/json"
	"eve/service/form"
	"eve/service/model"
	"eve/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
)

// RegisterNewResident ...
func (s *Controller) RegisterNewResident(c echo.Context) error {

	dbc := s.env.Dbc
	log := s.log
	response := utils.Response{}

	form := form.RegistrationForm{}

	if err := c.Bind(&form); err != nil {
		return err
	}

	site := model.Site{}

	if err := dbc.Model(&site).Where("id = ? ", form.SiteID).Select(); err != nil {
		return err
	}

	// copy to table model
	fullName := fmt.Sprintf("%s %s", form.FirstName, form.LastName)
	record := model.NewResidentRegistrations{
		Name:      fullName,
		ID:        xid.New().String(),
		SiteID:    form.SiteID,
		SiteName:  site.Name,
		Email:     strings.ToLower(form.Email),
		FirstName: form.FirstName,
		LastName:  form.LastName,
		Address:   form.Address,
		Phone:     form.Phone,
		Attr:      form.Attr,
	}

	if err := dbc.Insert(&record); err != nil {
		log.Debug(err)
		return err
	}

	response.Set("status", "ok")

	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
	}

	return nil
}

// RegisterAssoc ...
func (s *Controller) RegisterAssoc(c echo.Context) error {
	dbc := s.env.Dbc
	log := s.log
	var err error
	// siteID := getSiteID(c)

	// get form data sent
	form := form.Registration{}
	if err := c.Bind(&form); err != nil {
		log.Debug(err)
		return err
	}

	// copy to table model
	record := model.Registration{}
	record.Data, err = json.Marshal(form)
	if err != nil {
		log.Debug(err)
		return err
	}

	// insert record
	record.ID = xid.New().String()
	if err := dbc.Insert(&record); err != nil {
		log.Debug(err)
		return err
	}

	// send ok response
	resp := utils.Response{}
	resp.Set("count", 1)

	if err := c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil
}
