package utils

import (
	"errors"
	"reflect"
	"sync"
)

// TypeRegister ...
type TypeRegister struct {
	list map[string]reflect.Type
	gate sync.Mutex
}

var typeRegistry TypeRegister

func init() {
	typeRegistry = TypeRegister{}
	typeRegistry.list = make(map[string]reflect.Type)
}

// Set ...
func (s *TypeRegister) Set(item interface{}) {
	s.gate.Lock()
	defer s.gate.Unlock()

	vType := reflect.TypeOf(item)
	if vType.Kind().String() == "ptr" {
		vType = vType.Elem()
	}

	name := vType.Name()
	if _, ok := s.list[name]; ok {
		// this type has already been registered
		return
	}

	s.list[name] = vType
}

// SetName ...
func (s *TypeRegister) SetName(name string, item interface{}) {
	s.gate.Lock()
	defer s.gate.Unlock()

	vType := reflect.TypeOf(item)
	if vType.Kind().String() == "ptr" {
		vType = vType.Elem()
	}

	if _, ok := s.list[name]; ok {
		// this type has already been registered
		return
	}

	s.list[name] = vType
}

// Get ...
func (s *TypeRegister) Get(name string) (interface{}, error) {
	s.gate.Lock()
	defer s.gate.Unlock()

	if typ, ok := s.list[name]; ok {
		return reflect.New(typ).Elem().Interface(), nil
	}

	return nil, errors.New("unknown type " + name)
}

// GetPointer ...
func (s *TypeRegister) GetPointer(name string) (interface{}, error) {
	s.gate.Lock()
	defer s.gate.Unlock()

	if typ, ok := s.list[name]; ok {
		return reflect.New(typ).Interface(), nil
	}

	return nil, errors.New("unknown type " + name)
}

// GetSlice ...
func (s *TypeRegister) GetSlice(name string) (interface{}, error) {
	s.gate.Lock()
	defer s.gate.Unlock()

	if typ, ok := s.list[name]; ok {
		return reflect.MakeSlice(reflect.SliceOf(typ), 0, 0).Interface(), nil
	}

	return nil, errors.New("unknown type " + name)
}

// GetSlicePointer ...
func (s *TypeRegister) GetSlicePointer(name string) (interface{}, error) {
	s.gate.Lock()
	defer s.gate.Unlock()

	if typ, ok := s.list[name]; ok {
		return reflect.New(reflect.SliceOf(typ)).Interface(), nil
	}

	return nil, errors.New("unknown type " + name)
}

// RegisterType ...
func RegisterType(t interface{}) {
	typeRegistry.Set(t)
}

// RegisterTypeWithName ...
func RegisterTypeWithName(name string, t interface{}) {
	typeRegistry.SetName(name, t)
}

// MakeType ...
func MakeType(name string) (interface{}, error) {
	return typeRegistry.Get(name)
}

// MakePointerType ...
func MakePointerType(name string) (interface{}, error) {
	return typeRegistry.GetPointer(name)
}

// MakeSliceType ...
func MakeSliceType(name string) (interface{}, error) {
	return typeRegistry.GetSlice(name)
}

// MakeSlicePointerType ...
func MakeSlicePointerType(name string) (interface{}, error) {
	return typeRegistry.GetSlicePointer(name)
}

// SetStructField set the value of an attribute of a struct "programmatically"
// e.g
// type Foo struct {
// 	ID string
// }
// foo := &Foo{}
// SetStructField(foo, "ID", "1234")
func SetStructField(item interface{}, field string, value interface{}) {
	vItem := reflect.ValueOf(item)
	if vItem.Kind().String() == "ptr" {
		vItem = reflect.Indirect(vItem)
	}

	// note if this panics check to see if the value of field is an actual attribute
	vItem.FieldByName(field).Set(reflect.ValueOf(value))

	return

}

// GetStructField returns a field of a struct as reflect.Value
func GetStructField(item interface{}, field string) reflect.Value {
	vItem := reflect.ValueOf(item)
	if vItem.Kind().String() == "ptr" {
		vItem = vItem.Elem()
	}

	return vItem.FieldByName(field)
}

// ListStructFields returns the fields of a struct as []string. fields can be excluded
// using the exclude parameter
func ListStructFields(item interface{}, exclude []string, fmtUnderScore bool) (fields []string) {
	vType := reflect.TypeOf(item)
	if vType.Kind().String() == "ptr" {
		vType = vType.Elem()
	}

	fldCount := vType.NumField()

	fields = []string{}
	for i := 0; i < fldCount; i++ {
		fldName := vType.Field(i).Name

		if InStringSlice(fldName, exclude) {
			continue
		}

		if fmtUnderScore {
			fields = append(fields, Underscore(fldName))
		} else {
			fields = append(fields, fldName)
		}
	}

	return
}

// InStringSlice find str in []string and return true if found
func InStringSlice(str string, slice []string) bool {

	for _, s := range slice {
		if str == s {
			return true
		}
	}

	return false
}

// StructHasField ...
func StructHasField(item interface{}, name string) bool {
	vType := reflect.TypeOf(item)
	if vType.Kind().String() == "ptr" {
		vType = vType.Elem()
	}

	_, found := vType.FieldByName(name)
	return found
}

// IfToInt return value as int else zero value
func IfToInt(v interface{}) int {
	val, isString := v.(string)
	if isString {
		return Atoi(val)
	}

	iVal, isInt := v.(int)
	if isInt {
		return iVal
	}

	return 0
}

// IfToInt64 return value as int else zero value
func IfToInt64(v interface{}) int64 {

	val, isString := v.(string)
	if isString {
		return Atoi64(val)
	}

	iVal, isInt := v.(int64)
	if isInt {
		return iVal
	}

	return 0
}

// IfToString return value as int else zero value
func IfToString(v interface{}) string {
	val, ok := v.(string)
	if !ok {
		return ""
	}

	return val
}
