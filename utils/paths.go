package utils

import (
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
)

// Paths ...
type Paths struct {
	Base string
	Data map[string]string
}

// Load read data from ini file
func (s *Paths) Load(cfg *ini.File) {
	s.Data = cfg.Section("path").KeysHash()
}

// Path ...
func (s Paths) Path(name string) string {
	val, _ := s.Data[name]

	return val
}

// FullPath returns a path item with Base prefixed if
// the path doesn't start with a '/'
func (s Paths) FullPath(name string) string {
	path, _ := s.Data[name]
	if path[0] == '/' {
		return path
	}

	return filepath.Join(s.Base, path)
}

// Set ...
func (s *Paths) Set(name, value string) {
	s.Data[name] = value
}

// Get ...
func (s Paths) Get(name string) string {
	value, found := s.Data[name]
	if !found {
		return ""
	}

	return value
}

func PathInfo(path string) (info os.FileInfo, err error) {

	info, err = os.Stat(path)
	if err != nil {
		return
	}

	return
}

// IsDir checks if path is a directory
func IsDir(path string) (bool, error) {

	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

// Exists determines if a path exists
func Exists(path string) (bool, error) {

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

// Available determines if a path exists on error returns false
func Available(path string) bool {
	retv, _ := Exists(path)
	return retv
}
