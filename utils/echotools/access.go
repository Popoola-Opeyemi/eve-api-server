// File: access.go
// File Created: Saturday, 20th July 2019 12:40:44 am
// Author: Akinmayowa Akinyemi
// -----
// Copyright 2019 Techne Efx Ltd

// cspell: ignore sess

package echotools

import (
	"net/http"
	"strings"
	"sync"

	"eve/utils"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

//AccessPermission ...
type AccessPermission int

const (
	// PermissionNone ...
	PermissionNone AccessPermission = iota
	// PermissionDenied ...
	PermissionDenied
	// PermissionReadOnly ...
	PermissionReadOnly
	// PermissionReadWrite ...
	PermissionReadWrite
	// PermissionReadWriteDelete ...
	PermissionReadWriteDelete
	// PermissionAll ...
	PermissionAll
)

// AccessRole ...
type AccessRole int

const (
	// RoleEveryone ...
	RoleEveryone AccessRole = iota
	// RoleUser ...
	RoleUser
	// RoleSupervisor ...
	RoleSupervisor
	// RoleManager ...
	RoleManager
	// RoleSuperUser ...
	RoleSuperUser
)

// AccessRule ...
type AccessRule struct {
	Path       string
	Role       AccessRole
	Permission AccessPermission
}

// AccessMgr ...
type AccessMgr struct {
	mtx   *sync.Mutex
	rules []AccessRule
}

// NewAccessMgr returns an instance of AccessMgr
func NewAccessMgr() *AccessMgr {
	mgr := &AccessMgr{
		mtx:   &sync.Mutex{},
		rules: []AccessRule{},
	}

	return mgr
}

// AccessControllerOptions ...
type AccessControllerOptions struct {
	// name used to store the users role in the session
	RoleField   string
	RedirectURL string
	// name used to store the users siteID in the session
	SiteIDField string
	// name used to get siteID from c.Param()
	SiteIDParam string
}

// AccessController midleware
func AccessController(mgr *AccessMgr, log *zap.SugaredLogger, opts AccessControllerOptions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			method := c.Request().Method

			roleFld := opts.RoleField
			redirect := opts.RedirectURL
			sidFld := opts.SiteIDField

			if path == redirect {
				return next(c)
			}

			rule := mgr.Find(path)
			if rule.Permission == PermissionNone {
				// Access to controller managed handlers must be specified or
				// access is denied
				log.Errorf("path not found: access denied - %s", path)
				return ErrResponse(c, "access denied", redirect)
			}

			sess, err := NewSessionMgr(c, "session")
			if err != nil {
				// c.Error(err)
				log.Error(err)
				return ErrResponse(c, "access denied", redirect)
			}

			// put siteID into context
			sSid := sess.String(sidFld)
			if len(sSid) > 0 {
				c.Set("siteID", sSid)
			}

			role := AccessRole(sess.Int(roleFld))

			if method == "GET" && mgr.HasAccess(path, role) == false {
				log.Errorf(" access denied for path: %s, role(%d)", path, role)

				return ErrResponse(c, "access denied", redirect)
			}

			if (method == "POST" || method == "PUT") && mgr.HasWriteAccess(path, role) == false {
				log.Errorf(" access denied(RW) for path: %s", path)
				return ErrResponse(c, "access denied", redirect)
			}

			if (method == "DELETE") && mgr.HasDeleteAccess(path, role) == false {
				log.Errorf(" access denied(Del) for path: %s", path)
				return ErrResponse(c, "access denied", redirect)
			}

			return next(c)
		}
	}
}

// ErrResponse ...
func ErrResponse(c echo.Context, msg, redirect string) error {
	log := zap.S()
	if IsAjax(c) {
		log.Errorf("isAjax: %s", msg)
		return c.JSON(http.StatusForbidden, utils.Map{"error": msg})
	}

	if len(redirect) > 0 {
		c.Redirect(http.StatusTemporaryRedirect, redirect)
	} else {
		return c.String(http.StatusForbidden, msg)
	}

	return nil
}

// AddRule a rule to the rules list
func (s *AccessMgr) AddRule(rule AccessRule) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.rules = append(s.rules, rule)
}

// AddRules ...
func (s *AccessMgr) AddRules(rules []AccessRule) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.rules = append(s.rules, rules...)
}

// Find a rule based on a match from the path, if match not found
// returns AccessRule{RoleEveryone, PermissionNone}
func (s *AccessMgr) Find(path string) AccessRule {
	// log := zap.S()

	for _, i := range s.rules {
		// log.Debug(i, path)
		if strings.HasPrefix(path, i.Path) {
			// log.Debug("path found", i.Path)
			return i
		}
	}

	return AccessRule{
		Path:       path,
		Role:       RoleEveryone,
		Permission: PermissionNone,
	}
}

// HasXXXX functions find a rule that matches the supplied path.
// If the path isn't found in the rules list access is denied (the functions return false).
// if the supplied role is greater than the role specified in the matching rule,
// access is granted

// HasAccess returns true if the requested role is a match or is greater
// then the role specified in the rule
func (s AccessMgr) HasAccess(path string, role AccessRole) bool {

	rule := s.Find(path)
	if role >= rule.Role {
		return true
	}

	return false
}

// HasReadAccess returns true if the supplied role is a is greater than or equal
// to the role specified in the rule. If the supplied role == rule.role and
// rule.permission >= PermissionReadonly readonly access is granted
func (s AccessMgr) HasReadAccess(path string, role AccessRole) bool {
	// role = RoleManager and rule == AccessRule{RoleAdmin, PermissionReadOnly}
	//
	rule := s.Find(path)
	if role > rule.Role {
		return true
	} else if role == rule.Role && rule.Permission >= PermissionReadOnly {
		return true
	}

	return false
}

// HasWriteAccess returns true if the supplied role is a is greater than or equal
// to the role specified in the rule. If the supplied role == rule.role and
// rule.permission >= PermissionReadWrite write access is granted
func (s AccessMgr) HasWriteAccess(path string, role AccessRole) bool {

	rule := s.Find(path)
	if role > rule.Role {
		return true
	} else if role == rule.Role && rule.Permission >= PermissionReadWrite {
		return true
	}

	return false
}

// HasDeleteAccess returns true if the supplied role is a is greater than or equal
// to the role specified in the rule. If the supplied role == rule.role and
// rule.permission >= PermissionReadWriteDelete delte access is granted
func (s AccessMgr) HasDeleteAccess(path string, role AccessRole) bool {

	rule := s.Find(path)
	if role > rule.Role {
		return true
	} else if role == rule.Role && rule.Permission >= PermissionReadWriteDelete {
		return true
	}

	return false
}
