package echotools

import (
	"gopkg.in/go-playground/validator.v9"

	"github.com/go-pg/pg"
	"github.com/labstack/echo/v4"
)

var gValidate *validator.Validate

// Validator an interface for models that can be validated by the Validate function
type Validator interface {
	Validate(*pg.DB) error
}

// ValidationError the error type returned by Validater
type ValidationError struct {
	Field  string
	ErrMsg string
}

// Error implements the Error interface
func (s ValidationError) Error() string {
	return s.ErrMsg
}

// Validate binds and validates a struct
func Validate(c echo.Context, tx *pg.DB, obj interface{}) (err error) {

	if err = c.Bind(obj); err != nil {
		return
	}

	if err = ValidateOnly(tx, obj); err != nil {
		return
	}

	return nil
}

// ValidateOnly ...
func ValidateOnly(tx *pg.DB, obj interface{}) (err error) {

	err = gValidate.Struct(obj)
	if err != nil {
		return
	}

	vObj, ok := obj.(Validator)
	if ok {
		err = vObj.Validate(tx)
		return
	}

	return nil
}
