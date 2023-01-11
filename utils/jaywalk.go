package utils

import (
	"encoding/json"
)

// JayWalk ...
type JayWalk struct {
	node interface{}
}

// Node expose s.node
func (s *JayWalk) Node() interface{} {
	return s.node
}

// Set all nodes containing key to val (node[key] = val)
func (s *JayWalk) Set(key string, val interface{}) *JayWalk {

	if s.IsMap(s.node) {
		nMap := s.node.(map[string]interface{})

		if _, found := nMap[key]; !found {
			keys := s.getKeys(nMap)
			for _, k := range keys {
				s.Branch(nMap[k]).Set(key, val)
			}

			return s
		}

		nMap[key] = val
		return s

	} else if s.IsList(s.node) {
		list := s.node.([]interface{})

		for i := 0; i < len(list); i++ {
			s.Branch(list[i]).Set(key, val)
		}
	} else {
		s.node = val
	}

	return s

}

// AlterFn a callback function for Alter()
type AlterFn func(string, interface{}) interface{}

// Alter for each node that contains "key", do node[key] = fn()
func (s *JayWalk) Alter(key string, fn AlterFn) *JayWalk {
	// log := zap.S()

	if s.IsMap(s.node) {
		// log.Debugf("node is a map")
		nMap := s.node.(map[string]interface{})

		keys := s.getKeys(nMap)
		for _, k := range keys {
			if k == key {
				// log.Debugf("key(%s) found in node", key)
				nMap[key] = fn(key, nMap[key])
				// log.Debugf("node[%s] ==>: %v", key, nMap[key])

				continue
			}

			// log.Debugf("branch on map node(%s)", k)
			s.Branch(nMap[k]).Alter(key, fn)
		}

		return s

	} else if s.IsList(s.node) {
		// log.Debugf("node is an array")
		list := s.node.([]interface{})

		for i := 0; i < len(list); i++ {
			// log.Debugf("branch on list node[%d]", i)
			s.Branch(list[i]).Alter(key, fn)
		}
	} else {
		s.node = fn(key, s.node)
	}

	return s

}

// Bytes return []byte
func (s JayWalk) Bytes() []byte {
	byt, err := json.Marshal(s.node)
	if err != nil {
		return []byte{}
	}

	return byt
}

func (s JayWalk) String() string {
	str, err := json.Marshal(s.node)
	if err != nil {
		return ""
	}

	return string(str)
}

// Seek find a key in a node of Maps
// to be refactored based on alter
func (s *JayWalk) Seek(key string) (walkers []*JayWalk) {

	walkers = []*JayWalk{}
	if s.IsMap(s.node) {
		nMap := s.node.(map[string]interface{})

		if _, found := nMap[key]; !found {
			keys := s.getKeys(nMap)
			for _, k := range keys {
				walkers = append(
					walkers,
					s.Branch(nMap[k]).Seek(key)...,
				)
			}

			return
		}

		node := s.node.(map[string]interface{})[key]
		walkers = append(walkers, s.Branch(&node))
		return

	} else if s.IsList(s.node) {
		list := s.node.([]interface{})

		for i := 0; i < len(list); i++ {
			nMap, isMap := list[i].(map[string]interface{})
			if !isMap {
				walkers = append(
					walkers,
					s.Branch(s.node.([]interface{})[i]).Seek(key)...,
				)

				continue
			}

			keys := s.getKeys(nMap)
			for _, k := range keys {
				if k == key {
					walkers = append(
						walkers, s.Branch(list[i].(map[string]interface{})[k]),
					)

					continue
				}

				walkers = append(
					walkers,
					s.Branch(list[i].(map[string]interface{})[k]).Seek(key)...,
				)
			}
		}
	}

	return
}

// Parse unmarshal json into JayWalk
func (s *JayWalk) Parse(data []byte) error {

	if err := json.Unmarshal(data, &s.node); err != nil {
		return err
	}

	return nil
}

// Branch create a new instance of JayWalk with the provided node as its root
func (s *JayWalk) Branch(node interface{}) *JayWalk {
	return &JayWalk{node: node}
}

func (s JayWalk) getKeys(node interface{}) (keys []string) {
	keys = []string{}

	mapNode, ok := s.ToMap(node)
	if !ok {
		return
	}

	for k := range mapNode {
		keys = append(keys, k)
	}

	return
}

// IsList ...
func (JayWalk) IsList(node interface{}) (ok bool) {
	_, ok = node.([]interface{})

	return
}

// IsMap ...
func (JayWalk) IsMap(node interface{}) (ok bool) {
	_, ok = node.(map[string]interface{})

	return
}

// ToList ...
func (JayWalk) ToList(node interface{}) (listNode []interface{}, ok bool) {
	listNode, ok = node.([]interface{})

	return
}

// ToMap ...
func (JayWalk) ToMap(node interface{}) (mapNode map[string]interface{}, ok bool) {
	mapNode, ok = node.(map[string]interface{})

	return
}
