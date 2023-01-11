package utils

// Copyright (c) 2013 github.com/go-pg/pg Authors. All rights reserved.

// IsUpper ...
func IsUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

// IsLower ...
func IsLower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

// ToUpper ...
func ToUpper(c byte) byte {
	return c - 32
}

// ToLower ...
func ToLower(c byte) byte {
	return c + 32
}

// Underscore converts "CamelCasedString" to "camel_cased_string".
func Underscore(s string) string {
	r := make([]byte, 0, len(s)+5)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsUpper(c) {
			if i > 0 && i+1 < len(s) && (IsLower(s[i-1]) || IsLower(s[i+1])) {
				r = append(r, '_', ToLower(c))
			} else {
				r = append(r, ToLower(c))
			}
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}

// CamelCased converts "under_score_string" to "UnderScoreString"
func CamelCased(s string) string {
	r := make([]byte, 0, len(s))
	upperNext := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			upperNext = true
			continue
		}
		if upperNext {
			if IsLower(c) {
				c = ToUpper(c)
			}
			upperNext = false
		}
		r = append(r, c)
	}
	return string(r)
}

// ToExported ...
func ToExported(s string) string {
	if len(s) == 0 {
		return s
	}
	if c := s[0]; IsLower(c) {
		b := []byte(s)
		b[0] = ToUpper(c)
		return string(b)
	}
	return s
}

// UpperString ...
func UpperString(s string) string {
	if isUpperString(s) {
		return s
	}

	b := make([]byte, len(s))
	for i := range b {
		c := s[i]
		if IsLower(c) {
			c = ToUpper(c)
		}
		b[i] = c
	}
	return string(b)
}

/// isUpperString ...
func isUpperString(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsLower(c) {
			return false
		}
	}
	return true
}
