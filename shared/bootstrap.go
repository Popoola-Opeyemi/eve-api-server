package shared

// cspell: ignore curwd, cfgpath, sysdba, dbname

import (
	"fmt"
	"os"

	"eve/utils"

	"bitbucket.org/mayowa/helpers/config"
	"bitbucket.org/mayowa/helpers/path"

	"github.com/go-ini/ini"
	"go.uber.org/zap"
)

// CreateDefaultConfig ...
func CreateDefaultConfig(name, curwd, cfgPath string, log *zap.SugaredLogger) (cfg *utils.Config, err error) {
	log.Debug("creating default confing in: ", cfgPath)

	cfgStr := fmt.Sprintf(`
		workdir = %s
		cfgpath = %s
		host =
		port = 4000

		cors_origin = "http://localhost:3000"
		base_url =
		base_path =

		[url]

		[path]
		templates = templates

		[mail]
		smtp_host =
		smtp_port =
		smtp_from =
		smtp_user =
		smtp_password =
		# EncryptionNone = 0, EncryptionSSL = 1, EncryptionTLS = 2
		smtp_encryption = 0
		# 	AuthPlain = 0, AuthLogin = 1, AuthCRAMMD5 = 2
		smtp_auth = 0

		[db]
		driver   = postgres
		host     = localhost:5432
		user     = sysdba
		password = masterkey
		dbname   = eve
		sslmode  = disable
		`,
		curwd, cfgPath,
	)

	cf := &ini.File{}
	if cf, _, err = config.Create(name, cfgStr); err != nil {
		log.Error(err)
		return
	}

	cfg = utils.NewConfig((cf))

	return
}

// InitConfig load config file and create defaults if config not found
func InitConfig(appName string, log *zap.SugaredLogger) (cfg *utils.Config, cfgPath string, err error) {
	cf := &ini.File{}

	cfgPath, err = config.GetPath(appName)
	if err != nil {
		return
	}

	curWd, err := os.Getwd()
	if err != nil {
		return
	}

	// found a config location
	cf, cfgPath, err = config.Load(appName)
	if err != nil {
		if len(cfgPath) == 0 {
			log.Error(err)
			return
		}

		if cfg, err = CreateDefaultConfig(appName, curWd, cfgPath, log); err != nil {
			log.Error(err)
			return
		}
	} else {
		log.Debug("working directory: ", curWd)
		log.Debug("loading config from: ", cfgPath)
		cfg = utils.NewConfig(cf)
	}

	log.Debugf("CORS allowed-origins: %s", cfg.Section("").Key("cors_origin").Strings(","))
	log.Debugf("base_url: %s", cfg.Section("").Key("base_url").String())
	log.Debugf("public_url: %s", cfg.Section("").Key("public_url").String())

	// check if paths exist
	paths := cfg.Section("path").KeysHash()
	for k, pth := range paths {
		if !path.Available(pth) {
			log.Errorf("path[%s]: %s not accessible", k, pth)
			if err := os.MkdirAll(pth, 0700); err != nil {
				log.Errorf("path[%s]: %s cannot create folder", k, pth)
			}
		} else {
			log.Debugf("path[%s]: %s", k, pth)
		}
	}

	return
}
