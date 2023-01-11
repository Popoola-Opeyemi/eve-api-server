package utils

// cSpell:ignore mkr, gocraft, gommon, Sprintf, dbname, Infof
import (
	"crypto/tls"
	"fmt"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type dbLogger struct {
	log *zap.SugaredLogger
}

func (s dbLogger) BeforeQuery(q *pg.QueryEvent) {}

func (s dbLogger) AfterQuery(event *pg.QueryEvent) {
	if event == nil {
		return
	}

	query, err := event.FormattedQuery()
	if err != nil {
		s.log.Error(err)
		return
	}

	if event.Result != nil {
		s.log.Debugf("\n%s - [%d, %d] \n",
			query,
			event.Result.RowsAffected(), event.Result.RowsReturned(),
		)
	} else {
		s.log.Debugf("\n%s", query, "\n")
	}

	return
}

// InitDb setup the applications database connection
func InitDb(cfg *Config, log *zap.SugaredLogger) (db *pg.DB, err error) {
	log.Debug("initializing database connection")

	dbCfg := cfg.Section("db").KeysHash()
	orm.SetTableNameInflector(func(s string) string {
		return s
	})

	// ssl mode = disable
	if dbCfg["sslmode"] == "disable" {
		url := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
			dbCfg["user"],
			dbCfg["password"],
			dbCfg["dbname"])

		opt, err := pg.ParseURL(url)
		if err != nil {
			return nil, err
		}

		db = pg.Connect(opt)

	} else {

		db = pg.Connect(&pg.Options{
			Addr:     dbCfg["host"],
			User:     dbCfg["user"],
			Password: dbCfg["password"],
			Database: dbCfg["dbname"],

			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})
	}
	db.AddQueryHook(dbLogger{log: log})

	log.Debugf("connected to db with: driver:%s, dbname:%s", dbCfg["driver"], dbCfg["dbname"])

	return
}

type txFunc func(*pg.Tx) error

// Transact is a closure that wraps a transaction
func Transact(db *pg.DB, log *zap.SugaredLogger, fn txFunc) error {
	tx, err := db.Begin()
	if err != nil {
		log.Error(err)
		return err
	}

	if err := fn(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Error(err)
			return err
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
