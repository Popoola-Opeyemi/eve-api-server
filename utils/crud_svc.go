package utils

// cspell: ignore frms, ICRUD

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/jinzhu/copier"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

// CRUDService ...
type CRUDService interface {
	Get(typeName, field, value, table string) (interface{}, error)
	GetBy(typeName, field, value string, filter Options, table string) (interface{}, error)
	List(typeName string, filter Options, table string) (interface{}, error)
	ListAndCount(typeName string, filter Options, table string) (interface{}, int, error)
	Create(tx *pg.Tx, typeName string, frm interface{}, useID bool) error
	CreateMultiple(recs []TypeRecord) error
	Save(tx *pg.Tx, typeName string, frm interface{}, exclude []string) error
	Delete(tx *pg.Tx, typeName string, id string) error
}

// CRUD an instance of the ICRUD service
type CRUD struct {
	db  *pg.DB
	log *zap.SugaredLogger
}

var _ CRUDService = CRUD{}

// CRUDServiceInstance ...
var CRUDServiceInstance *CRUD

// Init initialize this instance and satisfy IService interface
func (s *CRUD) Init(db *pg.DB, log *zap.Logger) (err error) {
	s.db = db
	s.log = log.Sugar()

	if CRUDServiceInstance == nil {
		CRUDServiceInstance = s
	}

	return nil
}

// Get a record by field and value.
// i.e select * from table where field=value
func (s CRUD) Get(typeName, field, value, table string) (record interface{}, err error) {
	record, err = MakePointerType(typeName)
	if err != nil {
		return
	}

	qry := s.db.Model(record)
	if len(table) > 0 {
		qry = qry.Table(table)
	}

	err = qry.Where(
		fmt.Sprintf("%s = ?", field),
		value,
	).
		Limit(1).
		Select()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}

		return
	}

	return
}

// GetBy query for a record by field and value with filters
// i.e select * from table where field=value and <QueryFilter(filter)>
func (s CRUD) GetBy(typeName, field, value string, filter Options, table string) (record interface{}, err error) {

	if typeName == "Resident" {

		record, err = MakePointerType(typeName)

		if err != nil {
			return nil, err
		}

		qry := s.db.Model(record)
		if len(table) > 0 {
			qry = qry.Table(table)
		}

		joinQry := fmt.Sprint("left join residency as rs on rs.id = resident.residency_id")

		qry = qry.ColumnExpr("resident.*").
			ColumnExpr("rs.site_id, rs.unit_id, rs.active_status").
			Join(joinQry).
			Where(fmt.Sprintf("%s = ?", "resident."+field), value)

		filter["$limit"] = 1
		delete(filter, "site_id")
		qry = QueryFilter(filter, qry)

		err = qry.Select()
		if err != nil {
			return nil, err
		}

	}

	if typeName != "Resident" {

		record, err = MakePointerType(typeName)
		if err != nil {
			return
		}

		qry := s.db.Model(record)
		if len(table) > 0 {
			qry = qry.Table(table)
		}

		qry = qry.Where(
			fmt.Sprintf("%s = ?", field),
			value,
		)

		filter["$limit"] = 1
		qry = QueryFilter(filter, qry)

		if err = qry.Select(); err != nil {
			s.log.Debug(err)
			return
		}
	}

	return
}

// List param filter should be of the form:
// column:value,..., $limit:x,$offset:y,$order:column
//
// Example: List("Users", Options{"id": 1234, "$limit": 5,} "")
// ->> select * from users where id=123 limit 5
//
// Example: List("Users", Options{"id": 1234, "$limit": 5,} "people")
// ->> select * from people where id=123 limit 5
func (s CRUD) List(typeName string, filter Options, table string) (retv interface{}, err error) {
	record, err := MakeSlicePointerType(typeName)
	if err != nil {
		s.log.Error(err)
		return
	}

	qry := s.db.Model(record)
	qry = QueryFilter(filter, qry)

	if err = qry.Select(); err != nil {
		s.log.Debug(err)
		return
	}

	retv = record
	return
}

// ListAndCount same as list but also returns a count of records in the table
func (s CRUD) ListAndCount(typeName string, filter Options, table string) (retv interface{}, count int, err error) {
	record, err := MakeSlicePointerType(typeName)
	if err != nil {
		s.log.Error(err)
		return
	}

	qry := s.db.Model(record)
	qry = QueryFilter(filter, qry)

	if count, err = qry.SelectAndCount(); err != nil {
		s.log.Debug(err)
		return
	}

	retv = record
	return
}

// Create ...
func (s CRUD) Create(tx *pg.Tx, typeName string, frm interface{}, useID bool) (err error) {
	record, err := MakePointerType(typeName)
	if err != nil {
		s.log.Error(err)
		return
	}

	if err := copier.Copy(record, frm); err != nil {
		s.log.Debug(err)
		return err
	}

	var newID string
	if useID == false {
		newID = xid.New().String()
		SetStructField(record, "ID", newID)
		SetStructField(frm, "ID", newID)
	}

	if StructHasField(frm, "Password") {
		password := GetStructField(frm, "Password").String()
		if len(password) > 0 && password != "***" {
			retv, err := HashPassword(password)
			if err != nil {
				s.log.Debug(err)
				return err
			}

			SetStructField(record, "Password", retv)
		}
	}

	sqlFn := func(tx *pg.Tx) error {

		if err := tx.Insert(record); err != nil {
			s.log.Debug(err)
			return err
		}

		return nil
	}

	if tx == nil {
		err = Transact(s.db, s.log, sqlFn)
	} else {
		err = sqlFn(tx)
	}

	if err != nil {
		return err
	}

	return nil
}

