package handlers

import (
	"eve/service/form"
	"eve/service/model"
	"eve/service/view"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-pg/pg"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

// Controller ...
type ResidentUtil struct {
	log  *zap.SugaredLogger
	env  *et.Env
	svc  utils.CRUDService
	Path string

	AcsMgr *et.AccessMgr
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

// Initialize ...
func (s *ResidentUtil) Initialize(env *et.Env) error {
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
		{Path: s.Path, Role: et.RoleEveryone, Permission: et.PermissionReadWrite},
	})

	acOpts := et.AccessControllerOptions{
		RoleField:   "admin_role",
		SiteIDField: "admin_site_id",
	}

	grp := env.Rtr.Group(s.Path, et.AccessController(s.AcsMgr, s.log, acOpts))
	grp.GET("/ws", s.InitWebsocket)
	grp.GET("/dashboard", s.ResidentDashboard)
	grp.POST("/alert", s.SaveResidentAlert)
	grp.POST("/alert/:id", s.UpdateResidentAlert)
	grp.GET("/list_alerts", s.ListResidentAlert)
	grp.GET("/alert_count", s.CountAlertUsingStatus)
	grp.GET("/get_last_alert", s.GetLastResidentAlert)
	grp.POST("/update_push_notification", s.UpdatePushNotification)

	return nil

}

func (s *ResidentUtil) UpdatePushNotification(c echo.Context) error {
	dbc := s.env.Dbc
	log := s.log
	response := utils.Response{}

	form := form.UpdatePushNotification{}
	if err := c.Bind(&form); err != nil {
		return err
	}

	s.log.Debug("=========== form %+v", form)

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}

	siteID := ses.String("admin_site_id")

	if form.UserType == model.ResidentUser {

		residentUser := model.Resident{}

		err = dbc.Model(&residentUser).Where("id = ?", form.UserID, siteID).Select()
		if err != nil && err != pg.ErrNoRows {
			return err
		}

		if residentUser.ID == "" {
			if err := c.JSON(http.StatusBadRequest, response); err != nil {
				s.log.Error(err)
				return nil
			}
			return nil
		}

		resident := model.Resident{
			ID:        form.UserID,
			PushToken: form.Token,
		}
		_, err = dbc.Model(&resident).Set("push_token=?push_token").
			Where("id = ?id").Update()

		if err != nil {
			return err
		}

	}

	if form.UserType == model.SecurityUser {

		securityUser := model.User{}

		err = dbc.Model(&securityUser).Where("id = ? and site_id = ? and type = ?", form.UserID, siteID, model.SecurityUser).Select()
		if err != nil && err != pg.ErrNoRows {
			return err
		}

		if securityUser.ID == "" {
			if err := c.JSON(http.StatusBadRequest, response); err != nil {
				s.log.Error(err)
				return nil
			}
			return nil
		}

		user := model.User{
			ID:        form.UserID,
			PushToken: form.Token,
		}
		_, err = dbc.Model(&user).Set("push_token=?push_token").
			Where("id = ?id").Update()

		if err != nil {
			return err
		}
	}

	response.Set("status", "updated")
	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil

}

func (s *ResidentUtil) CountAlertUsingStatus(c echo.Context) error {
	dbc := s.env.Dbc
	log := s.log
	response := utils.Response{}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}
	siteID := ses.String("admin_site_id")

	model := model.ResidentAlerts{}

	queryParam := c.QueryParam("status")

	var query string
	if queryParam != "" {
		status, err := strconv.Atoi(queryParam)
		if err != nil {
			return err
		}
		query = fmt.Sprintf("site_id = '%s' and status = %d", siteID, status)
	} else {
		query = fmt.Sprintf("site_id = '%s'", siteID)
	}

	count, err := dbc.Model(&model).Where(query).Count()
	if err != nil {
		return err
	}

	response.Set("count", count)
	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil

}

