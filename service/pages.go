package service

import (
	"encoding/json"
	"fmt"

	"eve/service/model"

	"eve/utils"

	"github.com/go-pg/pg"
	"github.com/rs/xid"
)

// PageSvc an instance of the IEntity service
type PageSvc struct {
}

// Page ...
type Page struct {
	ID     string
	SiteID string
	Name   string
	URL    string
	Data   json.RawMessage
}

// IPage ...
type IPage interface {
	Get(siteID, field, id string) (*Page, error)
	Create(siteID string, record *Page) error
	Save(siteID string, record *Page) error
	Delete(siteID, id string) error
}

var _ IPage = PageSvc{}

// Get ...
func (PageSvc) Get(siteID, field, value string) (record *Page, err error) {
	db := Env.Db
	log := Env.Log

	record = &Page{}
	err = db.Model(record).
		Where(
			fmt.Sprintf("site_id = ? and %s = ?", field),
			siteID, value,
		).
		Select()

	if err != nil {
		if err != pg.ErrNoRows {
			log.Error(err)
		}

		return nil, err
	}

	return
}

// List ...
func (PageSvc) List(siteID string, filter utils.Options) (records []Page, err error) {
	db := Env.Db
	log := Env.Log

	qry := db.Model(&records).
		Where("site_id = ?", siteID)
	qry = utils.QueryFilter(filter, qry)

	if err = qry.Select(); err != nil {
		log.Error(err)
		return
	}

	return
}

// Create ...
func (PageSvc) Create(siteID string, record *Page) (err error) {
	db := Env.Db
	log := Env.Log

	record.ID = xid.New().String()
	record.SiteID = siteID

	err = utils.Transact(db, log, func(tx *pg.Tx) error {

		if err := tx.Insert(record); err != nil {
			log.Error(err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// Save ...
func (PageSvc) Save(siteID string, record *Page) (err error) {
	db := Env.Db
	log := Env.Log

	allowedColumns := []string{
		"name", "url", "data",
	}

	err = utils.Transact(db, log, func(tx *pg.Tx) error {

		_, err := tx.Model(record).
			Column(allowedColumns...).
			Where(
				"site_id = ? and id = ?",
				siteID, record.ID,
			).Update()
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// Delete ...
func (PageSvc) Delete(siteID, id string) (err error) {
	db := Env.Db
	log := Env.Log

	if len(id) == 0 {
		log.Error(err)
		err = model.ValidationError{Msg: "id not supplied"}
		return
	}

	record := &Page{ID: id}

	err = utils.Transact(db, log, func(tx *pg.Tx) error {
		_, err := tx.Model(record).
			Where(
				"site_id = ? and id = ?",
				siteID, id,
			).Delete()
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		return
	}

	return
}
