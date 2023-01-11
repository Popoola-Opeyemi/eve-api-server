package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	"golang.org/x/crypto/bcrypt"
)

// Map an alias for map[string]interface{}
type Map map[string]interface{}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789" +
	"ABCDEGHIJKLMNOPQRSTUVWXYZ"

// Pager ...
func Pager(pageIndex, pageSize int64) (offset, limit int64) {
	if pageIndex <= 0 {
		pageIndex = 1
	}

	limit = pageSize
	if pageSize <= 0 {
		limit = 15
	}

	offset = (pageIndex - 1) * limit
	return
}

// DelSliceItem delete an item from a []string
func DelSliceItem(s *[]string, i int) {
	copy((*s)[i:], (*s)[i+1:]) // Shift a[i+1:] left one index.
	(*s)[len(*s)-1] = ""       // Erase last element (write zero value).
	*s = (*s)[:len(*s)-1]
}

// Atoi Converts a string to an int, returns zero if string is empty or contains
// an invalid int
func Atoi(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return i
}

// Atoi64 Converts a string to an int64, returns zero if string is empty or contains
// an invalid int64
func Atoi64(val string) int64 {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return int64(0)
	}

	return i
}

// Atoui64 Converts a string to uint64, returns zero if string is empty or contains
// an invalid uint64
func Atoui64(val string) uint64 {
	i, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return uint64(0)
	}

	return i
}

// RecToPageCount converts as record count to page count based on the provided limit
func RecToPageCount(recCount, limit uint64) (pageCount uint64) {
	pageCount = recCount / limit
	if pageCount%limit != 0 {
		pageCount++
	}

	return
}

// GetFileName extract the filename from a path and sanitize it
func GetFileName(file string) string {
	fName := filepath.Base(filepath.Clean(file))

	fName = strings.Replace(fName, " ", "", -1)

	return fName
}

// GetBareFileName extract the filename from a path and sanitize it
func GetBareFileName(file string) string {
	fName := filepath.Base(filepath.Clean(file))

	fName = strings.Replace(fName, " ", "", -1)
	fName = strings.Replace(fName, filepath.Ext(fName), "", 1)

	return fName
}

// StringWithCharset ...
func StringWithCharset(length int, charset string) string {

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]

	}
	return string(b)
}

// MakeRandText ...
func MakeRandText(length int) string {

	return StringWithCharset(length, charset)
}

// HashPassword returns a bcrypt hash of the input string
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a bcrypt hash with a string
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ParseISOTime converts YYYY-MM-DD HH:MM:SS to time.Time
func ParseISOTime(val string, end bool) time.Time {
	retv, err := time.Parse("2006-01-02 15:04:05", val)
	if err != nil {
		retv, err = time.Parse("2006-01-02", val)
		if err != nil {
			retv, err = time.Parse("20060102", val)
			if err != nil {
				return time.Time{}
			}
		}
	}

	if end {
		return time.Date(retv.Year(), retv.Month(), retv.Day(), 23, 59, 59, 0, time.Local)
	}

	return retv
}

// GetSessionValue gets an item from the session. performs check for existence
func GetSessionValue(s *sessions.Session, key string) (val interface{}, exists bool) {
	if s == nil {
		return nil, false
	}

	if _, exists = s.Values[key]; !exists {
		return nil, false
	}

	return s.Values[key], true
}

// GetSessionStrValue string typed version of GetSessionValue
func GetSessionStrValue(s *sessions.Session, key string) (val string, exists bool) {
	retv, exists := GetSessionValue(s, key)
	if !exists {
		return
	}

	return retv.(string), true
}

// GetSessionIntValue in64 typed version of GetSessionValue
func GetSessionIntValue(s *sessions.Session, key string) (val int, exists bool) {
	retv, exists := GetSessionValue(s, key)
	if !exists {
		return
	}

	return retv.(int), true
}

// GetSessionInt64Value in64 typed version of GetSessionValue
func GetSessionInt64Value(s *sessions.Session, key string) (val int64, exists bool) {
	retv, exists := GetSessionValue(s, key)
	if !exists {
		return
	}

	return retv.(int64), true
}

// ImageData ...
type ImageData struct {
	Data string `json:"data"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// SaveImageData save a data uri to a file and return a url where the file can be reached
func SaveImageData(cfg *Config, imgData json.RawMessage) (image *ImageData, err error) {

	saver := &AssetSave{}
	saver.Init(cfg.Path("assets"), cfg.URL("assets"))

	asset := &Asset{}
	image = &ImageData{}

	// convert images in json
	if len(imgData) > 0 {
		err = json.Unmarshal(imgData, image)
		if err != nil {
			return
		}

		if !strings.HasPrefix(image.Data, "data:") {
			// not a data uri
			return
		}

		// extract data uri and save to disk
		asset, err = saver.Save(&image.Data)
		if err != nil {
			return
		}

		image.Data = asset.FileURL
	}

	return
}

// DeleteImageData save a data uri to a file and return a url where the file can be reached
func DeleteImageData(cfg *Config, imgData json.RawMessage) (err error) {
	saver := &AssetSave{}
	saver.Init(cfg.Path("assets"), cfg.URL("assets"))

	asset := &Asset{}
	image := &ImageData{}

	// convert images in json
	if len(imgData) == 0 {
		return fmt.Errorf("no image data")
	}

	if err := json.Unmarshal(imgData, image); err != nil {
		return err
	}

	asset.FileURL = image.Data
	if err := saver.Delete(asset); err != nil {
		return err
	}

	return
}

// SaveJSONImageData walks through a json object and saves data ui in matching keys to file
func SaveJSONImageData(cfg *Config, key string, data *json.RawMessage) (err error) {

	jw := JayWalk{}
	if err = jw.Parse(*data); err != nil {
		return
	}

	jw.Alter(key, func(key string, node interface{}) interface{} {
		if !jw.IsMap(node) {
			return node
		}

		nBytes, err := json.Marshal(node)
		if err != nil {
			return node
		}

		img, err := SaveImageData(cfg, nBytes)
		if err != nil {
			return node
		}

		return img
	})

	// unmarshal
	*data = jw.Bytes()

	return
}

// GetSubdomain gets the subdomain by spitting the host string by "."
// where there are two parts, e.g domain.tld this function will return "www"
// where there are more than two parts, this function will return the leftmost
// part: e.g for sub.domain.tld this function will return "sub"
func GetSubdomain(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) > 2 {
		return parts[0]
	}

	return "www"
}

// ShallowMapMerge the go equivalent of Object.assign({}, obj)
// for map[string]interface{}
func ShallowMapMerge(to, from map[string]interface{}) map[string]interface{} {
	if to == nil {
		to = map[string]interface{}{}
	}

	if from == nil {
		return to
	}

	for k, v := range from {
		to[k] = v
	}

	return to
}

// URLJoin ...
func URLJoin(urls ...string) string {
	retv := make([]string, len(urls))
	for i, u := range urls {
		u = strings.Trim(u, " ")
		if len(u) == 0 {
			continue
		}

		u = strings.Trim(u, "/")
		retv[i] = u
	}

	return "/" + strings.Join(retv, "/")
}

// Getkey ...
func Getkey(section, key string) (skey string, err error) {
	cfg := Env.Cfg
	skey = cfg.Section(section).Key(key).String()

	if skey == "" {
		return "", errors.New("empty key")
	}
	return
}

func GetConfigList(section, key string) (slice []string) {

	cfg := Env.Cfg
	slice = strings.Split(cfg.Section(section).Key(key).String(), ",")

	return
}
