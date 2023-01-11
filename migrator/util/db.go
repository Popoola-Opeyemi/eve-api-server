package util

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/mattes/migrate/database/postgres"
)

func InitDatabase(dbc DBConfig) *gorm.DB {

	user := dbc.Username
	dbname := dbc.DBName
	pass := dbc.Password
	dbHost := dbc.Host

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, user, dbname, pass))

	db.LogMode(true)
	db.SingularTable(true)

	if err != nil {
		log.Println("Could not connect to the database:", err)
	}

	return db
}
