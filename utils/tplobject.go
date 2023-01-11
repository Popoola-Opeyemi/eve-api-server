package utils

// TplPage
/*
Every TplPage expects the following attributes
{
	"attr":{},
	"sections":{},
	"meta":[]
}
*/
type TplPage map[string]interface{}

// Attr returns attr[name]
func (s TplPage) Attr(name string) interface{} {
	retv, found := s["attr"]
	if !found {
		return ""
	}

	attr := retv.(map[string]interface{})

	val, found := attr[name]
	if !found {
		return ""
	}

	return val
}

// Section returns sections[name]
func (s TplPage) Section(name string) interface{} {
	retv, found := s["sections"]
	if !found {
		return map[string]interface{}{}
	}

	sections := retv.(map[string]interface{})

	val, found := sections[name]
	if !found {
		return map[string]interface{}{}
	}

	return val
}

// Meta returns meta[name]
func (s TplPage) Meta(idx int) interface{} {
	retv, found := s["meta"]
	if !found {
		return []interface{}{}
	}

	meta := retv.([]interface{})

	if idx >= len(meta) {
		return []interface{}{}
	}

	return meta[idx]
}
