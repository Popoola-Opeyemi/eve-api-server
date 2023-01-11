package service

import (
	"eve/service/model"
	"fmt"

	"eve/utils"

	"github.com/go-pg/pg"
	"github.com/rs/xid"
)

// SiteSvc an instance of the IEntity service
type SiteSvc struct {
}

// Site ...
type Site struct {
	ID   string
	Name string
	Host string
}

// ISite ...
type ISite interface {
	Get(field, id string) (*Site, error)
	Create(record *Site) error
	Save(record *Site) error
	Delete(id string) error
}

var _ ISite = SiteSvc{}

// Get ...
func (SiteSvc) Get(field, value string) (record *Site, err error) {
	db := Env.Db
	log := Env.Log

	record = &Site{}
	err = db.Model(record).
		Where(
			fmt.Sprintf("%s = ?", field),
			value,
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
func (SiteSvc) List(filter utils.Options) (records []Site, err error) {
	db := Env.Db
	log := Env.Log

	qry := db.Model(&records)
	qry = utils.QueryFilter(filter, qry)

	if err = qry.Select(); err != nil {
		log.Error(err)
		return
	}

	return
}

// Create ...
func (SiteSvc) Create(record *Site) (err error) {
	db := Env.Db
	log := Env.Log

	record.ID = xid.New().String()

	err = utils.Transact(db, log, func(tx *pg.Tx) error {
		if err := tx.Insert(record); err != nil {
			log.Error(err)
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Save ...
func (SiteSvc) Save(record *Site) (err error) {
	db := Env.Db
	log := Env.Log

	allowedColumns := []string{
		"name", "host",
	}

	err = utils.Transact(db, log, func(tx *pg.Tx) error {

		_, err := tx.Model(record).
			Column(allowedColumns...).
			WherePK().
			Update()
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Delete ...
func (SiteSvc) Delete(id string) (err error) {
	db := Env.Db
	log := Env.Log

	if len(id) == 0 {
		log.Error(err)
		err = model.ValidationError{Msg: "id not supplied"}
		return
	}

	record := &Site{ID: id}

	err = utils.Transact(db, log, func(tx *pg.Tx) error {
		err := tx.Delete(record)
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
