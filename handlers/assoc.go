package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"eve/service/form"
	"eve/service/model"
	"eve/service/view"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Associations ...
type Associations struct {
	log  *zap.SugaredLogger
	env  *et.Env
	svc  utils.CRUDService
	Path string

	AcsMgr *et.AccessMgr
}

// Initialize ...
func (s *Associations) Initialize(env *et.Env) error {
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
		{Path: s.Path, Role: et.RoleEveryone, Permission: et.PermissionAll},
	})

	acOpts := et.AccessControllerOptions{
		RoleField:   "admin_role",
		SiteIDField: "admin_site_id",
	}

	grp := env.Rtr.Group(s.Path, et.AccessController(s.AcsMgr, s.log, acOpts))

	grp.GET("", s.List)
	grp.GET("/:id", s.Get)
	grp.POST("/enable_support", s.EnableSupport)
	grp.POST("", s.Save)
	grp.POST("/:id", s.Save)
	grp.DELETE("/:id", s.Delete)

	return nil
}

// Get ...
func (s *Associations) Get(c echo.Context) (err error) {

	resp := utils.Response{}
	modelType := "AssociationView"
	oid := c.Param("id")
	// siteID := getSiteID(c)

	filter := utils.Options{}
	// filter["site_id"] = siteID
	record, err := s.svc.GetBy(modelType, "id", oid, filter, "")
	if err != nil {
		s.log.Debug(err)

		resp := utils.Response{}
		resp.APIError(fmt.Errorf("bad request"))
		return c.JSON(http.StatusBadRequest, resp)
	}

	if record == nil {
		resp := utils.Response{}
		resp.APIError(fmt.Errorf("record not found"))
		return c.JSON(http.StatusNotFound, resp)
	}

	resp.Set("record", record)

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

func (s *Associations) EnableSupport(c echo.Context) (err error) {
	log := utils.Env.Log
	dbc := utils.Env.Db
	response := utils.Response{}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}

	role := ses.String("admin_type")

	if role != fmt.Sprint(role) {
		return errors.New("Invalid User")
	}

	frm := form.SupportAccountForm{}
	if err = c.Bind(&frm); err != nil {
		s.log.Debug(err)
		return
	}

	User := model.User{}

	err = dbc.Model(&User).
		Where("support_account = ?", true).
		Where("type = ?", model.SupportUser).
		Where("site_id = ?", frm.SiteID).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return errors.New("Support Account not for association not present")
		}
		log.Debug(err)
		return err
	}

	password := utils.MakeRandText(7)
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	User.Password = passwordHash
	User.Status = model.IsEnabled

	log.Debug("password", password)
	log.Debug("hash", passwordHash)

	_, err = dbc.Model(&User).Set("password =?password, status=?status").
		Where("id = ?id").Update()
	if err != nil {
		return err
	}

	data := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		User.Email,
		password,
	}
	response.Set("status", true)
	response.Set("data", data)

	if err = c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return
	}
	return nil
}

// List ...
func (s *Associations) List(c echo.Context) (err error) {
	resp := utils.Response{}
	filter := c.QueryParam("_filter")

	err = s.getList(filter, "", &resp)
	if err != nil {
		resp := utils.Response{}

		resp.APIError(fmt.Errorf("bad request"))
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

func (s *Associations) getList(filter, name string, resp *utils.Response) (err error) {
	modelType := "AssociationView"
	if len(name) == 0 {
		name = "list"
	}

	opts := utils.Options{}
	opts.Parse(filter, ",")

	records, count, err := s.svc.ListAndCount(modelType, opts, "")
	if err != nil {
		s.log.Error(err)

		return err
	}

	resp.Set(name, records)
	resp.Set("count", count)

	return nil
}

// Save insert a new record or update an existing one
func (s *Associations) Save(c echo.Context) (err error) {

	resp := utils.Response{}
	oid := c.Param("id")

	frm := &view.AssociationView{}
	site := &model.Site{}
	user := &model.User{}

	if err = c.Bind(frm); err != nil {
		s.log.Error(err)
		return
	}

	site.ID = frm.ID
	site.Name = frm.Name
	site.Subdomain = frm.Subdomain
	site.Status = frm.Status
	site.DateRegistered = utils.NewDateTime(time.Now())
	site.Attr = frm.Attr

	user.ID = frm.AdminID
	user.SiteID = site.ID
	user.FirstName = frm.AdminFirstName
	user.LastName = frm.AdminLastName
	user.Email = frm.AdminEmail
	user.Phone = frm.AdminPhone
	user.Password = frm.AdminPassword
	user.Status = 1
	user.Role = int(et.RoleSuperUser)
	user.Type = 5 // admin official
	user.IsSiteUser = true

	err = utils.Transact(s.env.Dbc, s.log, func(tx *pg.Tx) error {

		if len(oid) == 0 || oid == "new" {
			site.Status = 1
			// create site record
			if err = s.svc.Create(tx, "Site", site, false); err != nil {
				s.log.Debug(err)
				return err
			}

			// create user record for the site
			user.SiteID = site.ID
			if err = s.svc.Create(tx, "User", user, false); err != nil {
				s.log.Debug(err)
				return err
			}

			resp.Set("id", site.ID)
		} else {
			excludedFields := []string{"ID", "DateRegistered"}

			if err = s.svc.Save(tx, "Site", site, excludedFields); err != nil {
				s.log.Debug(err)
				return err
			}

			excludedFields = []string{"ID", "SiteID", "FirstName", "LastName", "Email", "Phone", "Status", "Attr", "Role", "Type", "IsSiteUser"}
			if err = s.svc.Save(tx, "User", user, excludedFields); err != nil {
				s.log.Debug(err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}

	filter := c.QueryParam("_list")
	err = s.getList(filter, "association", &resp)
	if err != nil {
		resp := utils.Response{}

		resp.APIError(fmt.Errorf("bad request"))
		return c.JSON(http.StatusBadRequest, resp)
	}

	resp.Set("status", "ok")

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// Delete ...
func (s *Associations) Delete(c echo.Context) (err error) {

	resp := utils.Response{}
	oid := c.Param("id")

	// if site has streets registered refuse to delete site
	count, err := s.env.Dbc.Model((*model.Street)(nil)).Where("site_id = ?", oid).Count()
	if err != nil {
		s.log.Debug(err)
		return err
	}
	if count > 0 {
		err = fmt.Errorf("cant delete: association is active")

		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = s.svc.Delete(nil, "Site", oid); err != nil {
		s.log.Debug(err)
		return err
	}

	filter := c.QueryParam("_list")
	err = s.getList(filter, "association", &resp)
	if err != nil {
		resp := utils.Response{}

		resp.APIError(fmt.Errorf("bad request"))
		return c.JSON(http.StatusBadRequest, resp)
	}

	resp.Set("status", "ok")

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}
