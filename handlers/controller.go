package handlers

import (
	"bytes"
	"encoding/json"
	"eve/service/form"
	"eve/service/model"
	"eve/service/view"
	"eve/shared"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Controller ...
type Controller struct {
	log  *zap.SugaredLogger
	env  *et.Env
	svc  utils.CRUDService
	Path string

	AcsMgr *et.AccessMgr
}

// Initialize ...
func (s *Controller) Initialize(env *et.Env) error {
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
		// {Path: s.Path, Role: et.RoleUser, Permission: et.PermissionAll},
		{Path: s.Path, Role: et.RoleEveryone, Permission: et.PermissionAll},
		{Path: s.Path, Role: et.RoleEveryone, Permission: et.PermissionAll},
		{Path: utils.URLJoin(s.Path, "/paystack"), Role: et.RoleUser, Permission: et.PermissionReadWrite},
	})

	acOpts := et.AccessControllerOptions{
		RoleField:   "admin_role",
		SiteIDField: "admin_site_id",
	}

	grp := env.Rtr.Group(s.Path, et.AccessController(s.AcsMgr, s.log, acOpts))

	grp.GET("/residentDues/:id", s.GetResidentDues)
	grp.GET("/association/site_code/:id", s.GetAssociationByCode)
	grp.POST("/register_assoc", s.RegisterAssoc)
	grp.POST("/register_newresident", s.RegisterNewResident)
	grp.GET("/tpl/:id", s.GetTpl)
	grp.GET("/pub/:id", s.GetPublicKey)
	grp.GET("/pvdrs/:id", s.GetProviders)
	grp.POST("/email", s.VerifyEmail)
	grp.POST("/paystack", s.VerifyPaystack)

	return nil
}

func (s *Controller) GetProviders(c echo.Context) (err error) {
	dbc := utils.Env.Db
	response := utils.Response{}

	param := c.Param("id")
	ID, _ := strconv.Atoi(param)

	name := ""
	if ID == model.ResidentUser {
		name = "ResidentProvider"
	} else {
		name = "AssociationProvider"
	}

	ProviderSection := utils.GetConfigList("", name)

	ProvidersList := []model.PaymentProviders{}
	list := []model.PaymentProviders{}

	if err := dbc.Model(&list).Select(); err != nil {
		return err
	}

	for _, list := range list {
		for _, item := range ProviderSection {
			if item == list.Name {
				ProvidersList = append(ProvidersList, list)
			}
		}
	}

	response.Set("list", ProvidersList)

	if err = c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return
	}
	return nil
}

func (s *Controller) VerifyEmail(c echo.Context) (err error) {
	dbc := s.env.Dbc
	response := utils.Response{}

	form := form.VerifyEmail{}
	if err := c.Bind(&form); err != nil {
		return err
	}
	form.Email = strings.ToLower(form.Email)

	newReg := &model.NewResidentRegistrations{}
	res := &view.ResidentFamilyList{}

	if err := dbc.Model(res).Where("email = ?", form.Email, form.SiteID).Select(); err != nil && err != pg.ErrNoRows {
		return err
	}

	if err := dbc.Model(newReg).Where("email = ?", form.Email).Select(); err != nil && err != pg.ErrNoRows {
		return err
	}

	if res.Email == "" && newReg.Email == "" {
		response.Set("valid", true)
	} else {
		response.Set("valid", false)
	}

	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil
}

// GetPublicKey ...
func (s *Controller) GetPublicKey(c echo.Context) (err error) {
	response := utils.Response{}

	param := c.Param("id")
	ID, _ := strconv.Atoi(param)

	provider := model.GetProvider(model.PayProvider(ID))

	ProviderSection, err := utils.Getkey("", "ProviderSection")
	if err != nil {
		return err
	}

	publicKey, err := utils.Getkey(ProviderSection, fmt.Sprintf("%s_public", provider.Name))
	if err != nil {
		return err
	}

	if ID == int(model.ProviderFlutterwave) {
		encryptionKey, _ := utils.Getkey(ProviderSection, fmt.Sprintf("%s_encryption_key", provider.Name))
		if err != nil {
			return err
		}
		response.Set("encryption_key", encryptionKey)
	}

	response.Set("public_key", publicKey)

	if err = c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return
	}

	return
}

// GetTpl ...
func (s *Controller) GetTpl(c echo.Context) (err error) {
	dbc := utils.Env.Db
	log := utils.Env.Log

	rid := c.Param("id")
	typ := c.QueryParam("typ")
	var eml *shared.EMailMsg

	utils.Transact(dbc, log, func(tx *pg.Tx) error {
		if typ == "invoice" {
			eml, err = shared.MakeInvoice(tx, rid)
			if err != nil {
				return err
			}
		} else if typ == "receipt" {
			eml, err = shared.MakeReceipt(tx, rid)
			if err != nil {
				return err
			}
		} else {
			eml = &shared.EMailMsg{
				HTML: "<h1>url parameter &typ=\"???\" not provided</h1>",
			}
		}

		return nil
	})

	c.HTML(200, eml.HTML)
	return nil
}

func (s *Controller) VerifyPaystack(c echo.Context) (err error) {
	log := utils.Env.Log
	response := utils.Response{}

	form := model.InitializePaystackRequest{}

	if err := c.Bind(&form); err != nil {
		return err
	}

	field := fmt.Sprintf("%s_secret", "paystack")

	secretKey, err := utils.Getkey("testpayments", field)
	if err != nil {
		return err
	}

	// log.Debug("secret Key", secretKey)
	// log.Debug("data", )

	byteData, err := json.Marshal(form)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.paystack.co/transaction/initialize", bytes.NewBuffer(byteData))
	if err != nil {
		return
	}

	// setting the header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKey))

	client := &http.Client{}

	// send request
	resp, err := client.Do(req)

	if err != nil {
		return
	}

	var byteRes []byte

	// read response as byte
	byteRes, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return

	}

	log.Debug("got response from paystack")

	paystackRes := model.InitializePaystackResponse{}

	err = json.Unmarshal(byteRes, &paystackRes)

	if err != nil {
		return err
	}

	response.SetStore(paystackRes.Data)

	if err = c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return
	}

	return
}
