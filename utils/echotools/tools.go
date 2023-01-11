package echotools

import (
	"eve/utils"

	"github.com/labstack/echo/v4"
)

// GetParam checks if p can be found in c.Param or c.QueryParam and returns the value found
func GetParam(c echo.Context, p string) string {
	val := c.Param(p)
	if len(val) > 0 {
		return val
	}

	return c.QueryParam(p)

}

// IntParam converts c.Param(p) or c.QueryParam(p) to int
func IntParam(c echo.Context, p string) int {
	val := c.Param(p)
	if len(val) > 0 {
		return utils.Atoi(val)
	}

	return utils.Atoi(c.QueryParam(p))

}

// Int64Param converts c.Param() to int64
func Int64Param(c echo.Context, p string) int64 {
	val := c.Param(p)
	if len(val) > 0 {
		return utils.Atoi64(val)
	}

	return utils.Atoi64(c.QueryParam(p))
}

// IsAjax .returns true if request is an ajax request is an ajax request
// this function depends on the precense of the X-Requested-With" header
func IsAjax(c echo.Context) bool {
	return c.Request().Header.Get("X-Requested-With") == "xmlhttprequest"
}
