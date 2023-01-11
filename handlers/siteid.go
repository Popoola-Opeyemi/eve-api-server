package handlers

import (
	"eve/utils"
	"sync"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var cache *sync.Map

// SiteIDMw ...
func SiteIDMw(dbc *pg.DB, log *zap.SugaredLogger) echo.MiddlewareFunc {
	if cache == nil {
		cache = &sync.Map{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			log.Debug("c.Request().Host -> ", c.Request().Host)
			subdomain := utils.GetSubdomain(c.Request().Host)
			retv, found := cache.Load(subdomain)
			siteID := "unknown"

			if !found {
				// get site_id from db
				_, err := dbc.Query(pg.Scan(&siteID), "SELECT id from site where subdomain = ?", subdomain)
				if err != nil {
					log.Error(err)
				} else if siteID != "unknown" {
					// update cache
					cache.Store(subdomain, siteID)
				}

			} else {
				siteID = retv.(string)
			}

			c.Set("hostSiteID", siteID)

			if err := next(c); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}
