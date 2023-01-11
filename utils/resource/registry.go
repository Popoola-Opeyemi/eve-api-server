// File: registry.go
// File Created: Sunday, 14th July 2019 1:24:12 pm
// Author: Akinmayowa Akinyemi
// -----
// Copyright 2019 Techne Efx Ltd

package resource

import (
	"encoding/json"
	"eve/utils"
	"fmt"
	"sync"

	"github.com/go-pg/pg"
	"github.com/rs/xid"
)

// Type stores meta data for a Resource type
type Type struct {
	tableName struct{} `sql:"resource_type"`
	ID        string
	SiteID    string
	Name      string

	// model: {
	//   name:{type:0, required: false, length:100, rows: 3, choices:[{value:'', label:''}, ...]},
	//   ...
	// }
	Model json.RawMessage
}

// Registry ...
type Registry struct {
	mtx       sync.Mutex
	Resources map[string]Type
}

var registry *Registry

// FieldJSON ...
func (s Type) FieldJSON() (json.RawMessage, error) {
	model := map[string]interface{}{}
	extract := map[string]interface{}{}

	err := json.Unmarshal(s.Model, &model)
	if err != nil {
		return nil, err
	}

	for k, v := range model {
		attr, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("value of k not map[string]interface{}")
		}

		defa, exists := attr["default"]
		if exists {
			extract[k] = defa
		} else {
			extract[k] = ""
		}
	}

	retv, err := json.Marshal(extract)
	if err != nil {
		return nil, err
	}

	return retv, nil
}

// RegisterResource create a record for a resource in the registry and store in the local cache
func RegisterResource(siteID string, record *Type) error {
	db := utils.Env.Db
	log := utils.Env.Log

	record.ID = xid.New().String()
	record.SiteID = siteID

	err := utils.Transact(db, log, func(tx *pg.Tx) error {

		if err := tx.Insert(record); err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	registry.mtx.Lock()
	defer registry.mtx.Unlock()

	regName := fmt.Sprintf("%s:%s", siteID, record.Name)
	registry.Resources[regName] = *record

	return nil
}

// FindResource find resource type in local cache if not there get it from the database
func FindResource(siteID, name string) (Type, error) {
	registry.mtx.Lock()
	defer registry.mtx.Unlock()

	var resource Type
	// find resource type in cache
	regName := fmt.Sprintf("%s:%s", siteID, name)
	resource, found := registry.Resources[regName]
	if found {
		return resource, nil
	}

	db := utils.Env.Db
	log := utils.Env.Log

	err := db.Model(&resource).
		Where("site_id = ? and name = ?", siteID, name).
		Select()
	if err != nil {
		log.Error(err)
		return resource, err
	}

	registry.Resources[regName] = resource

	return resource, nil
}
