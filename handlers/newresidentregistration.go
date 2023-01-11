package handlers

import (
	"errors"
	"eve/service/form"
	"eve/service/model"
	"eve/service/view"
	"eve/shared"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"net/http"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

// NewResidentRegistration ...
type NewResidentRegistration struct {
	log  *zap.SugaredLogger
	env  *et.Env
	svc  utils.CRUDService
	Path string

	AcsMgr *et.AccessMgr
}

// Initialize ...
func (s *NewResidentRegistration) Initialize(env *et.Env) error {
	s.env = env
	s.log = env.Log.Sugar()
	svc := utils.CRUD{}
	err := svc.Init(s.env.Dbc, env.Log)
	if err != nil {
		s.log.Error(err)
		return err
	}
	s.svc = svc

	s.AcsMgr = et.NewAccessMgr()
	s.AcsMgr.AddRules([]et.AccessRule{
		{Path: s.Path, Role: et.RoleUser, Permission: et.PermissionAll},
	})

	acOpts := et.AccessControllerOptions{
		RoleField:   "admin_role",
		SiteIDField: "admin_site_id",
	}

	grp := env.Rtr.Group(s.Path, et.AccessController(s.AcsMgr, s.log, acOpts))

	grp.POST("/accept", s.Save)
	grp.GET("/List/:id", s.List)

	return nil
}

// List ...
func (s *NewResidentRegistration) List(c echo.Context) (err error) {

	resp := utils.Response{}
	modelType := "NewResidentRegistrations"
	oid := c.Param("id")

	filter := utils.Options{}
	record, err := s.svc.GetBy(modelType, "id", oid, filter, "")

	if err != nil {
		s.log.Debug(err)
		resp.APIError(fmt.Errorf("bad request"))
		return c.JSON(http.StatusBadRequest, resp)
	}

	newRegResident := record.(*model.NewResidentRegistrations)

	filter["site_id"] = newRegResident.SiteID
	delete(filter, "$limit")

	list, err := s.svc.List("AvailableUnitsList", filter, "")

	if err != nil {
		return err
	}

	availableList := list.(*[]view.AvailableUnitsList)

	resp.Set("AvailableUnitsList", availableList)

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// Save ...
func (s *NewResidentRegistration) Save(c echo.Context) (err error) {

	resp := utils.Response{}
	form := form.AcceptNewRegistration{}
	modelType := "NewResidentRegistrations"

	if err = c.Bind(&form); err != nil {
		s.log.Error(err)
		return
	}

	// sanitize form
	if form.ID == "" || form.UnitID == "" {
		err := errors.New("cannot approve resident, invaid information")
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}

	filter := utils.Options{}
	record, err := s.svc.GetBy(modelType, "id", form.ID, filter, "")

	if err != nil {
		s.log.Debug(err)
		resp.APIError(fmt.Errorf("bad request"))
		return c.JSON(http.StatusBadRequest, resp)
	}

	newResident := record.(*model.NewResidentRegistrations)

	unitActiveQuery, err := s.svc.GetBy("Residency", "unit_id", form.UnitID, filter, "")

	if err != nil && err != pg.ErrNoRows {
		return err
	}

	unitActive := unitActiveQuery.(*model.Residency)

	// checking if the unit is active
	if unitActive.ID != "" && unitActive.ActiveStatus == model.ResidencyActive {
		err := errors.New("cannot accept resident, active resident on unit")
		return err
	}

	// beginning of transaction
	err = utils.Transact(s.env.Dbc, s.log, func(tx *pg.Tx) error {
		residencyID := xid.New().String()
		residency := model.Residency{
			ID:           residencyID,
			UnitID:       form.UnitID,
			DateStart:    utils.DateTime{}.Now(),
			SiteID:       newResident.SiteID,
			ActiveStatus: model.ResidencyActive,
		}

		_, err = tx.Model(&residency).Insert()

		if err != nil {
			log.Debug(err)
			return err
		}

		residentID := xid.New().String()

		password := utils.MakeRandText(10)
		passwordHash, err := utils.HashPassword(password)
		if err != nil {
			return err
		}

		// initializing the resident model with new resident
		resident := model.Resident{
			ID:          residentID,
			FirstName:   newResident.FirstName,
			LastName:    newResident.LastName,
			Attr:        newResident.Attr,
			Email:       newResident.Email,
			Password:    passwordHash,
			Phone:       newResident.Phone,
			Type:        model.PrimaryResident,
			Status:      model.IsEnabled,
			ResidencyID: residencyID,
			CanLogin:    true,
		}

		_, err = tx.Model(&resident).Insert()

		if err != nil {
			log.Debug(err)
			return err
		}

		// deleting the resident from the new resident table
		_, err = tx.Model(newResident).Where("id = ?", newResident.ID).Delete()

		if err != nil {
			log.Debug(err)
			return err
		}

		regEml, err := shared.NewResidents(tx, &resident, password)

		if err != nil {
			return err
		}

		// setting the email to the person being sent to
		_, err = tx.Exec(`insert into task_queue (site_id, type, data) values(?, 1, ?)
			`, newResident.SiteID, &regEml)

		if err != nil {
			log.Debug(err)
			return err
		}

		return nil
	})

	if err != nil {
		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Set("status", "ok")

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return nil
}
