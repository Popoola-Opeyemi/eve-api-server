package utils

import (
	"path/filepath"

	"github.com/go-ini/ini"
)

// Urls ...
type Urls struct {
	Base string
	Data map[string]string
}

// Load read data from ini file
func (s *Urls) Load(cfg *ini.File) {
	s.Data = cfg.Section("url").KeysHash()
}

// Path ...
func (s Urls) Path(name string) string {
	val, _ := s.Data[name]

	return val
}

// FullURL returns a path item with Base prefixed if
// the path doesn't start with a '/'
func (s Urls) FullURL(name string) string {
	path, _ := s.Data[name]
	if len(path) == 0 {
		return ""
	}

	if path[0] == '/' {
		return path
	}

	return filepath.Join(s.Base, path)
}

// Set ...
func (s *Urls) Set(name, value string) {
	s.Data[name] = value
}

// Get ...
func (s Urls) Get(name string) string {
	value, found := s.Data[name]
	if !found {
		return ""
	}

	return value
}
