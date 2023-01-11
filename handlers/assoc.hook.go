package handlers

import (
	"eve/service/model"
	"eve/service/view"
	"eve/shared"
	"eve/utils"
	et "eve/utils/echotools"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
)

type siteDomain struct {
	ID        string
	Subdomain string
}

// BeforeSaveAssoc ...
func BeforeSaveAssoc(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, frm interface{}, resp *utils.Response) (stop bool, err error) {
	svc := utils.CRUDServiceInstance
	log := utils.Env.Log

	oid := c.Param("id")
	record := frm.(*view.AssociationView)
	site := &model.Site{}
	user := &model.User{}

	site.ID = record.ID
	site.Name = record.Name
	site.Subdomain = record.Subdomain
	site.Status = record.Status
	site.DateRegistered = utils.NewDateTime(time.Now())
	site.Attr = record.Attr
	site.SiteCode = utils.MakeRandText(10)

	user.ID = record.AdminID
	user.SiteID = site.ID
	user.FirstName = record.AdminFirstName
	user.LastName = record.AdminLastName
	user.Email = strings.ToLower(record.AdminEmail)
	user.Phone = record.AdminPhone
	user.Password = record.AdminPassword
	user.Status = 1
	user.Role = int(et.RoleSuperUser)
	user.Type = 5 // admin official
	user.IsSiteUser = true

	subd := siteDomain{}
	if _, err = tx.QueryOne(&subd, "select id, subdomain from site where subdomain=?", record.Subdomain); err != nil {
		if err != pg.ErrNoRows {
			log.Debug(err)
			return
		}
	}

	if len(oid) == 0 || oid == "new" {
		if len(subd.Subdomain) > 0 {
			err = fmt.Errorf("subdomain '%s' has already been assigned", subd.Subdomain)
			return
		}

		site.Status = 1
		// create site record
		if err = svc.Create(tx, "Site", site, false); err != nil {
			log.Debug(err)
			return
		}

		// create user record for the site
		user.SiteID = site.ID
		if err = svc.Create(tx, "User", user, false); err != nil {
			log.Debug(err)
			return
		}

		// update registration record if the request is an approval
		if len(record.RegistrationID) > 1 {
			_, err = tx.ExecOne("update registration set status = 1 where id = ?", record.RegistrationID)
			if err != nil {
				log.Debug(err)
				return
			}
		}

		if len(c.QueryParam("reg_apr")) > 0 {

			// email confirmation
			regEml, err := shared.MakeRegConfirmation(tx, site, user)
			if err != nil {
				log.Debug(err)
				return false, err
			}

			regEml.To = user.Email
			_, err = tx.Exec(`
			insert into task_queue (site_id, type, data)
				values(?, 1, ?)
			`, site.ID, &regEml)
			if err != nil {
				log.Debug(err)
				return false, err
			}
		}

		resp.Set("id", site.ID)

	} else {
		if len(subd.Subdomain) > 0 && subd.ID != record.ID {
			err = fmt.Errorf("subdomain '%s' has already been assigned", subd.Subdomain)
			return
		}

		excludedFields := []string{"ID", "DateRegistered"}
		log.Debug("site --->", site, excludedFields)
		if err = svc.Save(tx, "Site", site, excludedFields); err != nil {
			log.Debug(err)
			return
		}

		if len(subd.Subdomain) > 0 && subd.ID != record.ID {
			excludedFields = []string{"ID", "SiteID", "FirstName", "LastName", "Email", "Phone", "Status", "Attr", "Role", "Type", "SubType", "IsSiteUser"}
			if err = svc.Save(tx, "User", user, excludedFields); err != nil {
				log.Debug(err)
				return
			}
		}
	}

	/* Beginning of creating a hidden user
	- once an association is registered create a new user with the hidden field
	- @ view on the hidden user, a password is generated for the session
	-  passwords are valid for 4 hours
	*/

	log.Debug("here now and direct ")
	passwordHash, err := utils.HashPassword("admin")
	if err != nil {
		return false, err
	}

	if len(subd.Subdomain) > 0 && subd.ID != record.ID {
		hiddenUser := &model.User{}

		hiddenUser.Email = fmt.Sprintf("%s@%s.com", "support", strings.ToLower(record.Subdomain))
		hiddenUser.ID = xid.New().String()
		hiddenUser.SiteID = site.ID
		hiddenUser.FirstName = fmt.Sprintf("%s", "support")
		hiddenUser.LastName = fmt.Sprintf("%s", record.Subdomain)

		hiddenUser.Password = passwordHash
		hiddenUser.Status = 1
		hiddenUser.Role = int(et.RoleSuperUser)
		hiddenUser.Type = model.SupportUser
		hiddenUser.IsSiteUser = false
		hiddenUser.Phone = "1111111111"
		hiddenUser.SupportAccount = true

		// hiddenUser.Attr = []byte(`{}`)

		if err = svc.Create(tx, "User", hiddenUser, false); err != nil {
			log.Debug(err)
			return
		}

	}

	return true, nil
}

// DeleteAssoc ...
func DeleteAssoc(tx *pg.Tx, c echo.Context, mi *et.ModelInfo, resp *utils.Response) (stop bool, err error) {
	svc := utils.CRUDServiceInstance
	dbc := utils.Env.Db
	log := utils.Env.Log
	oid := c.Param("id")

	// if site has streets registered refuse to delete site
	count, err := dbc.Model((*model.Street)(nil)).Where("site_id = ?", oid).Count()
	if err != nil {
		log.Debug(err)
		return
	}
	if count > 0 {
		err = fmt.Errorf("cant delete: association is active")

		resp.APIError(err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err = svc.Delete(tx, "Site", oid); err != nil {
		log.Debug(err)
		return
	}

	return true, nil
}
