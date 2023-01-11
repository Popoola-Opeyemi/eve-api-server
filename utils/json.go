package utils

import (
	"encoding/json"
)

// JSONMerge ...
func JSONMerge(src, defa json.RawMessage) (json.RawMessage, error) {

	sMap := map[string]interface{}{}
	dMap := map[string]interface{}{}

	err := json.Unmarshal(src, &sMap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(defa, &dMap)
	if err != nil {
		return nil, err
	}

	// find items in default not in src and set it
	for k, v := range dMap {
		_, found := sMap[k]
		if !found {
			sMap[k] = v
		}
	}

	retv := json.RawMessage{}
	retv, err = json.Marshal(sMap)
	if err != nil {
		return nil, err
	}

	return retv, nil
}
