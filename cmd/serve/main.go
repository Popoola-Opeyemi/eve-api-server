package main

import (
	"eve/handlers"
	"eve/service"
	"eve/service/form"
	"eve/service/model"
	"eve/service/view"
	"os"

	"eve/shared"
	"eve/utils"
	et "eve/utils/echotools"
	"eve/utils/resource"
	"runtime"
)

// AppName ...
const AppName = "eve"

func main() {

	// configure the global logger
	logger := utils.InitLogger(utils.DebugMode, utils.ServerLogger)
	defer logger.Sync()

	// get initial config
	cfg, _, err := shared.InitConfig(AppName, logger.Sugar())
	if err != nil {
		os.Exit(model.ServiceUser)
	}

	// init db connection
	dbc, err := utils.InitDb(cfg, logger.Sugar())
	if err != nil {
		os.Exit(model.ServiceUser)
	}

	// register models
	models := registerModels()

	// set service environment
	service.Env.Db = dbc
	service.Env.Log = logger.Sugar()

	utils.Env.Db = dbc
	utils.Env.Cfg = cfg
	utils.Env.Log = logger.Sugar()

	// init utils/resource
	resource.Init(nil)

	// stat task server
	go func() {
		if err := shared.Worker(); err != nil {
			utils.Env.Log.Debug(err)
			os.Exit(model.ServiceUser)
		}
	}()

	// stat task server
	go func() {
		if err := shared.SupportAccount(); err != nil {
			utils.Env.Log.Debug(err)
			os.Exit(model.ServiceUser)
		}
	}()

	// create server
	srv := shared.NewServer(AppName, logger, cfg, dbc)
	sList := map[string]service.IService{}
	hList := []et.Handler{
		&et.CrudAPI{Path: "/api/db", Models: models},
		&handlers.AuthAPI{Path: "/api/auth"},
		&handlers.Associations{Path: "/api/db/association"},
		&handlers.NewResidentRegistration{Path: "/api/db/newresidents"},
		&handlers.Controller{Path: "/api/ctl"},
		&handlers.ResidentUtil{Path: "/api/resident"},
	}

	if err := srv.Start(hList, sList); err != nil {
		logger.Sugar().Error(err)

		os.Exit(model.ServiceUser)
	}
}

type regModel struct {
	Type          interface{}
	Name          string
	Exclude       string
	NoSiteID      bool
	MinAccessType int

	BeforeReadHook et.CrudBeforeReadHook
	AfterReadHook  et.CrudAfterReadHook
	BeforeListHook et.CrudBeforeListHook
	AfterListHook  et.CrudAfterListHook
	BeforeSaveHook et.CrudSaveHook
	AfterSaveHook  et.CrudSaveHook
	DeleteHook     et.CrudDeleteHook
}

