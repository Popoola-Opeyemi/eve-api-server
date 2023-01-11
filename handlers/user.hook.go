package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

func SaveUser(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (bool, error) {
	log := utils.Env.Log
	dbc := utils.Env.Db

	log.Debug("=================before saving user")

	paramID := c.Param("id")

	log.Debug("ParamID", paramID)

	apiForm := frm.(*model.User)

	userModel := &model.User{}

	if paramID == "" {
		err := tx.Model(userModel).Where("email = ?", apiForm.Email).Select()

		if err != nil && err != pg.ErrNoRows {
			return true, err
		}

		if userModel.Email != "" {
			nError := errors.New("email already in use, choose another")
			return true, nError
		}
	}

	if paramID != "" {
		err := dbc.Model(userModel).Where("id = ?", paramID).Select()
		if err != nil {
			return true, err
		}

		if userModel.ID == "" {
			return true, nil
		}

		passwordHash := userModel.Password

		if apiForm.Password != "" && apiForm.Password != "***" {
			log.Debug("============ user supplied new password")
			passwordHash, err = utils.HashPassword("admin")
			if err != nil {
				return false, err
			}
		}

		updateForm := &model.User{
			ID:       userModel.ID,
			Status:   apiForm.Status,
			Email:    apiForm.Email,
			Phone:    apiForm.Phone,
			Password: passwordHash,
			SiteID:   userModel.SiteID,
		}

		_, err = dbc.Model(updateForm).Set(`password =?password, status =?status, 
		phone=?phone, email=?email`).Where("id =?id and site_id =?site_id").Update()
		if err != nil {
			return true, err
		}

	}

	return true, nil
}

// DeleteUser ...
func DeleteUser(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (stop bool, err error) {
	svc := utils.CRUDServiceInstance
	log := utils.Env.Log

	modelType := "User"
	oid := c.Param("id")
	siteID := getSiteID(c)

	// get user record
	filter := utils.Options{}
	filter["site_id"] = siteID
	record, err := svc.GetBy(modelType, "id", oid, filter, "")
	if err != nil {
		log.Debug(err)

		resp := utils.Response{}
		resp.APIError(fmt.Errorf("bad request"))
		c.JSON(http.StatusBadRequest, resp)

		return false, err
	}
	if record == nil {
		err = fmt.Errorf("unknown user")
		resp := utils.Response{}
		resp.APIError(err)
		c.JSON(http.StatusNotFound, resp)

		return false, err
	}

	user := record.(*model.User)
	if user.IsSiteUser {
		err = fmt.Errorf("cannot delete default association user")
		resp := utils.Response{}
		resp.APIError(err)
		c.JSON(http.StatusBadRequest, resp)

		return false, err
	}

	return false, nil
}
