package shared

import (
	"eve/service"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"

	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// Server ...
type Server struct {
	AppName string
	Rtr     *echo.Echo
	Dbc     *pg.DB
	Cfg     *utils.Config
	Log     *zap.Logger
	log     *zap.SugaredLogger

	serviceList map[string]interface{}
}

// Service ...
func (s Server) Service(name string) interface{} {
	svc, found := s.serviceList[name]
	if !found {
		return nil
	}

	return svc
}

// NewServer ...
func NewServer(appName string, lgr *zap.Logger, cfg *utils.Config, dbc *pg.DB) (srv *Server) {

	// create echo instance
	rtr := echo.New()

	srv = &Server{
		AppName: appName,
		Rtr:     rtr,
		Dbc:     dbc,
		Cfg:     cfg,
		Log:     lgr,

		log:         lgr.Sugar(),
		serviceList: map[string]interface{}{},
	}

	return
}

// Start ...
func (s *Server) Start(handlerList []et.Handler, serviceList map[string]service.IService) error {
	if err := s.initRouter(); err != nil {
		return err
	}

	if err := s.initServices(serviceList); err != nil {
		return err
	}

	if err := s.initHandlers(handlerList); err != nil {
		return err
	}

	bind := fmt.Sprintf(
		"%s:%s",
		s.Cfg.Section("").Key("host").String(),
		s.Cfg.Section("").Key("port").String(),
	)

	if err := s.Rtr.Start(bind); err != nil {
		return err
	}

	return nil
}

func (s *Server) initRouter() error {
	log := s.log
	rtr := s.Rtr
	cfg := s.Cfg

	// rtr = echo.New()
	rtr.HideBanner = true
	// rtr.Logger = log

	// template renderer
	tMgr, err := et.NewTemplateMgr(cfg, log, cfg.FullPath("templates"))
	if err != nil {
		log.Error(err)
		return err
	}
	rtr.Renderer = tMgr

	// middleware
	lc := middleware.LoggerConfig{
		Format: `[${method}] ${status} - ${uri}` +
			` - ${latency_human}, rx:${bytes_in}, tx:${bytes_out}` + "\n",
	}
	rtr.Use(middleware.LoggerWithConfig(lc))

	origins := cfg.Section("url").Key("origin").Strings(",")
	// fmt.Println(origins)
	rtr.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     origins,
		AllowCredentials: true,
		AllowHeaders: []string{
			echo.HeaderOrigin, echo.HeaderContentType,
			echo.HeaderAccept, echo.HeaderXRequestedWith,
		},
	}))

	rtr.Use(middleware.Gzip())

	rtr.Use(middleware.Recover())
	rtr.Use(session.Middleware(sessions.NewFilesystemStore("./tmp", []byte("!chidinmaisafinegirl!"))))
	// rtr.Use(handlers.SiteIDMw(s.Dbc, s.log))

	return nil
}

func (s *Server) initServices(services map[string]service.IService) (err error) {

	for name, svc := range services {
		s.serviceList[name] = svc.Init(s.Dbc, s.Log)
	}
	return
}

func (s *Server) initHandlers(hList []et.Handler) (err error) {

	env := &et.Env{
		Dbc: s.Dbc,
		Log: s.Log,
		Cfg: s.Cfg,
		Rtr: s.Rtr,
	}

	for _, i := range hList {
		i.Initialize(env)
	}

	return
}
