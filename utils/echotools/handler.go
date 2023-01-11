package echotools

import (
	"eve/utils"

	"github.com/labstack/echo/v4"
)

// // Service ...
// func (s Environment) Service(name string) interface{} {
// 	_, exists := s.ServiceList[name]
// 	if !exists {
// 		return nil
// 	}

// 	return s.ServiceList[name]
// }

// PageHook ...
type PageHook func(c echo.Context, path string) (utils.Map, error)

// Page ...
type Page struct {
	Template string
	Hook     PageHook
}

// RenderPage ...
func RenderPage(c echo.Context, pages map[string]Page, tplEnv utils.Map, url bool) (rendered bool, err error) {

	path := c.Param("path")
	if url {
		path = c.Request().URL.Path
	}

	if len(path) == 0 {
		path = "/"
	}
	// clone tplEnv
	data := utils.ShallowMapMerge(tplEnv, utils.Map{})

	// is path in pageTables ?
	if p, inTable := pages[path]; inTable {
		// does page have a hook?
		if p.Hook != nil {
			newData, err := p.Hook(c, path)
			if err != nil {
				return false, err
			}
			// merge returned data with existing data
			data = utils.ShallowMapMerge(data, newData)
		}

		// render the page
		return true, c.Render(200, p.Template, data)
	}

	// path not found in page table
	return false, nil
}