func (s *ResidentUtil) ListResidentAlert(c echo.Context) error {

	dbc := s.env.Dbc
	log := s.log
	response := utils.Response{}

	list := []model.ResidentAlerts{}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}
	siteID := ses.String("admin_site_id")

	queryParam := c.QueryParam("status")

	var query string
	if queryParam != "" {
		status, err := strconv.Atoi(queryParam)
		if err != nil {
			return err
		}
		query = fmt.Sprintf("site_id = '%s' and status = %d", siteID, status)
	} else {
		query = fmt.Sprintf("site_id = '%s'", siteID)
	}

	if err := dbc.Model(&list).Where(query).Order("time_logged desc").Select(); err != nil {
		return err
	}

	response.Set("list", list)
	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil
}

func (s *ResidentUtil) GetLastResidentAlert(c echo.Context) error {
	dbc := s.env.Dbc
	log := s.log
	response := utils.Response{}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}

	siteID := ses.String("admin_site_id")
	usrID := ses.String("admin_id")

	residentAlert := model.ResidentAlerts{}

	err = dbc.Model(&residentAlert).
		Where("resident_id = ? and site_id = ? ", usrID, siteID).
		Order("time_logged desc").
		Limit(1).Select()
	if err != nil && err != pg.ErrNoRows {
		s.log.Debug("====== err", err)
		return err
	}

	if residentAlert.ID == "" {
		response.Set("data", "no alert present")
		if err := c.JSON(http.StatusOK, response); err != nil {
			s.log.Error(err)
			return nil
		}
		return nil
	}

	response.Set("data", residentAlert)
	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil
}

func (s *ResidentUtil) UpdateResidentAlert(c echo.Context) error {
	dbc := s.env.Dbc
	log := s.log
	response := utils.Response{}

	form := form.ResidentAlertForm{}
	if err := c.Bind(&form); err != nil {
		return err
	}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}

	siteID := ses.String("admin_site_id")

	ID := c.Param("id")
	residentAlert := model.ResidentAlerts{}

	err = dbc.Model(&residentAlert).Where("id = ? and site_id = ?", ID, siteID).Select()
	if err != nil {
		return err
	}

	if model.ResidentAlert(form.Status) <= model.ResidentAlert(model.AlertCompleted) {
		residentAlert.Status = model.ResidentAlert(form.Status)
		residentAlert.Attr = form.Attr
	}

	_, err = dbc.Model(&residentAlert).Set("status=?status, attr=?attr").
		Where("id = ?id").Update()

	response.Set("message", "alert status updated")
	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil
}

func (s *ResidentUtil) SaveResidentAlert(c echo.Context) error {
	dbc := s.env.Dbc
	log := s.log
	response := utils.Response{}

	form := form.ResidentAlertForm{}

	if err := c.Bind(&form); err != nil {
		return err
	}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}

	usrID := ses.String("admin_id")
	siteID := ses.String("admin_site_id")

	currentTime := utils.NewDateTime(time.Now())

	residentAlertCheck := []model.ResidentAlerts{}
	err = dbc.Model(&residentAlertCheck).Where("resident_id = ? and site_id = ? and status = ?", usrID, siteID, model.AlertSent).Select()
	if err != nil {
		return err
	}

	if len(residentAlertCheck) > 0 {
		response.Set("message", "alert has already been raised by user")

		if err := c.JSON(http.StatusOK, response); err != nil {
			s.log.Error(err)
			return err
		}
		return nil
	}

	resident := view.ResidentView{}
	err = dbc.Model(&resident).Where("id = ? and site_id = ?", usrID, siteID).Limit(1).Select()
	if err != nil {
		return err
	}

	residentAlert := model.ResidentAlerts{
		ID:            xid.New().String(),
		SiteID:        siteID,
		ResidentID:    usrID,
		TimeLogged:    currentTime,
		TimeResponded: currentTime,
		Status:        model.ResidentAlert(form.Status),
		Attr:          form.Attr,
		Name:          resident.Name,
		Address:       resident.Unit,
		PhoneNumber:   resident.Phone,
	}

	_, err = dbc.Model(&residentAlert).Insert()
	if err != nil {
		return err
	}

	response.Set("message", "alert sent")
	response.Set("data", residentAlert)

	if err := c.JSON(http.StatusOK, response); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil

}

