// File: multicrud.go
// File Created: Friday, 19th July 2019 6:12:20 am
// Author: Akinmayowa Akinyemi
// -----
// Copyright 2019 Techne Efx Ltd
//
// A version of MultiCRUD that scopes requests to a site

package echotools

import (
	"fmt"
	"net/http"
	"strings"

	"eve/utils"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// CrudAPIInstance ...
var CrudAPIInstance *CrudAPI

// CrudAPI ...
type CrudAPI struct {
	log  *zap.SugaredLogger
	srv  *Env
	svc  utils.CRUDService
	Path string
	// Entities a list of types for which this api will respond
	// Each entry in this table includes map of other tables that can
	// be accessed in relation to this entity
	Models []ModelInfo

	AcsMgr *AccessMgr
}

// ModelInfo ...
type ModelInfo struct {
	Type          string
	Exclude       string
	TableName     string
	NoSiteID      bool
	MinAccessType int

	BeforeReadHook CrudBeforeReadHook
	AfterReadHook  CrudAfterReadHook
	BeforeListHook CrudBeforeListHook
	AfterListHook  CrudAfterListHook
	BeforeSaveHook CrudSaveHook
	AfterSaveHook  CrudSaveHook
	DeleteHook     CrudDeleteHook
}

// CrudBeforeReadHook ...
type CrudBeforeReadHook func(c echo.Context, mi *ModelInfo, field, value string, filter *utils.Options, resp *utils.Response) (bool, error)

// CrudAfterReadHook ...
type CrudAfterReadHook func(c echo.Context, mi *ModelInfo, frm interface{}, resp *utils.Response) error

// CrudBeforeListHook ...
type CrudBeforeListHook func(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error)

// CrudAfterListHook ...
type CrudAfterListHook func(c echo.Context, records interface{}, resp *utils.Response) error

// CrudSaveHook ...
type CrudSaveHook func(tx *pg.Tx, c echo.Context, mi *ModelInfo, frm interface{}, resp *utils.Response) (bool, error)

// CrudDeleteHook ...
type CrudDeleteHook func(tx *pg.Tx, c echo.Context, mi *ModelInfo, resp *utils.Response) (bool, error)

// Initialize ...
func (s *CrudAPI) Initialize(srv *Env) error {

	if CrudAPIInstance == nil {
		CrudAPIInstance = s
	}

	s.srv = srv
	s.log = srv.Log.Sugar()
	svc := utils.CRUD{}
	err := svc.Init(s.srv.Dbc, srv.Log)
	if err != nil {
		s.log.Error(err)
		return err
	}
	s.svc = svc

	if s.AcsMgr == nil {
		s.AcsMgr = NewAccessMgr()
		s.AcsMgr.AddRules([]AccessRule{
			{utils.URLJoin(s.Path, "/"), RoleEveryone, PermissionAll},
		})
	}

	acOpts := AccessControllerOptions{
		RoleField:   "admin_role",
		SiteIDField: "admin_site_id",
	}

	grp := srv.Rtr.Group(s.Path, AccessController(s.AcsMgr, s.log, acOpts))

	grp.GET("/multi/:list", s.GetMulti)
	grp.GET("/:model", s.List)
	grp.GET("/:model/:id", s.Get)
	grp.GET("/:model/:field/:value", s.GetByField)
	grp.POST("/:model", s.Save)
	grp.POST("/:model/:id", s.Save)
	grp.DELETE("/:model/:id", s.Delete)

	return nil
}

func getSiteID(c echo.Context) string {
	ses, err := NewSessionMgr(c, "")
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
	if retv != nil {
		return retv.(string)
	}

	return "unknown"
}

func (s CrudAPI) findModel(name string, usrType int) *ModelInfo {

	for i := range s.Models {
		if s.Models[i].Type == name && usrType >= s.Models[i].MinAccessType {
			return &(s.Models[i])
		}
	}

	return nil
}

// Get ...
func (s *CrudAPI) Get(c echo.Context) (err error) {
	ses, err := NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")

	resp := utils.Response{}
	modelType := strings.Title(c.Param("model"))
	oid := c.Param("id")
	siteID := getSiteID(c)

	s.log.Debug("model: ", c.Param("model"), " - id: ", c.Param("id"), " - siteID: ", getSiteID(c))

	// get the model and associated attributes
	var model *ModelInfo
	if model = s.findModel(modelType, usrType); model == nil {
		err := fmt.Errorf("unknown entity: %s", modelType)
		s.log.Debug(err)

		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}

	// call the entity service to get data for this entity
	filter := utils.Options{}
	if len(siteID) > 0 && model.NoSiteID == false {
		filter["site_id"] = siteID
	}

	stop := false
	if model.BeforeReadHook != nil {
		stop, err = model.BeforeReadHook(c, model, "id", oid, &filter, &resp)
		if err != nil {
			return err
		}
	}

	if !stop {
		record, err := s.svc.GetBy(modelType, "id", oid, filter, model.TableName)
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

		// if record includes a password field replace it with '***
		if utils.StructHasField(record, "Password") {
			utils.SetStructField(record, "Password", "***")
		}

		// call Read hook if provided
		if model.AfterReadHook != nil {
			err = model.AfterReadHook(c, model, record, &resp)
			if err != nil {
				return err
			}
		}

		resp.Set("record", record)
	}

	// _list=category|comments:1234
	items := c.QueryParam("_list")
	if err = s.GetItems(c, &resp, siteID, items, usrType); err != nil {
		s.log.Error(err)
		return
	}

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// GetByField ...
func (s *CrudAPI) GetByField(c echo.Context) (err error) {
	ses, err := NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")

	resp := utils.Response{}
	modelType := strings.Title(c.Param("model"))
	siteID := getSiteID(c)
	field := c.Param("field")
	value := c.Param("value")

	// get the model and associated attributes
	var model *ModelInfo
	if model = s.findModel(modelType, usrType); model == nil {
		err := fmt.Errorf("unknown entity: %s", modelType)
		s.log.Debug(err)

		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}

	// call the entity service to get data for this entity
	filter := utils.Options{}
	if len(siteID) > 0 && model.NoSiteID == false {
		filter["site_id"] = siteID
	}

	stop := false
	if model.BeforeReadHook != nil {
		stop, err = model.BeforeReadHook(c, model, field, value, &filter, &resp)
		if err != nil {
			return err
		}
	}

	if !stop {
		record, err := s.svc.GetBy(modelType, field, value, filter, model.TableName)
		if err != nil {
			s.log.Debug(err)

			resp := utils.Response{}
			if err == pg.ErrNoRows {
				resp.APIError(fmt.Errorf("record not found"))
				return c.JSON(http.StatusNotFound, resp)
			}

			resp.APIError(fmt.Errorf("bad request"))
			return c.JSON(http.StatusBadRequest, resp)
		}

		if record == nil {
			resp := utils.Response{}
			resp.APIError(fmt.Errorf("record not found"))
			return c.JSON(http.StatusNotFound, resp)
		}

		// if record includes a password field replace it with '***
		if utils.StructHasField(record, "Password") {
			utils.SetStructField(record, "Password", "***")
		}

		// call Read hook if provided
		if model.AfterReadHook != nil {
			err = model.AfterReadHook(c, model, record, &resp)
			if err != nil {
				return err
			}
		}

		resp.Set("record", record)
	}
	// _list=category|comments:1234
	items := c.QueryParam("_list")
	if err = s.GetItems(c, &resp, siteID, items, usrType); err != nil {
		s.log.Error(err)
		return
	}

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// List query registered models as follows
// GET /model?_filter=sex:1,age:>25,$order:age
//  --> where sex = 1 age > 25 order by age
func (s *CrudAPI) List(c echo.Context) (err error) {
	ses, err := NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")

	resp := utils.Response{}
	modelType := strings.Title(c.Param("model"))
	siteID := getSiteID(c)

	var model *ModelInfo
	// check if entity is in Entities list
	if model = s.findModel(modelType, usrType); model == nil {
		err := fmt.Errorf("unknown entity: %s", modelType)
		s.log.Debug(err)

		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}
	// s.log.Debug("model: ", model)

	// parse _filter query param into Options
	filter := c.QueryParam("_filter")
	// s.log.Debug("list filters: ", filter)
	opts := utils.Options{}
	opts.Parse(filter, ",")
	if len(siteID) > 0 && model.NoSiteID == false {
		opts["site_id"] = siteID
	}

	stop := false
	if model.BeforeListHook != nil {
		stop, err = model.BeforeListHook(c, &opts, &resp)
		if err != nil {

			resp := utils.Response{}
			resp.APIError(fmt.Errorf("bad request %s", err))
			return c.JSON(http.StatusBadRequest, resp)
		}
	}

	s.log.Debug("parsed filters: ", opts)

	if !stop {
		records, count, err := s.svc.ListAndCount(model.Type, opts, model.TableName)
		if err != nil {
			s.log.Error(err)

			resp := utils.Response{}
			resp.APIError(fmt.Errorf("bad request"))
			return c.JSON(http.StatusBadRequest, resp)
		}

		// call List hook if provided
		if model.AfterListHook != nil {
			err = model.AfterListHook(c, records, &resp)
			if err != nil {
				return err
			}
		}

		resp.Set("list", records)
		resp.Set("count", count)
	}

	// return additional data
	// _list=category|comments:1234
	items := c.QueryParam("_list")
	if err = s.GetItems(c, &resp, siteID, items, usrType); err != nil {
		s.log.Debug(err)

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

// GetMulti ...
func (s *CrudAPI) GetMulti(c echo.Context) (err error) {
	ses, err := NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")

	resp := utils.Response{}
	siteID := getSiteID(c)

	// return additional data
	// list=category|comments:1234
	items := c.Param("list")
	if err = s.GetItems(c, &resp, siteID, items, usrType); err != nil {
		s.log.Debug(err)

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

// Save insert a new record or update an existing one
func (s *CrudAPI) Save(c echo.Context) (err error) {
	ses, err := NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")

	resp := utils.Response{}
	modelType := strings.Title(c.Param("model"))

	fmt.Println("Model ==> ", modelType)

	oid := c.Param("id")
	siteID := getSiteID(c)

	var model *ModelInfo
	// check if entity is in Entities list
	if model = s.findModel(modelType, usrType); model == nil {
		err := fmt.Errorf("unknown entity: %s", modelType)
		s.log.Debug(err)

		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}

	// formType := fmt.Sprintf("frm%s", modelType)
	excludedFields := strings.Split(model.Exclude, ",")

	frm, err := utils.MakePointerType(modelType)
	if err != nil {
		s.log.Debug(err)
		return
	}

	if err = c.Bind(frm); err != nil {
		s.log.Error(err)
		return
	}

	err = utils.Transact(s.srv.Dbc, s.log, func(tx *pg.Tx) error {
		var (
			stop bool
			err  error
		)
		if model.BeforeSaveHook != nil {
			stop, err = model.BeforeSaveHook(tx, c, model, frm, &resp)
			if err != nil {
				s.log.Debug(err)

				if len(resp.Error) > 0 {
					// error has already been set in hook
					return err
				}

				resp := utils.Response{}
				resp.APIError(err)
				c.JSON(http.StatusInternalServerError, resp)
				return err
			}
		}

		if !stop {
			if len(oid) == 0 || oid == "new" {
				// set siteID if available and the struct has it
				if len(siteID) > 0 && utils.StructHasField(frm, "SiteID") {
					utils.SetStructField(frm, "SiteID", siteID)
				}

				if err = s.svc.Create(tx, modelType, frm, false); err != nil {
					s.log.Debug(err)

					resp := utils.Response{}
					resp.APIError(err)
					c.JSON(http.StatusBadRequest, resp)
					return err
				}

				newID := utils.GetStructField(frm, "ID")
				resp.Set("id", newID.String())
			} else {
				if err = s.svc.Save(tx, modelType, frm, excludedFields); err != nil {
					s.log.Debug(err)

					resp := utils.Response{}
					resp.APIError(err)
					c.JSON(http.StatusBadRequest, resp)
					return err
				}
			}

			if model.AfterSaveHook != nil {
				_, err = model.AfterSaveHook(tx, c, model, frm, &resp)
				if err != nil {
					s.log.Debug(err)

					resp := utils.Response{}
					resp.APIError(err)
					c.JSON(http.StatusInternalServerError, resp)
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil
	}

	// return additional data
	// _list=category|comments:1234
	items := c.QueryParam("_list")
	if err = s.GetItems(c, &resp, siteID, items, usrType); err != nil {
		s.log.Debug(err)

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
func (s *CrudAPI) Delete(c echo.Context) (err error) {
	ses, err := NewSessionMgr(c, "")
	if err != nil {
		s.log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")

	resp := utils.Response{}
	modelType := strings.Title(c.Param("model"))
	oid := c.Param("id")
	siteID := getSiteID(c)

	var model *ModelInfo
	// check if entity is in Entities list
	if model = s.findModel(modelType, usrType); model == nil {
		err := fmt.Errorf("unknown entity: %s", modelType)
		s.log.Warn(err)

		resp := utils.Response{}
		resp.APIError(err)
		return c.JSON(http.StatusBadRequest, resp)
	}

	sqlFn := func(tx *pg.Tx) error {
		if err = s.svc.Delete(tx, modelType, oid); err != nil {
			s.log.Debug(err)

			resp := utils.Response{}
			resp.APIError(fmt.Errorf("internal server error"))
			c.JSON(http.StatusInternalServerError, resp)
			return err
		}

		return nil
	}

	if model.DeleteHook == nil {
		if err = sqlFn(nil); err != nil {
			return
		}
	} else {
		err = utils.Transact(s.srv.Dbc, s.log, func(tx *pg.Tx) error {
			stop, err := model.DeleteHook(tx, c, model, &resp)
			if err != nil {

				if len(resp.Error) > 0 {
					// error has already been set in hook
					return err
				}
				resp := utils.Response{}
				resp.APIError(err)
				c.JSON(http.StatusInternalServerError, resp)
				return err
			}

			if !stop {
				if err := sqlFn(tx); err != nil {
					resp.APIError(err)

					return err
				}
			}

			return nil
		})
		if err != nil {
			return
		}
	}

	// return additional data
	// _list=category|comments:1234
	items := c.QueryParam("_list")
	if err = s.GetItems(c, &resp, siteID, items, usrType); err != nil {
		s.log.Debug(err)

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

// GetItems ...
func (s CrudAPI) GetItems(c echo.Context, resp *utils.Response, siteID, itemList string, usrType int) (err error) {

	// roles|categories-type:>1234
	// people-type:>1,age:<22
	// each model must be pre defined in s.Model where additional info is located
	//
	// format:
	// type-clause, ...|type-clause,...|...
	//
	// clause format: field:[op]<value>,...
	//
	// e.g category:>2
	// category:>=2,age:<30
	//
	// operators: >. >=, < ,<=

	// split itemList into separate items)
	items := strings.Split(strings.TrimSpace(itemList), "|")

	for _, i := range items {
		if len(i) < 1 {
			continue
		}

		parts := strings.SplitN(i, "-", 2)
		typeName := parts[0]

		// find model infomation for this type
		info := s.findModel(strings.Title(typeName), usrType)
		if info == nil {
			// requested type not found in information list
			s.log.Debugf("unknown type: (%s)", typeName)
			continue
		}

		val := ""
		if len(parts) > 1 {
			val = parts[1]
		}

		filter := utils.Options{}
		filter.Parse(val, ",")

		// query db for data
		if len(siteID) > 0 && info.NoSiteID == false {
			filter["site_id"] = siteID
		}

		stopped := false
		if info.BeforeListHook != nil {
			stopped, err = info.BeforeListHook(c, &filter, resp)
			if err != nil {
				return err
			}
		}

		if !stopped {
			retv, err := s.svc.List(info.Type, filter, info.TableName)
			if err != nil {
				s.log.Debug(err)
				return err
			}

			if info.AfterListHook != nil {
				err = info.AfterListHook(c, retv, resp)
				if err != nil {
					return err
				}
			}

			// store in response
			resp.Set(typeName, retv)
		}
	}

	return
}