func registerModels() []et.ModelInfo {
	modelInfo := []et.ModelInfo{}

	models := []regModel{
		{Type: &model.Site{}, Name: "Site", Exclude: "SiteID", NoSiteID: true, MinAccessType: model.PlatformUser},
		{Type: &model.Registration{}, Name: "Registration", Exclude: "", NoSiteID: true, MinAccessType: model.PlatformUser},
		{Type: &model.NewResidentRegistrations{}, Name: "NewResidentRegistrations", Exclude: "UnitID,Status", NoSiteID: true, MinAccessType: model.ResidentUser,
			BeforeListHook: handlers.BeforeListNewRegistration,
			BeforeSaveHook: handlers.BeforeSaveNewRegistration,
		},
		{Type: &model.User{}, Name: "User", Exclude: "SiteID,IsSiteUser,SubType,ActiveStatus", MinAccessType: 5,
			DeleteHook:     handlers.DeleteUser,
			BeforeSaveHook: handlers.SaveUser,
		},
		{Type: &model.UserType{}, Name: "UserType", Exclude: "SiteID", NoSiteID: true},
		{Type: &model.Street{}, Name: "Street", Exclude: "SiteID", MinAccessType: model.OfficialUser,
			DeleteHook: handlers.DeleteStreet},
		{Type: &model.UnitType{}, Name: "UnitType", Exclude: "SiteID,Type", NoSiteID: true},
		{Type: &model.Unit{}, Name: "Unit", Exclude: "SiteID", MinAccessType: model.OfficialUser,
			DeleteHook:     handlers.DeleteUnit,
			BeforeSaveHook: handlers.BeforeSaveUnit,
		},
		{Type: &model.Resident{}, Name: "Resident", Exclude: "SiteID,Unit,UnitID,PrimaryID,Type,ActiveStatus",
			MinAccessType:  model.ResidentUser,
			BeforeSaveHook: handlers.BeforeSaveResident,
			AfterReadHook:  handlers.ReadResident,
			DeleteHook:     handlers.DeleteResident,
		},
		{Type: &model.Residency{}, Name: "Residency", MinAccessType: model.ResidentUser, Exclude: "ID,Type,SiteID",
			BeforeSaveHook: handlers.SaveResidencyProfile,
			AfterSaveHook:  handlers.AfterSaveResidency,
		},
		{Type: &model.Due{}, Name: "Due", Exclude: "SiteID,DateCreated", MinAccessType: model.ServiceUser},
		{Type: &model.Bill{}, Name: "Bill", Exclude: "SiteID,Items", MinAccessType: model.ResidentUser,
			BeforeSaveHook: handlers.BeforeBillSave,
			AfterReadHook:  handlers.AfterReadBill,
			AfterSaveHook:  handlers.AfterSaveBill},
		{Type: &model.BillItem{}, Name: "BillItem", Exclude: "SiteID", MinAccessType: model.ResidentUser},
		{Type: &model.Transaction{}, Name: "Transaction", Exclude: "SiteID", MinAccessType: model.ResidentUser},
		{Type: &model.NoticeBoard{}, Name: "NoticeBoard", Exclude: "SiteID", MinAccessType: model.ServiceUser},
		{Type: &view.ActiveNotice{}, Name: "ActiveNotice", Exclude: "SiteID", MinAccessType: model.ServiceUser},
		{Type: &view.ExpiredNotice{}, Name: "ExpiredNotice", Exclude: "SiteID", MinAccessType: model.ResidentUser},
		{Type: &model.GatePass{}, Name: "GatePass", Exclude: "SiteID,Token,ResidentID,Resident", MinAccessType: model.SecurityUser,
			AfterReadHook:  handlers.AfterReadGatePass,
			BeforeListHook: handlers.BeforeListGatePass,
			AfterListHook:  handlers.AfterListGatePass,
			AfterSaveHook:  handlers.AfterSaveGatePass,
			BeforeSaveHook: handlers.BeforeSaveGatePass},

		{Type: &model.Visitor{}, Name: "Visitor", Exclude: "SiteID,Security,Resident", MinAccessType: model.SecurityUser,
			BeforeReadHook: handlers.BeforeReadVisitor,
			BeforeSaveHook: handlers.BeforeSaveVisitor,
		},

		{Type: &model.BillGenerate{}, Name: "BillGenerate", Exclude: "SiteID", MinAccessType: model.OfficialUser,
			BeforeSaveHook: handlers.BeforeSaveBillGenerate,
		},
		{
			Type: &model.ResidentAlerts{}, Name: "ResidentAlert", Exclude: "SiteID", MinAccessType: model.SecurityUser,
			BeforeSaveHook: handlers.BeforeResidentSaveAlerts,
		},

		{Type: &model.Invoice{}, Name: "Invoice", Exclude: "SiteID,PaidDues", MinAccessType: model.ResidentUser,
			BeforeReadHook: handlers.BeforeReadInvoice,
			BeforeSaveHook: handlers.BeforeSaveInvoice,
		},

		{Type: &model.Payment{}, Name: "Payment", Exclude: "SiteID", MinAccessType: model.OfficialUser},
		{Type: &model.PaymentPending{}, Name: "PaymentPending", Exclude: "SiteID", MinAccessType: model.ResidentUser,
			BeforeSaveHook: handlers.SavePendingPayment,
			BeforeListHook: handlers.ListPendingPayment,
			DeleteHook:     handlers.DeletePendingPayment,
		},
		{Type: &model.PaymentLog{}, Name: "PaymentLog", Exclude: "", MinAccessType: model.OfficialUser},
		{Type: &model.Content{}, Name: "Content", Exclude: "SiteID", MinAccessType: model.OfficialUser,
			BeforeSaveHook: handlers.BeforeSaveContent},

		{Type: &view.InvoiceList{}, Name: "InvoiceList", Exclude: "SiteID", MinAccessType: model.ResidentUser,
			BeforeListHook: handlers.BeforeInvoiceList,
		},
		{Type: &view.UserView{}, Name: "UserView", Exclude: "SiteID", MinAccessType: model.OfficialUser},
		{Type: &view.ResidentView{}, Name: "ResidentView", Exclude: "", MinAccessType: model.ResidentUser},
		{Type: &view.UnitStreetView{}, Name: "UnitStreetView", Exclude: "", MinAccessType: model.OfficialUser},
		{Type: &view.AssociationView{}, Name: "AssociationView", NoSiteID: true, MinAccessType: model.PlatformUser,
			BeforeSaveHook: handlers.BeforeSaveAssoc,
			DeleteHook:     handlers.DeleteAssoc},
		{Type: &view.AssociationList{}, Name: "AssociationList", NoSiteID: true, MinAccessType: model.PlatformUser},
		{Type: &view.PlatformResidentView{}, Name: "PlatformResidentView", NoSiteID: true, MinAccessType: model.PlatformUser},
		{Type: &view.BillList{}, Name: "BillList", MinAccessType: model.OfficialUser},
		{Type: &view.OldUnitsResidents{}, Name: "OldUnitsResidents", MinAccessType: model.SecurityUser},
		{Type: &view.BillDetailList{}, Name: "BillDetailList", MinAccessType: model.OfficialUser},
		{Type: &view.BillItemList{}, Name: "BillItemList", MinAccessType: model.OfficialUser},
		{Type: &view.UnitList{}, Name: "UnitList", Exclude: "SiteID", MinAccessType: model.OfficialUser,
			BeforeListHook: handlers.BeforeUnitList,
		},
		{Type: &view.AvailableUnitsList{}, Name: "AvailableUnitsList", Exclude: "SiteID", MinAccessType: model.OfficialUser},
		{Type: &view.GatePassList{}, Name: "GatePassList", Exclude: "SiteID", MinAccessType: model.SecurityUser,
			BeforeListHook: handlers.BeforeListGatePass,
			AfterListHook:  handlers.AfterListGatePass,
		},
		{Type: &view.VisitorList{}, Name: "VisitorList", MinAccessType: model.SecurityUser,
			BeforeListHook: handlers.BeforeVisitorList,
		},
		{Type: &view.ResidentList{}, Name: "ResidentList", MinAccessType: model.SecurityUser},

		{Type: &view.ResidentFamilyList{}, Name: "ResidentFamilyList", MinAccessType: model.SecurityUser,
			BeforeListHook: handlers.BeforeResidentFamilyList,
			DeleteHook:     handlers.DeleteResidentFamilyMember,
		},

		{Type: &view.ReportingResidents{}, Name: "ReportingResidents", MinAccessType: model.SecurityUser},
		{Type: &view.ReportingPayments{}, Name: "ReportingPayments", MinAccessType: model.SecurityUser},
		{Type: &view.ReportingBill{}, Name: "ReportingBill", MinAccessType: model.SecurityUser},
		{Type: &view.ReportingUnit{}, Name: "ReportingUnit", MinAccessType: model.SecurityUser},
		{Type: &view.ReportingInvoice{}, Name: "ReportingInvoice", MinAccessType: model.SecurityUser},

		{Type: &view.SecondaryResidentList{}, Name: "SecondaryResidentList", MinAccessType: model.SecurityUser},
		{Type: &view.ResidentAccountStatus{}, Name: "ResidentAccountStatus", MinAccessType: 0},
		{Type: &view.ResidentDueStatus{}, Name: "ResidentDueStatus", MinAccessType: 0},
		{Type: &view.SecurityResidentList{}, Name: "SecurityResidentList", MinAccessType: model.SecurityUser,
			AfterListHook: handlers.AfterListSecurityResidents,
		},
		{Type: &view.InvoiceMasterList{}, Name: "InvoiceMasterList", MinAccessType: model.ResidentUser},
		{Type: &view.PaymentList{}, Name: "PaymentList", MinAccessType: model.ResidentUser,
			BeforeListHook: handlers.BeforeListPayment,
			DeleteHook:     handlers.DeletePayment,
		},
		{Type: &view.ResidentBillingSummary{}, Name: "ResidentBillingSummary", MinAccessType: model.ResidentUser,
			BeforeReadHook: handlers.BeforeReadResidentBilling,
		},
		{Type: &view.InvoiceSummary{}, Name: "InvoiceSummary", MinAccessType: model.ResidentUser,
			BeforeListHook: handlers.BeforeInvoiceSummaryList,
			DeleteHook:     handlers.DeleteInvoiceSummary,
		},
		{Type: &view.ServiceDueStatus{}, Name: "ServiceDueStatus", MinAccessType: model.ServiceUser,
			BeforeListHook: handlers.BeforeServiceDueStatusList,
		},
		{Type: &view.AccountHistory{}, Name: "AccountHistory", MinAccessType: model.ServiceUser,
			BeforeListHook: handlers.BeforeAccountHistoryList,
		},

		{Type: &model.LoginForm{}, Name: "LoginForm", Exclude: "SiteID"},
		{Type: &form.ResidentProfileSave{}, Name: "ResidentProfileSave", Exclude: "", MinAccessType: model.ResidentUser,
			BeforeSaveHook: handlers.SaveResidentProfile,
		},
		{Type: &form.PaymentForm{}, Name: "PaymentForm", Exclude: "", MinAccessType: model.ResidentUser,
			BeforeSaveHook: handlers.SavePayment,
			DeleteHook:     handlers.DeletePayment,
		},
		{Type: &view.PaymentDetails{}, Name: "PaymentDetails", Exclude: "", MinAccessType: model.ResidentUser},
	}

	for i := range models {
		modelInfo = append(modelInfo, et.ModelInfo{
			Type:          models[i].Name,
			Exclude:       models[i].Exclude,
			NoSiteID:      models[i].NoSiteID,
			MinAccessType: models[i].MinAccessType,

			AfterReadHook:  models[i].AfterReadHook,
			BeforeReadHook: models[i].BeforeReadHook,
			BeforeListHook: models[i].BeforeListHook,
			AfterListHook:  models[i].AfterListHook,
			BeforeSaveHook: models[i].BeforeSaveHook,
			AfterSaveHook:  models[i].AfterSaveHook,
			DeleteHook:     models[i].DeleteHook,
		})

		utils.RegisterTypeWithName(models[i].Name, models[i].Type)
	}

	runtime.GC()

	return modelInfo
}