func (s *ResidentUtil) ResidentDashboard(c echo.Context) error {
	dbc := utils.Env.Db
	log := utils.Env.Log
	resp := utils.Response{}

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return err
	}

	dashboard := model.ResidentDashboard{}

	residentID := ses.String("admin_id")
	siteID := ses.String("admin_site_id")
	udata, exist := ses.Value("user")
	user := &model.User{}

	if exist {
		user = udata.(*model.User)
		log.Debug("========= usa ", user)

	}

	// log.Debug("========= residencyID ", residentID)
	// log.Debug("========= siteID ", siteID)
	// log.Debug("========= user ", user)

	residentDues := []view.ResidentDueStatus{}

	err = dbc.Model(&residentDues).
		Where("id = ? and site_id = ?", residentID, siteID).Select()
	if err != nil {
		log.Debug("========= err ", err)
		return err
	}

	noticeBoard := []model.NoticeBoard{}
	err = dbc.Model(&noticeBoard).
		Where("site_id = ?", siteID).Limit(5).Select()
	if err != nil {
		log.Debug("========= err ", err)
		return err
	}

	gatePassList := []model.GatePass{}
	err = dbc.Model(&gatePassList).
		Where("resident_id = ? and site_id = ?", residentID, siteID).Select()
	if err != nil {
		log.Debug("========= err ", err)
		return err
	}

	resident := model.Resident{}
	err = dbc.Model(&resident).Where("id = ?", residentID).
		Limit(1).Select()
	if err != nil {
		return err
	}

	ResidentList := []model.Resident{}
	err = dbc.Model(&ResidentList).
		Where("residency_id = ? and type = ?", resident.ResidencyID, model.SecondaryResident).
		Limit(5).Select()
	if err != nil {
		log.Debug("========= err ", err)
		return err
	}

	subResidentList := make([]model.SubResident, len(ResidentList))

	for idx, subResident := range ResidentList {
		subResidentList[idx].FirstName = subResident.FirstName
		subResidentList[idx].LastName = subResident.LastName
		subResidentList[idx].Email = subResident.Email
		subResidentList[idx].Phone = subResident.Phone
		subResidentList[idx].Type = subResident.Type
	}

	accountSummary := view.ResidentBillingSummary{}
	err = dbc.Model(&accountSummary).
		Where("id = ?", residentID).Select()
	if err != nil {
		return err
	}

	// err = copier.Copy(&dashboard.ResidentDues, &residentDues)
	// if err != nil {
	// 	log.Debug("========= err ", err)

	// 	return err
	// }

	dashboard.InDebt = false
	dashboard.SiteID = siteID
	dashboard.ResidentID = residentID
	dashboard.GatePassList.List = gatePassList
	dashboard.GatePassList.Count = len(gatePassList)
	dashboard.AccountBalance = accountSummary.Balance
	dashboard.SubResidents.List = subResidentList
	dashboard.SubResidents.Count = len(subResidentList)
	dashboard.Notification.List = noticeBoard
	dashboard.Notification.Count = len(noticeBoard)
	if accountSummary.Balance.IsNegative() {
		dashboard.InDebt = true
	}

	resp.Set("dashboard", dashboard)

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil
}

func (s *ResidentUtil) InitWebsocket(c echo.Context) error {

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// Write
	err = ws.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
	if err != nil {
		c.Logger().Error(err)
	}

	s.reader(ws)
	return nil
}

func (s *ResidentUtil) reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			s.log.Error(err)
			return
		}

		if err := conn.WriteMessage(messageType, p); err != nil {
			s.log.Error(err)
			return
		}

	}
}
