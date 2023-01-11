package utils

import (
	"fmt"

	"github.com/fatih/structs"
	"github.com/serenize/snaker"
)

// cspell: ignore descr

// Response A map[string]interface{} type for responding to json clients
type Response struct {
	Store  map[string]interface{} `json:"store,omitempty"`
	Errors ErrMsg                 `json:"errors,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// ErrMsg holds error messages
type ErrMsg map[string]string

// Set stores an a named value in Data
func (s *Response) Set(name string, value interface{}) {
	if s.Store == nil {
		s.Store = map[string]interface{}{}
	}

	s.Store[name] = value
}

// SetErr helper for adding error messages
func (s *Response) SetErr(name, value string) {
	// create error field if it doesn't exist

	if s.Errors == nil {
		s.Errors = ErrMsg{}
	}

	s.Errors[name] = value
}

// APIError ...
func (s *Response) APIError(err error) {

	s.Error = err.Error()
}

func (s *Response) SetStore(value interface{}) (err error) {
	if s.Store == nil {
		s.Store = map[string]interface{}{}
	}

	if !structs.IsStruct(value) {
		err = fmt.Errorf("response: parameter is not a struct")
		return
	}

	stx := structs.New(value)

	// store the content of the struct in Data
	data := map[string]interface{}{}
	stx.FillMap(data)
	for k, v := range data {
		s.Store[snaker.CamelToSnake(k)] = v
	}

	return nil
}
