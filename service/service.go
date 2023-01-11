package service

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

// // Options option map
// type Options map[string]interface{}

// SharedEnv ...
type SharedEnv struct {
	Db  *pg.DB
	Log *zap.SugaredLogger
}

// IService interface for service initiation
type IService interface {
	Init(*pg.DB, *zap.Logger) (service interface{})
}

// Env a global instance of SharedEnv
var Env *SharedEnv

func init() {
	Env = &SharedEnv{}
}
