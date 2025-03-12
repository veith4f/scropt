package lua

// This file uses cgo to interact with the Lua C API.
// There really isn't much documentation about cgo around
// but a useful guide can be found here: https://go.dev/wiki/cgo
//
// Other than that you are advised to study this file.

/*
#cgo CFLAGS: -I /Users/fschuetz/opt/lua/include
#cgo LDFLAGS: -L /Users/fschuetz/opt/lua/lib/ -L /Users/fschuetz/.luarocks/lib/lua/5.1/ -l lua -l lpeg
#include <stdlib.h>
#include "lua.h"
#include "lauxlib.h"
#include "lualib.h"

// consts
const char* const_require = "require";
const char* const_lpeg = "lpeg";

// Wrapper functions for macros
static int gl_istable(lua_State *L, int index) {
    return lua_istable(L, index);
}

static void gl_getglobal(lua_State *L, const char *name) {
	lua_getglobal(L, name);
}

static void gl_pop(lua_State *L, const int i) {
	lua_pop(L, i);
}

static const char* gl_tostring(lua_State *L, const int i) {
	return lua_tostring(L, i);
}

static const int gl_isnil(lua_State *L, const int i) {
	return lua_isnil(L, i);
}

*/
import "C"
import (
	"fmt"
	"math"
	"reflect"
	"unsafe"

	lua "github.com/yuin/gopher-lua"
)

const (
	LUA_TNONE          = -1
	LUA_TNIL           = 0
	LUA_TBOOLEAN       = 1
	LUA_TLIGHTUSERDATA = 2
	LUA_TNUMBER        = 3
	LUA_TSTRING        = 4
	LUA_TTABLE         = 5
	LUA_TFUNCTION      = 6
	LUA_TUSERDATA      = 7
	LUA_TTHREAD        = 8
)

func GetValue(L *C.lua_State, pos C.int) (any, error) {
	switch C.lua_type(L, pos) {
	case LUA_TNONE:
		return nil, fmt.Errorf("No value at stack pos %d", pos)
	case LUA_TNIL:
		return nil, nil
	case LUA_TBOOLEAN:
		return C.lua_toboolean(L, pos) != 0, nil
	case LUA_TLIGHTUSERDATA, LUA_TUSERDATA:
	case LUA_TNUMBER:
		value := float64(C.lua_tonumber(L, pos))
		if value == math.Trunc(value) {
			return int(value), nil
		}
		return value, nil
	case LUA_TSTRING:
		return C.GoString(C.gl_tostring(L, pos)), nil
	case LUA_TTABLE:
		table := make(map[string]any)
		C.lua_pushnil(L)
		for {
			if C.lua_next(L, pos-1) == 0 {
				break
			}
			key := C.GoString(C.gl_tostring(L, pos-1))
			val, err := GetValue(L, pos)
			if err != nil {
				return nil, err
			}
			table[key] = val
			C.gl_pop(L, 1) // remove entry from stack
		}
		C.gl_pop(L, 1) // pop table
		return table, nil
	case LUA_TFUNCTION:
		// execute the function in C Lua state and push results onto Gopher Lua state
		return func(GL *lua.LState, nresults int, args ...string) error {

			C.lua_pushlightuserdata(L, C.lua_topointer(L, pos))

			for _, a := range args {
				ca := C.CString(a)
				defer C.free(unsafe.Pointer(ca))
				C.lua_pushstring(L, ca)
			}

			if C.lua_pcall(L, C.int(len(args)), C.int(nresults), 0) != 0 {
				return fmt.Errorf("Error calling function")
			}

			for i := 0; i < nresults; i++ {
				GL.Push(lua.LString(C.GoString(C.gl_tostring(L, pos))))
			}

			return nil
		}, nil
	case LUA_TTHREAD:
		return nil, fmt.Errorf("Cannot handle thread at stack pos %d", pos)
	}
	return nil, fmt.Errorf("Uknown type encountered at stack pos %d", pos)
}

func GetGlobal(L *C.lua_State, name *C.char) (any, error) {
	C.gl_getglobal(L, name) // push value onto stack
	defer C.gl_pop(L, 1)

	if C.gl_isnil(L, -1) == 1 {
		return nil, fmt.Errorf("Not found: %s", C.GoString(name))
	}

	return GetValue(L, -1)
}

func Require(L *C.lua_State, module *C.char) {
	C.gl_getglobal(L, C.const_require)
	C.lua_pushstring(L, module)
	C.lua_pcall(L, 1, 1, 0)
}

// Load LPeg into Gopher-Lua
func LoadLPeg(GL *lua.LState) error {

	// Create a new C Lua state
	L := C.luaL_newstate()
	defer C.lua_close(L)
	C.luaL_openlibs(L)

	Require(L, C.const_lpeg)

	global, err := GetGlobal(L, C.const_lpeg)
	if err != nil {
		return err
	}

	table := GL.NewTable()
	for k, v := range global.(map[string]any) {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Func:
			table.RawSetString(k, GL.NewFunction(func(LGL *lua.LState) int {

				return 1
			}))
		default:
			str := lua.LString(v.(string))
			table.RawSetString(k, str)
		}
	}
	GL.Push(table)

	return nil
}
