package lua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func CompileMoonscript(moonScript string) (string, error) {

	L := lua.NewState()
	defer L.Close()

	withLoader(L)

	L.SetGlobal("__moonscript_code", lua.LString(moonScript))

	if err := L.DoString(`
local moonparse = require("moonscript.parse")
local mooncompile = require("moonscript.compile")

local cleanUp = function()
	package.loaded.moonscript = nil
	moonparse = nil
	mooncompile = nil
	__moonscript_code = nil

	collectgarbage()
end

local tree, parseErr = moonparse.string(__moonscript_code)
if not tree then
	cleanUp()
	return nil, parseErr
end
	
local luaCode, compileErr = mooncompile.tree(tree)
if not luaCode or luaCode == "" then
	cleanUp()
	return nil, compileErr
end

cleanUp()
return luaCode, nil
`); err != nil {
		return "", err
	}

	luaCode := L.Get(-2)
	luaErr := L.Get(-1)
	L.Pop(2)

	if luaErr != lua.LNil {
		return "", fmt.Errorf("%s", luaErr)
	}

	return luaCode.String(), nil
}
