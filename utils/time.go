package utils

// cspell: ignore formatYYYYMMDD, FormatYYYYMMDDHHmmSS

import (
	"fmt"
	"strings"
	"time"
)

// FormatYYYYMMDDHHmmSS ...
const FormatYYYYMMDDHHmmSS = "2006-01-02 15:04:05"
const formatYYYYMMDD = "2006-01-02"

const OneYearINSeconds int = 86400 * 30 * 12

// DateTime wraps time.Time to handle custom parsing
type DateTime struct {
	time.Time
}

// NewDateTime ...
func NewDateTime(t time.Time) DateTime {
	return DateTime{Time: t}
}

// Now ...
func (DateTime) Now() DateTime {
	return NewDateTime(time.Now())
}

// MarshalJSON implements the json.Marshaler interface
func (s DateTime) MarshalJSON() ([]byte, error) {
	retv := s.Format(FormatYYYYMMDDHHmmSS)
	return []byte(fmt.Sprintf("\"%s\"", retv)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (s *DateTime) UnmarshalJSON(val []byte) error {
	var err error

	sv := strings.Trim(string(val), "\"")
	if len(sv) < 1 {
		return nil
	}

	if len(sv) == 10 {
		s.Time, err = time.Parse(formatYYYYMMDD, sv)
	} else {
		s.Time, err = time.Parse(FormatYYYYMMDDHHmmSS, sv)
	}
	if err != nil {
		return err
	}

	return nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (s *DateTime) Scan(value interface{}) error {
	var err error

	if value == nil {
		return nil
	}

	raw := value.([]uint8)
	strVal := string([]byte(raw))
	// Env.Log.Debug("scanner -->>", strVal)

	if len(strVal) < 1 {
		return nil
	}

	if len(strVal) == 10 {
		s.Time, err = time.Parse(formatYYYYMMDD, strVal)
	} else {
		s.Time, err = time.Parse(FormatYYYYMMDDHHmmSS, strVal)
	}
	if err != nil {
		return err
	}

	return nil
}
