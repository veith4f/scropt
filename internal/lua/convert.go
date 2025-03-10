package klua

import (
	"fmt"
	"math"
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

func luaValToGo(val lua.LValue) any {
	switch v := val.(type) {
	case *lua.LNilType:
		return nil
	case *lua.LTable:
		// either array
		if isArrayTable(v) {
			var result []any
			v.ForEach(func(_, value lua.LValue) {
				result = append(result, luaValToGo(value))
			})
			return result
		}
		// or map
		//typ := v.RawGetString("__type__")
		result := make(map[string]any)
		// if it is a map it may be a ptr
		maybePtr := v.RawGetString("__ptr__")
		if maybePtr != lua.LNil {
			return maybePtr.(*lua.LUserData).Value
		}
		// or a struct
		maybeStruct := v.RawGetString("__struct__")
		if maybeStruct != lua.LNil {
			return maybeStruct.(*lua.LUserData).Value
		}
		// otherwise it is a regular map
		v.ForEach(func(key lua.LValue, value lua.LValue) {
			result[fmt.Sprintf("%v", luaValToGo(key))] = luaValToGo(value)
		})
		return result
	case *lua.LUserData:
		return v.Value
	case lua.LString:
		return string(v)
	case lua.LNumber:
		f := float64(v)
		if f == math.Trunc(f) {
			return int(f)
		}
		return f
	case lua.LBool:
		return bool(v)
	default:
		//fmt.Printf("Don't know how to convert %s to go using string\n", v.Type().String())
		return v.String()
	}
}

func goValToLua(L *lua.LState, val reflect.Value) lua.LValue {

	// Handle invalid or nil values (e.g., uninitialized reflect.Value)
	if !val.IsValid() || val.Interface() == nil {
		return lua.LNil
	}

	switch val.Kind() {
	case reflect.Map:
		result := L.NewTable()
		for _, key := range val.MapKeys() {
			luaKey := goValToLua(L, key)
			luaValue := goValToLua(L, val.MapIndex(key))
			result.RawSet(luaKey, luaValue)
		}
		return result

	case reflect.Slice, reflect.Array:
		result := L.NewTable()
		for i := 0; i < val.Len(); i++ {
			result.RawSetInt(i+1, goValToLua(L, val.Index(i)))
		}
		return result

	case reflect.Ptr:
		result := L.NewTable()
		__ptr__ := L.NewUserData()

		if val.IsNil() {
			__ptr__.Value = nil
			L.SetField(result, "__ptr__", __ptr__)
			L.SetField(result, "__type__", lua.LNil)
			return result
		}

		__ptr__.Value = val.Interface()
		L.SetField(result, "__ptr__", __ptr__)

		__type__ := L.NewUserData()
		__type__.Value = val.Elem().Type() // Safe now
		L.SetField(result, "__type__", __type__)

		return result

	case reflect.Struct:
		result := L.NewTable()
		structType := val.Type()

		__struct__ := L.NewUserData()
		__struct__.Value = val.Interface()
		L.SetField(result, "__struct__", __struct__)

		__type__ := L.NewUserData()
		__type__.Value = structType
		L.SetField(result, "__type__", __type__)

		for i := 0; i < structType.NumField(); i++ {
			fieldVal := val.Field(i)
			if fieldVal.CanInterface() {
				result.RawSetString(structType.Field(i).Name, goValToLua(L, fieldVal))
			}
		}
		return result

	case reflect.String:
		return lua.LString(val.String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(val.Int())

	case reflect.Float32, reflect.Float64:
		return lua.LNumber(val.Float())

	case reflect.Bool:
		return lua.LBool(val.Bool())

	default:
		printStackTrace()
		L.RaiseError("Unsupported Go value type: %s", val.Kind())
		return lua.LNil
	}
}
