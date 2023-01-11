package utils

import (
	"path/filepath"

	"github.com/go-ini/ini"
)

// Config a wrapper of go-ini that provides usefull helpers
type Config struct {
	*ini.File
}

// NewConfig create new instance of Config
func NewConfig(cf *ini.File) *Config {
	cfg := &Config{cf}

	return cfg
}

// Path ...
func (s Config) Path(name string) string {
	data := s.Section("path").KeysHash()
	value, found := data[name]
	if !found {
		return ""
	}

	return value
}

// FullPath returns a path item with base_url prefixed if
// the path doesn't start with a '/'
func (s Config) FullPath(name string) string {
	data := s.Section("path").KeysHash()
	path, _ := data[name]
	if len(path) == 0 {
		return ""
	}

	if path[0] == '/' {
		return path
	}

	// get base_path, if empty use workdir
	base := s.Section("").KeysHash()
	basePath, _ := base["base_path"]
	if len(basePath) == 0 {
		basePath, _ = base["workdir"]
	}

	return filepath.Join(basePath, path)
}

// URL ...
func (s Config) URL(name string) string {
	data := s.Section("url").KeysHash()
	value, found := data[name]
	if !found {
		return ""
	}

	return value
}

// FullURL returns a path item with base_url prefixed if
// the path doesn't start with a '/'
func (s Config) FullURL(name string) string {
	data := s.Section("path").KeysHash()
	path, _ := data[name]
	if len(path) == 0 {
		return ""
	}

	if path[0] == '/' {
		return path
	}

	// get base_url
	base := s.Section("").KeysHash()
	baseURL, _ := base["base_url"]

	return filepath.Join(baseURL, path)
}
