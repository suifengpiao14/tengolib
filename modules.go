package tengolib

import (
	"github.com/d5/tengo/v2"
	"github.com/suifengpiao14/tengolib/tengocollection"
	"github.com/suifengpiao14/tengolib/tengocontext"
	"github.com/suifengpiao14/tengolib/tengogsjson"
)

var BuiltinModules = map[string]map[string]tengo.Object{
	"gsjson":  tengogsjson.GSjson,
	"context": tengocontext.Ctx,
}

var SourceModules = map[string]string{
	"collection": tengocollection.Tengocollection,
}

// AllModuleNames returns a list of all default module names.
func AllModuleNames() []string {
	var names []string
	for name := range BuiltinModules {
		names = append(names, name)
	}
	for name := range SourceModules {
		names = append(names, name)
	}
	return names
}

// GetModuleMap returns the module map that includes all modules
// for the given module names.
func GetModuleMap(names ...string) *tengo.ModuleMap {
	modules := tengo.NewModuleMap()
	for _, name := range names {
		if mod := BuiltinModules[name]; mod != nil {
			modules.AddBuiltinModule(name, mod)
		}
		if mod := SourceModules[name]; mod != "" {
			modules.AddSourceModule(name, []byte(mod))
		}
	}
	return modules
}
