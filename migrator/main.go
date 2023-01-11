package main

import (
	"fmt"

	"github.com/rs/xid"
)

func main() {
	// logger := util.InitLogger().Sugar()
	// config := util.DBConfig{
	// 	Host:     "127.0.0.1",
	// 	Username: "sysdba",
	// 	Password: "masterkey",
	// 	DBName:   "eve_old",
	// }

	// config2 := util.DBConfig{
	// 	Host:     "127.0.0.1",
	// 	Username: "sysdba",
	// 	Password: "masterkey",
	// 	DBName:   "eve",
	// }

	// oldeve := util.InitDatabase(config)
	// neweve := util.InitDatabase(config2)

	// _ = oldeve
	// _ = neweve
	// _ = logger

	// // operations.CreateResidency(oldeve, neweve, logger)
	// // operations.CreateResident(oldeve, neweve, logger)

	// // err := operations.MigrateUser(oldeve, neweve, logger)
	// // if err != nil {
	// // 	logger.Debug(err)
	// // }

	// // err = operations.MigrateSite(oldeve, neweve, logger)
	// // if err != nil {
	// // 	logger.Debug(err)
	// // }

	// // err = operations.MigrateStreet(oldeve, neweve, logger)
	// // if err != nil {
	// // 	logger.Debug(err)
	// // }

	// err := operations.MigrateContent(oldeve, neweve, logger)
	// if err != nil {
	// 	logger.Debug(err)
	// }

	fmt.Println(xid.New().String())

}
