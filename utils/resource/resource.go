// File: resource.go
// File Created: Sunday, 14th July 2019 8:06:29 am
// Author: Akinmayowa Akinyemi
// -----
// Copyright 2019 Techne Efx Ltd

package resource

import "eve/utils"

// cspell: ignore fkey, fkvalue

// Init ...
func Init(env *utils.SharedEnv) {
	if env != nil {
		utils.Env = new(utils.SharedEnv)
		utils.Env.Db = env.Db
		utils.Env.Log = env.Log
	}

	if registry == nil {
		registry = new(Registry)
	}

}
