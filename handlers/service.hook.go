package handlers

import (
	"eve/service/view"
	"eve/utils"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

// BeforeServiceDueStatusList ...
func BeforeServiceDueStatusList(c echo.Context, filter *utils.Options, resp *utils.Response) (bool, error) {
	dbc := utils.Env.Db
	log := utils.Env.Log
	siteID := filter.String("site_id")

	records := []view.ServiceDueStatus{}

	// if date not provided use current date
	if _, ok := (*filter)["date"]; !ok {
		(*filter)["date"] = time.Now().Format(utils.FormatYYYYMMDDHHmmSS)
	}
	dateStr := filter.String("date")
	if len(dateStr) == 8 {
		dateStr = fmt.Sprintf("%s 23:59:59", dateStr)
	}

	// if due is provided include due filter
	dueFilter := "?"
	if _, ok := (*filter)["due_id"]; ok {
		dueFilter = fmt.Sprintf(" and t.due_id = ?")
	}

	sql := fmt.Sprintf(`
		with "billed_dues" as (
			select
				resident_id,
				sum(t.amount) as balance

			from
				"transaction" as t

			where
				t.date_created <= ?
				%s

			group by
			resident_id
		)
		select
			r.id,
			concat(r.first_name, ' ', r.last_name) as resident,
			case when d.balance is null then 0.00 else d.balance end as balance
		from
			"resident" as r
		left join 
			"residency" as rs on rs.id = r.residency_id
		left join
			"billed_dues" as d on d.resident_id = r.id
		where
			r.type = 1 and
			rs.site_id = ?
	`, dueFilter)

	// get records from db
	_, err := dbc.Query(&records, sql,
		dateStr,
		filter.String("due_id"),
		siteID,
	)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	resp.Set("list", records)
	resp.Set("count", len(records))

	return true, nil
}
