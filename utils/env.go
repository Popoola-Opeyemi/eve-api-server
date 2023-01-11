package utils

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

// SharedEnv ...
type SharedEnv struct {
	Db  *pg.DB
	Log *zap.SugaredLogger
	Cfg *Config
}

// Env a global instance of SharedEnv
var Env *SharedEnv

func init() {
	if Env == nil {
		Env = new(SharedEnv)
	}
}
