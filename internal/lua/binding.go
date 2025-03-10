package lua

import (
	"errors"
	"fmt"
	"go/types"
	"reflect"

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/tools/go/packages"
)

// Detect if a Lua table is an array
func isArrayTable(tbl *lua.LTable) bool {
	var hasNonIntegerKey bool
	tbl.ForEach(func(key lua.LValue, _ lua.LValue) {
		if _, ok := key.(lua.LNumber); !ok {
			hasNonIntegerKey = true
		}
	})
	return !hasNonIntegerKey
}

func addFunction(L *lua.LState, namespace *lua.LTable, name string, fn reflect.Value) {
	handler := func(L *lua.LState) int {
		args := make([]reflect.Value, L.GetTop())
		for i := 1; i <= len(args); i++ {
			args[i-1] = reflect.ValueOf(luaValToGo(L.CheckAny(i)))
		}

		results := fn.Call(args)
		nresults := len(results)

		for i := 0; i < nresults; i++ {
			L.Push(goValToLua(L, results[i]))
		}
		return len(results)
	}
	if namespace == nil {
		L.SetGlobal(name, L.NewFunction(handler))
	} else {
		L.SetField(namespace, name, L.NewFunction(handler))
	}
}

func addType(L *lua.LState, namespace *lua.LTable, typ reflect.Type) {
	// Create a new table to represent the class
	class := L.NewTable()
	L.SetField(namespace, typ.Name(), class)

	// Define the "get" method to return a table containing all fields
	L.SetField(class, "get", L.NewFunction(func(L *lua.LState) int {

		nargs := L.GetTop()
		if nargs != 1 {
			L.RaiseError("Expected no arguments to function get")
			return 0
		}

		// Get the instance (self)
		self := luaValToGo(L.Get(1))
		if self == nil {
			L.RaiseError("Invalid object")
			return 0
		}

		structValue := reflect.ValueOf(self)
		if structValue.Kind() == reflect.Ptr {
			structValue = structValue.Elem()
		}
		structType := structValue.Type()

		// Create a table to store field values
		result := L.NewTable()

		// Iterate over struct fields and store them in the Lua table
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			fieldValue := structValue.Field(i)
			if fieldValue.CanInterface() {
				fmt.Printf("returned map has field: %s\n", field.Name)
				result.RawSetString(field.Name, goValToLua(L, fieldValue))
			}
		}

		// Push the result table to Lua
		L.Push(result)
		return 1
	}))

	// Define the "set" method to update fields from a Lua table
	L.SetField(class, "set", L.NewFunction(func(L *lua.LState) int {
		// Get the instance (self)
		self := L.CheckUserData(1)
		if self == nil {
			L.RaiseError("Invalid object")
			return 0
		}

		// Get the Lua table containing new values
		luaTable := L.CheckTable(2)
		if luaTable == nil {
			L.RaiseError("Expected a table as argument")
			return 0
		}

		// Get the underlying Go struct value
		structValue := reflect.ValueOf(self.Value).Elem()

		// Iterate over struct fields and update them if present in the table
		luaTable.ForEach(func(key, value lua.LValue) {
			if keyStr, ok := key.(lua.LString); ok {
				fieldName := string(keyStr)
				field := structValue.FieldByName(fieldName)

				if field.IsValid() && field.CanSet() {
					goVal := luaValToGo(value)

					switch field.Kind() {
					case reflect.String:
						if v, ok := goVal.(string); ok {
							field.SetString(v)
						}
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						if v, ok := goVal.(int64); ok {
							field.SetInt(v)
						}
					case reflect.Float32, reflect.Float64:
						if v, ok := goVal.(float64); ok {
							field.SetFloat(v)
						}
					case reflect.Bool:
						if v, ok := goVal.(bool); ok {
							field.SetBool(v)
						}
					default:
						L.RaiseError("Unsupported field type: %s", field.Kind())
					}
				}
			}
		})

		return 0
	}))

	// Define the "new" method for this class
	L.SetField(class, "new", L.NewFunction(func(L *lua.LState) int {
		// Ensure the number of arguments is either 1 (self) or 2 (self + table)
		nargs := L.GetTop()
		if nargs != 1 && nargs != 2 {
			L.RaiseError("Expected no arguments or exactly one table for 'new', but got %d.", nargs-1)
			return 0
		}

		// Create a new instance of the type (pointer to the struct)
		reflectVal := reflect.New(typ) // *T (pointer to struct)
		instance := goValToLua(L, reflectVal).(*lua.LTable)

		// Set the metatable for the new instance
		L.SetMetatable(instance, class)
		L.SetField(class, "__index", class)

		// If an initialization table is provided, use "set" to populate fields
		if nargs == 2 {
			L.Push(instance) // Push the instance (self)
			L.Push(L.Get(2)) // Push the table (argument 2)
			L.CallByParam(lua.P{
				Fn:      L.GetField(class, "set"),
				NRet:    0,
				Protect: true,
			})
		}

		// Push the new instance to the Lua stack
		L.Push(instance)
		return 1
	}))

}

func addTypes(L *lua.LState, namespace *lua.LTable, pkg string) {

	cfg := packages.Config{Mode: packages.NeedTypes}
	pkgs, _ := packages.Load(&cfg, pkg)

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)

			if fn, ok := obj.(*types.Func); ok {
				if reflectType, isKnown := GetRegistry().Lookup(fn.FullName()); isKnown {
					addFunction(L, namespace, fn.Name(), reflect.New(reflectType))
				}
			} else if reflectType, isKnown := GetRegistry().Lookup(obj.Type().String()); isKnown {
				addType(L, namespace, reflectType)
			}
		}
	}
}

func addObject(L *lua.LState, objName string, obj reflect.Value) error {
	typ := obj.Type()
	if typ.Kind() != reflect.Ptr {
		return errors.New("Object must be pointer: " + objName)
	}

	namespace := goValToLua(L, obj).(*lua.LTable)
	L.SetGlobal(objName, namespace)

	for i := 0; i < typ.NumMethod(); i++ {
		methodName := typ.Method(i).Name
		method := obj.MethodByName(methodName)

		if method.IsValid() {
			L.SetField(namespace, methodName, L.NewFunction(func(L *lua.LState) int {
				margs := method.Type().NumIn()
				largs := L.GetTop()
				if margs != largs {
					L.RaiseError("Method %s expected %d arguments\n", methodName, largs)
					return 0
				}

				args := make([]reflect.Value, largs)
				for i := 1; i <= largs; i++ {
					luaVal := L.Get(i)
					goVal := luaValToGo(luaVal)
					arg := reflect.ValueOf(goVal)
					args[i-1] = arg
				}

				results := method.Call(args)
				nresults := len(results)
				for i := 0; i < nresults; i++ {
					L.Push(goValToLua(L, results[i]))
				}
				return nresults
			}))
		}
	}
	addTypes(L, namespace, obj.Elem().Type().PkgPath())

	return nil
}

func addNamespace(L *lua.LState, name string) *lua.LTable {
	ns := L.NewTable()
	L.SetGlobal(name, ns)
	return ns
}
