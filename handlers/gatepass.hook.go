package handlers

import (
	"errors"
	"eve/service/view"
	"fmt"
	"net/http"
	"strings"
	"time"

	"eve/service/model"
	"eve/utils"
	et "eve/utils/echotools"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid"
)

// AfterReadGatePass ...
func AfterReadGatePass(c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (err error) {
	svc := utils.CRUDServiceInstance
	siteID := getSiteID(c)

	// log := utils.Env.Log

	record := frm.(*model.GatePass)

	filter := utils.Options{}
	if len(siteID) > 0 {
		filter["site_id"] = siteID
	}
	retv, err := svc.GetBy("GatePassList", "id", record.ID, filter, "")
	if err != nil {
		et.APIError(c, err, http.StatusInternalServerError)
		return err
	}

	gp, _ := retv.(*view.GatePassList)
	record.Resident = gp.Resident

	now := time.Now()
	if record.Status == 0 &&
		now.Format("2006-01-02") != record.DateCreated.Format("2006-01-02") {
		// expired
		record.Status = 4
	}

	return nil
}

// BeforeListGatePass ...
func BeforeListGatePass(c echo.Context, filter *utils.Options, resp *utils.Response) (stop bool, err error) {

	dbc := utils.Env.Db
	log := utils.Env.Log

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return
	}

	usrType := ses.Int("admin_type")
	usrSubtype := ses.Int("admin_subtype")
	usrId := ses.String("admin_id")

	// run's only for resident
	if usrType == model.ResidentUser {
		res := model.Resident{}
		if err := dbc.Model(&res).Where("id = ?", usrId).Select(); err != nil {
			return true, err
		}

		// primary resident, filter using residency id
		if usrSubtype == 0 {
			(*filter)["residency_id"] = res.ResidencyID
		}

		// secondary resident, filter using resident id
		if usrSubtype == 1 {
			(*filter)["resident_id"] = usrId
		}
	}

	return false, nil
}

// AfterListGatePass ...
func AfterListGatePass(c echo.Context, data interface{}, resp *utils.Response) (err error) {
	// svc := utils.CRUDServiceInstance
	// siteID := getSiteID(c)
	log := utils.Env.Log

	log.Debug("in hook ...")

	records := data.(*[]view.GatePassList)
	now := time.Now()
	for i := 0; i < len(*records); i++ {
		rec := &(*records)[i]
		if rec.Status == 0 &&
			now.Format("2006-01-02") != rec.DateCreated.Format("2006-01-02") {
			// expired
			//TODO: commented this because matthew said eve wants to be able to view all unused and expired gatepasses
			// rec.Status = 4
			log.Debug("hook -->", (*records)[i])
		}
	}

	return nil
}

// BeforeSaveGatePass ...
func BeforeSaveGatePass(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (stop bool, err error) {
	log := utils.Env.Log
	dbc := utils.Env.Db

	ses, err := et.NewSessionMgr(c, "")
	if err != nil {
		log.Debug(err)
		return
	}
	usrType := ses.Int("admin_type")

	pass := frm.(*model.GatePass)

	// only residents can create tokens
	if (len(pass.ID) == 0 || pass.ID == "new") && usrType == 3 {

		// check if the resident is disabled before continuing
		usrID := pass.ResidentID

		res := model.Resident{}

		joinQry := fmt.Sprintf("left join residency as rs on rs.id = resident.residency_id")

		err := dbc.Model(&res).ColumnExpr("resident.*").
			Join(joinQry).
			Where("resident.id = ?", usrID).Select()

		if err != nil {
			log.Debug(err)
			return false, err
		}

		// user disabled
		if res.Status == model.IsDisabled {
			dError := errors.New("Cannot issue gatepass, user is disabled")
			return true, dError
		}

		pass.Token, err = gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", 12)
		pass.Token = dashed(pass.Token)
		pass.ResidentID = ses.String("admin_id")
	}

	return false, err
}

// AfterSaveGatePass ...
func AfterSaveGatePass(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (stop bool, err error) {

	pass := frm.(*model.GatePass)
	resp.Set("token", pass.Token)

	return false, nil
}

func dashed(val string) string {
	buff := []string{}
	for i := 0; i < len(val); i += 4 {
		buff = append(buff, val[i:i+4])
	}

	return strings.Join(buff, "-")
}
