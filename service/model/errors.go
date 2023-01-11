package model

import (
	"github.com/go-pg/pg"
)

type DuplicateError struct {
	Msg string
}

func (s DuplicateError) Error() string {
	return s.Msg
}

type ValidationError struct {
	Msg string
}

func (s ValidationError) Error() string {
	return s.Msg
}

var RecordNotFound = pg.ErrNoRows
