// File: record.go
// File Created: Sunday, 14th July 2019 1:04:42 pm
// Author: Akinmayowa Akinyemi
// -----
// Copyright 2019 Techne Efx Ltd

package resource

import (
	"encoding/json"
	"fmt"
	"strings"

	"eve/utils"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/rs/xid"
)

// Record ...
type Record struct {
	tableName struct{}        `sql:"resource_record"`
	ID        string          `json:"id"`
	SiteID    string          `json:"site_id"`
	Name      string          `json:"resource"`
	Data      json.RawMessage `json:"data"`
}

func (s Record) String() string {
	// byt := []byte("")
	// byt, _ = json.Marshal(s)

	return string(s.Data)
}

// Get ...
func Get(siteID, id string) (record *Record, err error) {
	db := utils.Env.Db
	log := utils.Env.Log

	record = &Record{}
	err = db.Model(record).
		Where("site_id = ? and name = ? and id = ?", siteID, id).
		Select()

	if err != nil {
		if err != pg.ErrNoRows {
			log.Debug(err)
		}

		return nil, err
	}

	return
}

// GetByField get a single record based on the provided filter
func GetByField(siteID, resource string, filter utils.Options) (record *Record, err error) {
	db := utils.Env.Db
	log := utils.Env.Log

	record = &Record{}
	qry := db.Model(record).
		Where("site_id = ? and name = ?", siteID, resource).
		Limit(1)

	qry = QueryFilter(filter, qry)

	err = qry.Select()
	if err != nil {
		if err != pg.ErrNoRows {
			log.Debug(err)
		}

		return nil, err
	}

	return
}

// Find ...
func Find(siteID, resource string, filter utils.Options) (records []Record, err error) {
	db := utils.Env.Db
	log := utils.Env.Log

	records = []Record{}
	qry := db.Model(&records).
		Where("site_id = ? and name = ?", siteID, resource)
	qry = QueryFilter(filter, qry)

	err = qry.Select()
	if err != nil {
		if err != pg.ErrNoRows {
			log.Debug(err)
		}

		return nil, err
	}

	return
}

// Add ...
func Add(siteID string, record *Record) error {
	db := utils.Env.Db
	log := utils.Env.Log

	rType, err := FindResource(siteID, record.Name)
	if err != nil {
		if err == pg.ErrNoRows {
			err = fmt.Errorf("resource '%s' isn't registered", record.Name)
		}

		log.Debug(err)
		return err
	}

	// ensure that model fields are available in Data.
	defa, err := rType.FieldJSON()
	if err != nil {
		log.Debug(err)
		return err
	}

	record.Data, err = utils.JSONMerge(record.Data, defa)
	if err != nil {
		log.Debug(err)
		return err
	}

	// create record
	record.ID = xid.New().String()
	record.SiteID = siteID

	err = utils.Transact(db, log, func(tx *pg.Tx) error {

		if err := tx.Insert(record); err != nil {
			log.Debug(err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Update ...
func Update(siteID, id string, value json.RawMessage) error {
	log := utils.Env.Log

	record := &Record{}
	err := utils.Transact(utils.Env.Db, utils.Env.Log, func(tx *pg.Tx) error {
		_, err := tx.Model(record).
			Set("data = ?", value).
			Where("site_id = ? and id = ?", siteID, id).
			Update()

		if err != nil {
			log.Debug(err)
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
func Save(siteID string, record *Record) error {
	log := utils.Env.Log

	allowedColumns := []string{
		"data",
	}

	err := utils.Transact(utils.Env.Db, utils.Env.Log, func(tx *pg.Tx) error {
		qry := tx.Model(record).
			Column(allowedColumns...).
			Where("site_id = ? and id = ?", siteID, record.ID)

		_, err := qry.Update()
		if err != nil {
			log.Debug(err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// QueryFilter expects filter to contin pairs such asa:
// column:value,..., $limit:x, $offset:y, $order:column
func QueryFilter(filter utils.Options, qry *orm.Query) *orm.Query {
	for k := range filter {

		switch k {
		case "$order":
			qry = qry.Order(fmt.Sprintf("data ->> '%s'", filter.String(k)))
		case "$limit":
			qry = qry.Limit(filter.Int(k))
		case "$offset":
			qry = qry.Offset(filter.Int(k))
		default:
			val := filter.String(k)
			if val[0] == '>' && val[1] == '=' {
				qry = qry.Where(
					fmt.Sprintf("data ->> '%s' >= ?", k),
					strings.TrimLeft(filter.String(k), ">="),
				)
			} else if val[0] == '>' {
				qry = qry.Where(
					fmt.Sprintf("data ->> '%s' > ?", k),
					strings.TrimLeft(filter.String(k), ">"),
				)
			} else if val[0] == '<' && val[1] == '=' {
				qry = qry.Where(
					fmt.Sprintf("data ->> '%s' <= ?", k),
					strings.TrimLeft(filter.String(k), "<="),
				)
			} else if val[0] == '<' {
				qry = qry.Where(
					fmt.Sprintf("data ->> '%s' < ?", k),
					strings.TrimLeft(filter.String(k), "<"),
				)
			} else {
				qry = qry.Where(
					fmt.Sprintf("data ->> '%s' = ?", k),
					filter.String(k),
				)
			}

		}
	}

	return qry
}
