//
// session.go
// helper functions to simplify session management
// Copyright 2017 Akinmayowa Akinyemi
//

package echotools

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"

	"github.com/labstack/echo/v4"
)

// SessionMgr utility to manage sessions
type SessionMgr struct {
	session *sessions.Session
	ctx     echo.Context
}

// NewSessionMgr create an instance of SessionMgr
func NewSessionMgr(c echo.Context, sessName string, reset ...bool) (sMgr *SessionMgr, err error) {
	sMgr = new(SessionMgr)
	if len(sessName) == 0 {
		sessName = "session"
	}

	sMgr.session, err = session.Get(sessName, c)
	if err != nil {
		return nil, err
	}

	if len(reset) > 0 && reset[0] == true {
		sMgr.session, err = sMgr.session.Store().New(c.Request(), sessName)

		if err != nil {
			return nil, err
		}
	}

	sMgr.ctx = c

	return
}

// Set store a value into the session
func (s *SessionMgr) Set(key string, value interface{}) {
	s.session.Values[key] = value
}

func (s *SessionMgr) MaxAge(value int) {
	s.session.Options.MaxAge = value
}

// Value gets an item from the session. performs check for existence
func (s SessionMgr) Value(key string) (val interface{}, exists bool) {

	if _, exists = s.session.Values[key]; !exists {
		return nil, false
	}

	return s.session.Values[key], true
}

// Save ...
func (s *SessionMgr) Save() error {
	return s.session.Save(s.ctx.Request(), s.ctx.Response())
}

// StringValue string typed version of Value
func (s SessionMgr) StringValue(key string) (val string, exists bool) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	return retv.(string), true
}

// String ...
func (s SessionMgr) String(key string) (val string) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	val, _ = retv.(string)

	return
}

// IntValue int typed version of Value
func (s SessionMgr) IntValue(key string) (val int, exists bool) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	return retv.(int), true
}

// Int ...
func (s SessionMgr) Int(key string) (val int) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	val, _ = retv.(int)
	return
}

// Int64Value an int64 typed version of Value
func (s SessionMgr) Int64Value(key string) (val int64, exists bool) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	return retv.(int64), true
}

// Int64 ...
func (s SessionMgr) Int64(key string) (val int64) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	val, _ = retv.(int64)
	return
}

// BoolValue an bool typed version of Value
func (s SessionMgr) BoolValue(key string) (val bool, exists bool) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	return retv.(bool), true
}

// Bool ...
func (s SessionMgr) Bool(key string) (val bool) {
	retv, exists := s.Value(key)
	if !exists {
		return
	}

	val, _ = retv.(bool)
	return
}
