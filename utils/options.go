package utils

import (
	"strings"
)

// Options option map
// opt := Options{}
// opt.Split("name:mayowa|age:21", "|")
// opt["gender"] = "m"
// fmt.Println(opt["name"]) --> mayowa
// fmt.Println(opt["age"]) --> 21
// fmt.Println(opt["gender"]) --> m
// fmt.Println(opt.Int("age") --> 21
type Options map[string]interface{}

// Int return value as int else zero value
func (s Options) Int(key string) int {

	if _, ok := s[key]; !ok {
		// key not found
		return int(0)
	}

	val, isString := s[key].(string)
	if isString {
		return Atoi(val)
	}

	iVal, isInt := s[key].(int)
	if isInt {
		return iVal
	}

	return 0
}

// Int64 return value as int else zero value
func (s Options) Int64(key string) int64 {

	if _, ok := s[key]; !ok {
		// key not found
		return int64(0)
	}

	val, isString := s[key].(string)
	if isString {
		return Atoi64(val)
	}

	iVal, isInt := s[key].(int64)
	if isInt {
		return iVal
	}

	return 0
}

// String return value as int else zero value
func (s Options) String(key string) string {

	if _, ok := s[key]; !ok {
		// key not found
		return ""
	}

	val, ok := s[key].(string)
	if !ok {
		return ""
	}

	return val
}

// Split ...
func (s *Options) Split(str, sep string) {
	parts := strings.Split(str, sep)

	for _, p := range parts {
		pair := strings.SplitN(p, ":", 2)
		if len(pair) != 2 {
			continue
		}

		curVal, found := (*s)[pair[0]]
		// duplicate found
		if found {
			_, yes := curVal.([]interface{})
			// has it been converted to a list yet?
			if !yes {
				// convert to list and add both previous value
				(*s)[pair[0]] = []interface{}{}
				(*s)[pair[0]] = append((*s)[pair[0]].([]interface{}), curVal)
			}

			(*s)[pair[0]] = append((*s)[pair[0]].([]interface{}), pair[1])

		} else {
			(*s)[pair[0]] = pair[1]
		}
	}
}

// Parse an alias of Split
func (s *Options) Parse(str, sep string) {
	s.Split(str, sep)
}

// OptionsItem ...
type OptionsItem struct {
	Key   string
	Value interface{}
}

// List ...
func (s Options) List() []OptionsItem {
	output := []OptionsItem{}

	for k, v := range s {
		list, isList := s[k].([]interface{})
		if isList {
			for _, i := range list {
				retv := OptionsItem{Key: k, Value: i}
				output = append(output, retv)
			}
		} else {
			retv := OptionsItem{Key: k, Value: v}
			output = append(output, retv)
		}
	}

	return output
}
