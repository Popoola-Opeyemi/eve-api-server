package handlers

import (
	"encoding/json"
	"eve/service/view"
	"eve/utils"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

// GetResidentDues ...
func (s *Controller) GetResidentDues(c echo.Context) error {
	dbc := s.env.Dbc
	log := s.log
	// siteID := getSiteID(c)

	resID := c.Param("id")

	// get dues info from bill_elements
	dues := []view.ResidentDue{}
	_, err := dbc.Query(
		&dues,
		`
			select
				due_id,
				due,
				balance as amount

			from
				resident_due_status

			where
				id = ?
		`,
		resID,
	)
	if err != nil {
		log.Debug(err)
		return err
	}

	billDues := struct {
		ID    string
		Items json.RawMessage
	}{}
	_, err = dbc.Query(
		&billDues,
		`
			select
				r.id,
				b.items
			from resident_list as r
			left outer join
			 	bill_detail_list as b
				on b.unit_type = r.unit_type and b.status = 1
				and b.site_id = r.site_id 
			where
				r.id = ?
		`,
		resID,
	)
	if err != nil {
		log.Debug(err)
		return err
	}

	// if no dues afe available (resident hasn't been billed)
	if len(dues) == 1 && dues[0].Amount.Equal(decimal.Zero) {
		dues = []view.ResidentDue{}
	}

	// insert missing dues
	for _, i := range gjson.GetBytes(billDues.Items, "#[*]#").Array() {
		if containsDue(dues, i.Get("due_id").String()) {
			continue
		}

		d := view.ResidentDue{
			DueID:  i.Get("due_id").String(),
			Due:    i.Get("name").String(),
			Amount: decimal.Zero,
		}

		dues = append(dues, d)
	}

	resp := utils.Response{}

	resp.Set("list", dues)
	resp.Set("count", len(dues))

	if err = c.JSON(http.StatusOK, resp); err != nil {
		s.log.Error(err)
		return nil
	}

	return nil
}

func containsDue(list []view.ResidentDue, id string) bool {
	for _, i := range list {
		if i.DueID == id {
			return true
		}
	}

	return false
}
