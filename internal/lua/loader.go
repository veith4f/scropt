package lua

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type Loader struct {
	modules map[string]string
	state   *lua.LState
}

func (l *Loader) loadModule(moduleName string) error {

	pkg := l.state.GetGlobal("package").(*lua.LTable)
	loaded := pkg.RawGetString("loaded").(*lua.LTable)

	// Check if module is already loaded
	if module := loaded.RawGetString(moduleName); module != lua.LNil {
		l.state.Push(module.(*lua.LTable))
		return nil
	}

	if file, ok := l.modules[moduleName]; ok {
		// .lua file
		if strings.HasSuffix(file, ".lua") {
			if err := l.state.DoFile(file); err != nil {
				return fmt.Errorf("Error loading module %s: %v", moduleName, err)
			}
			module := l.state.Get(-1)
			loaded.RawSetString(moduleName, module)
			l.state.Push(module)
			return nil
		}
		// .so file
		plug, err := plugin.Open(file)
		if err != nil {
			return err
		}

		// Look for a function called luaopen_<modulename>
		openFunc, err := plug.Lookup("luaopen_" + strings.ReplaceAll(moduleName, ".", "_"))
		if err != nil {
			return fmt.Errorf("no luaopen_%s function found in %s", moduleName, file)
		}

		// Call the function and push result onto the stack
		luaFunc, ok := openFunc.(func(*lua.LState) int)
		if !ok {
			return fmt.Errorf("luaopen_%s has wrong function signature", moduleName)
		}

		// Call module initializer
		luaFunc(l.state)

		return nil
	}

	return fmt.Errorf("Module not found: %s", moduleName)
}

func (l *Loader) registerModules(suffix string, pathVar string) {
	pwd, _ := os.Getwd()

	pathArr := strings.Split(strings.ReplaceAll(strings.ReplaceAll(os.Getenv(pathVar), "?"+suffix, ""),
		"?/init"+suffix, "")+";"+pwd+"/", ";")

	// eliminate duplicates and remove relative paths
	dirs := make(map[string]struct{})
	var path string
	for _, path = range pathArr {
		if strings.HasPrefix(path, "/") {
			dirs[path] = struct{}{}
		}
	}

	for dir, _ := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if the file has the suffix's extension
			if !info.IsDir() && strings.HasSuffix(info.Name(), suffix) {
				var moduleName string
				if info.Name() == "init"+suffix {
					moduleName = strings.ReplaceAll(
						strings.TrimSuffix(strings.TrimPrefix(path, dir), "/"+info.Name()),
						"/", ".")
				} else {
					moduleName = strings.ReplaceAll(
						strings.TrimPrefix(strings.TrimSuffix(path, suffix), dir),
						"/", ".")
				}
				l.modules[moduleName] = path
			}
			return nil
		})
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (l *Loader) preload(preload map[string]lua.LValue) {
	pkg := l.state.GetGlobal("package").(*lua.LTable)
	preloaded := l.state.GetField(pkg, "preload").(*lua.LTable)
	for k, v := range preload {
		fmt.Printf("Preloading %s\n", k)
		preloaded.RawSetString(k, v)
	}
}

func withLoader(L *lua.LState, preload ...map[string]lua.LValue) *Loader {
	loader := &Loader{
		modules: make(map[string]string),
		state:   L,
	}
	if len(preload) > 0 {
		for _, p := range preload {
			fmt.Printf("Preloading %v\n", p)
			loader.preload(p)
		}
	}
	loader.registerModules(".lua", "LUA_PATH")
	//loader.registerModules(".so", "LUA_CPATH")

	require := func(L *lua.LState) int {
		moduleName := L.ToString(1)
		if err := loader.loadModule(moduleName); err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}
		return 1
	}
	L.SetGlobal("require", L.NewFunction(require))

	return loader
}
