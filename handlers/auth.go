package handlers

// cspell: ignore loggedin

import (
	"eve/service/model"
	"eve/service/view"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"strings"

	"net/http"

	"github.com/go-pg/pg"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthAPI ...
type AuthAPI struct {
	log  *zap.SugaredLogger
	env  *et.Env
	svc  utils.CRUDService
	Path string
}

// Initialize ...
func (s *AuthAPI) Initialize(env *et.Env) error {
	s.env = env
	s.log = env.Log.Sugar()

	svc := utils.CRUD{}
	err := svc.Init(s.env.Dbc, env.Log)
	if err != nil {
		s.log.Error(err)
		return err
	}
	s.svc = svc

	aMgr := et.NewAccessMgr()
	aMgr.AddRules([]et.AccessRule{
		{Path: utils.URLJoin(s.Path, "/login"), Role: et.RoleEveryone, Permission: et.PermissionReadWrite},
		{Path: utils.URLJoin(s.Path, "/logout"), Role: et.RoleEveryone, Permission: et.PermissionReadWrite},
		{Path: utils.URLJoin(s.Path, "/status"), Role: et.RoleEveryone, Permission: et.PermissionReadWrite},
		{Path: utils.URLJoin(s.Path, "/info"), Role: et.RoleEveryone, Permission: et.PermissionReadWrite},
	})

	acOpts := et.AccessControllerOptions{
		RoleField: "admin_role",
	}

	grp := env.Rtr.Group(s.Path, et.AccessController(aMgr, s.log, acOpts))
	grp.POST("/login/:subdomain", s.Login)
	grp.POST("/logout", s.Logout)
	grp.GET("/status", s.Status)
	grp.GET("/info/:subdomain", s.GetSiteInfo)

	return nil
}

// Login ...
func (s *AuthAPI) Login(c echo.Context) (err error) {

	frm := model.LoginForm{}
	if err = c.Bind(&frm); err != nil {
		s.log.Debug(err)
		return
	}

	frm.Email = strings.ToLower(frm.Email)

	resp := utils.Response{}

	//
	// get site record
	subdomain := c.Param("subdomain")
	retv, err := s.svc.Get("Site", "subdomain", subdomain, "")
	if err != nil {
		s.log.Debug(err)

		resp := utils.Response{}
		resp.SetErr("status", "unknown association")
		return c.JSON(http.StatusOK, resp)
	}
	if retv == nil {

		resp := utils.Response{}
		resp.SetErr("status", "unknown association")
		return c.JSON(http.StatusOK, resp)
	}

	site, _ := retv.(*model.Site)
	s.log.Debugf("%s - %s ", subdomain, site.ID)

	if model.Status(site.Status) == model.IsDisabled {
		resp.SetErr("status", "Your association has been deactivated! please contact your officials")
		return c.JSON(http.StatusOK, resp)
	}

	//
	// get user record
	opts := utils.Options{"site_id": site.ID}
	if !frm.IsResident {
		retv, err = s.svc.GetBy("User", "email", frm.Email, opts, "")
		if err != nil {
			if err == pg.ErrNoRows {
				resp.SetErr("status", "unknown user")
				return c.JSON(http.StatusOK, resp)
			}

			s.log.Debug(err)

			resp := utils.Response{}
			resp.APIError(fmt.Errorf("bad request"))
			return c.JSON(http.StatusBadRequest, resp)
		}
	} else {
		retv, err = s.getResident(c, site.ID, frm.Email, false)
		if err != nil {
			resp := utils.Response{}
			resp.APIError(err)
			return c.JSON(http.StatusBadRequest, resp)
		}
	}

	if retv == nil {
		resp.SetErr("status", "unknown user")
		return c.JSON(http.StatusOK, resp)
	}

	//
	// authenticate user
	user, _ := retv.(*model.User)

	// only run checks for residents
	if user.Type == model.ResidentUser {

		// disabled resident
		if user.Status == model.IsDisabled {
			resp.SetErr("status", "account is disabled cannot login")
			return c.JSON(http.StatusOK, resp)
		}
		// resident residency ended
		if user.ActiveStatus == model.ResidencyEnded {
			resp.SetErr("status", "this account is not active cannot login")
			return c.JSON(http.StatusOK, resp)
		}
	}

	// password authentication
	if utils.CheckPasswordHash(frm.Password, user.Password) == false {
		resp.SetErr("status", "invalid password")
		return c.JSON(http.StatusOK, resp)
	}

	// setup user session
	ses, err := et.NewSessionMgr(c, "", true)
	if err != nil {
		s.log.Error(err)
		return
	}

	if frm.IsResident && frm.IsMobile {
		ses.MaxAge(utils.OneYearINSeconds)
	}

	if frm.IsResident {
		s.log.Debugf("============ %+v", user)

		residentAddress, err := s.GetResidentAddress(c, user.ResidencyID)

		if err != nil {
			return err
		}

		user.Address = residentAddress
		user.ResidencyID = ""
	}

	ses.Set("admin_name", fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	ses.Set("admin_loggedin", true)
	ses.Set("admin_role", int(user.Role))
	ses.Set("admin_type", int(user.Type))
	ses.Set("admin_subtype", int(user.SubType))
	ses.Set("admin_id", user.ID)
	ses.Set("admin_site_id", user.SiteID)

	if err = ses.Save(); err != nil {
		s.log.Error(err)
		return
	}

	user.Password = "***"

	resp.Set("status", "login")
	resp.Set("user", user)
	resp.Set("site", site)

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// getResident retrieve a residents record and returns it as type User
func (s *AuthAPI) getResident(c echo.Context, siteID, value string, idFld bool) (interface{}, error) {
	opts := utils.Options{"site_id": siteID}
	field := "email"
	if idFld {
		field = "id"
	}

	retv, err := s.svc.GetBy("Resident", field, value, opts, "")
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, fmt.Errorf("unknown user")
		}

		s.log.Debug(err)
		return nil, fmt.Errorf("bad request")
	}

	res, _ := retv.(*model.Resident)
	user := &model.User{}
	if err := copier.Copy(user, res); err != nil {
		return nil, err
	}

	user.Role = int(et.RoleUser)
	user.Type = 3
	user.IsSiteUser = false

	if len(res.PrimaryID) > 0 {
		user.SubType = 1
	}

	var retIf interface{}
	retIf = user
	return retIf, nil
}

// Logout ...
func (s *AuthAPI) Logout(c echo.Context) (err error) {

	resp := utils.Response{}
	if err = s.ClearAdminSession(c); err != nil {
		return
	}

	resp.Set("status", "logout")
	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// Status ...
func (s *AuthAPI) Status(c echo.Context) (err error) {
	var retv interface{}

	resp := utils.Response{}
	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}

	if ses.Bool("admin_loggedin") {
		uid := ses.String("admin_id")
		if ses.Int("admin_type") != 3 {
			retv, err = s.svc.Get("User", "id", uid, "")
			if err != nil {
				s.log.Error(err)
				return err
			}
		} else {
			retv, err = s.getResident(c, ses.String("admin_site_id"), uid, true)
			if err != nil {
				s.log.Error(err)
				return err
			}
		}

		if retv == nil {
			// cant find the record for the logged in user, force logout
			return s.Logout(c)
		}

		user, _ := retv.(*model.User)
		user.Password = "***"

		retv, err = s.svc.Get("Site", "id", user.SiteID, "")
		if err != nil {
			s.log.Error(err)
			return err
		}
		site, _ := retv.(*model.Site)

		resp.Set("status", "login")
		resp.Set("user", user)
		resp.Set("site", site)

	} else {
		resp.Set("status", "logout")
	}

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// ClearAdminSession ...
func (s AuthAPI) ClearAdminSession(c echo.Context) (err error) {
	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}

	ses.Set("admin_loggedin", false)
	ses.Set("admin_role", int(et.RoleEveryone))
	ses.Set("admin_type", int(0))
	ses.Set("admin_subtype", int(0))
	ses.Set("admin_id", "")
	ses.Set("admin_site_id", "")

	if err = ses.Save(); err != nil {
		s.log.Debug(err)
		return
	}

	return
}

// GetSiteInfo ...
func (s AuthAPI) GetSiteInfo(c echo.Context) (err error) {

	//
	// get site record
	subdomain := c.Param("subdomain")
	retv, err := s.svc.Get("Site", "subdomain", subdomain, "")
	if err != nil {
		s.log.Debug(err)

		resp := utils.Response{}
		resp.SetErr("status", "unknown association")
		return c.JSON(http.StatusOK, resp)
	}
	site, _ := retv.(*model.Site)
	if site == nil {
		resp := utils.Response{}

		resp.Set("site", model.Site{})
		resp.Set("content", model.Content{})
		resp.Set("status", "NotFound")

		return c.JSON(http.StatusOK, resp)
	}

	retv, err = s.svc.Get("Content", "site_id", site.ID, "")
	if err != nil {
		s.log.Debug(err)

		resp := utils.Response{}
		resp.SetErr("status", "unknown association")
		return c.JSON(http.StatusOK, resp)
	}
	content, _ := retv.(*model.Content)

	resp := utils.Response{}
	resp.Set("site", site)
	resp.Set("content", content)
	resp.Set("status", "OK")

	return c.JSON(http.StatusOK, resp)
}

func getSiteID(c echo.Context) string {
	// siteID is set in echtotools/access.go

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		return ""
	}
	usrType := ses.Int("admin_type")

	// if the current user is a platform manager allow site_id overide
	if usrType == 6 {
		retv := c.QueryParam("_siteID_")
		if len(retv) > 0 {
			return retv
		}
	}

	// siteID is set in echtotools/access.go
	retv := c.Get("siteID")

	return retv.(string)
}

// ClearAdminSession ...
func (s AuthAPI) GetResidentAddress(c echo.Context, residencyID string) (string, error) {
	dbc := utils.Env.Db
	residency := model.Residency{}

	err := dbc.Model(&residency).Where("id = ?", residencyID).Select()

	if err != nil {
		return "", err
	}

	unitListView := view.UnitStreetView{}

	err = dbc.Model(&unitListView).Where("id = ?", residency.UnitID).Select()

	if err != nil {
		return "", err
	}

	return unitListView.Label, nil
}
