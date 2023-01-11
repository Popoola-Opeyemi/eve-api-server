package echotools

import (
	"eve/utils"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Env ...
type Env struct {
	Dbc *pg.DB
	Log *zap.Logger
	Cfg *utils.Config
	Rtr *echo.Echo
}

// Handler ...
type Handler interface {
	Initialize(env *Env) error
}

// APIError ...
func APIError(c echo.Context, err error, status int) error {
	utils.Env.Log.Debug(err)

	resp := utils.Response{}
	resp.APIError(err)
	return c.JSON(status, resp)
}