// TypeRecord ...
type TypeRecord struct {
	Name   string
	Struct interface{}
}

// CreateMultiple ...
func (s CRUD) CreateMultiple(frms []TypeRecord) (err error) {
	records := []interface{}{}
	for i := 0; i < len(frms); i++ {
		rec, err := MakePointerType(frms[i].Name)
		if err != nil {
			s.log.Error(err)
			return err
		}

		if err := copier.Copy(rec, frms[i].Struct); err != nil {
			s.log.Debug(err)
			return err
		}

		newID := xid.New().String()
		SetStructField(rec, "ID", newID)
		SetStructField(frms[i].Struct, "ID", newID)

		if StructHasField(frms[i].Struct, "Password") {
			password := GetStructField(frms[i].Struct, "Password").String()
			if len(password) > 0 && password != "***" {
				retv, err := HashPassword(password)
				if err != nil {
					s.log.Debug(err)
					return err
				}

				SetStructField(rec, "Password", retv)
			}
		}

		records = append(records, rec)
	}

	err = Transact(s.db, s.log, func(tx *pg.Tx) error {

		for i := 0; i < len(records); i++ {
			if err := tx.Insert(records[i]); err != nil {
				s.log.Debug(err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Save ...
func (s CRUD) Save(tx *pg.Tx, typeName string, frm interface{}, exclude []string) (err error) {

	record, err := MakePointerType(typeName)
	if err != nil {
		s.log.Error(err)
		return
	}

	if err := copier.Copy(record, frm); err != nil {
		s.log.Error(err)
		return err
	}

	allowedColumns := ListStructFields(record, exclude, true)

	if StructHasField(frm, "Password") {
		password := GetStructField(frm, "Password").String()
		if len(password) > 0 && password != "***" {
			retv, err := HashPassword(password)
			if err != nil {
				s.log.Error(err)
				return err
			}

			SetStructField(record, "Password", retv)
		} else {
			exclude = append(exclude, "Password")
			allowedColumns = ListStructFields(record, exclude, true)
		}
	}

	// if there are no columns left to update, return
	if len(allowedColumns) == 0 {
		return nil
	}

	sqlFn := func(tx *pg.Tx) error {
		_, err := tx.Model(record).
			Column(allowedColumns...).
			WherePK().
			Update()
		if err != nil {
			s.log.Debug(err)
			return err
		}

		return nil
	}

	if tx == nil {
		err = Transact(s.db, s.log, sqlFn)
	} else {
		err = sqlFn(tx)
	}

	if err != nil {
		return err
	}

	return nil
}

// Delete ...
func (s CRUD) Delete(tx *pg.Tx, typeName string, id string) (err error) {

	if len(id) == 0 {
		err = fmt.Errorf("id not supplied")
		s.log.Debug()
		return
	}

	record, err := MakePointerType(typeName)
	if err != nil {
		s.log.Error(err)
		return
	}

	SetStructField(record, "ID", id)

	sqlFn := func(tx *pg.Tx) error {
		if err := tx.Delete(record); err != nil {
			s.log.Debug(err)
			return err
		}

		return nil
	}

	if tx == nil {
		err = Transact(s.db, s.log, sqlFn)
	} else {
		err = sqlFn(tx)
	}
	if err != nil {
		return
	}

	return
}

// QueryFilter expects filter to contin pairs such asa:
// column:value,..., $limit:x, $offset:y, $order:column
// column:>value, column:>=value, column:<value, column:<=value
func QueryFilter(filter Options, qry *orm.Query) *orm.Query {
	Env.Log.Debug("filter: ", filter)

	for _, k := range filter.List() {

		switch k.Key {
		case "$order":
			order := strings.Split(IfToString(k.Value), "$")
			qry = qry.Order(order...)
		case "$limit":
			if IfToInt(k.Value) > 0 {
				qry = qry.Limit(IfToInt(k.Value))
			}
		case "$offset":
			qry = qry.Offset(IfToInt(k.Value))
		default:
			val := IfToString(k.Value)
			if len(val) > 2 && val[0] == '>' && val[1] == '=' {
				qry = qry.Where(
					fmt.Sprintf("%s >= ?", k.Key),
					strings.TrimLeft(IfToString(k.Value), ">="),
				)
			} else if len(val) > 0 && val[0] == '>' {
				qry = qry.Where(
					fmt.Sprintf("%s > ?", k.Key),
					strings.TrimLeft(IfToString(k.Value), ">"),
				)
			} else if len(val) > 2 && val[0] == '<' && val[1] == '=' {
				qry = qry.Where(
					fmt.Sprintf("%s <= ?", k.Key),
					strings.TrimLeft(IfToString(k.Value), "<="),
				)
			} else if len(val) > 0 && val[0] == '<' {
				qry = qry.Where(
					fmt.Sprintf("%s < ?", k.Key),
					strings.TrimLeft(IfToString(k.Value), "<"),
				)
			} else if len(val) > 0 && (val[0] == '%' || val[len(val)-1] == '%') {
				qString := strings.ReplaceAll(IfToString(k.Value), "__", ",")
				qry = qry.Where(
					fmt.Sprintf("%s ILIKE ?", k.Key),
					qString,
				)
			} else {
				qry = qry.Where(
					fmt.Sprintf("%s = ?", k.Key),
					IfToString(k.Value),
				)
			}
		}
	}

	return qry
}
