```text
███████  ██████ ██████   ██████  ██████  ████████ 
██      ██      ██   ██ ██    ██ ██   ██    ██    
███████ ██      ██████  ██    ██ ██████     ██    
     ██ ██      ██   ██ ██    ██ ██         ██    
███████  ██████ ██   ██  ██████  ██         ██    
Script Operator for Kubernetes
```

## Overview
ScrOpt adds scripting to kubernetes by exposing the api to Lua scripts that can be created as resources. 

## How it works
ScrOpt uses https://github.com/yuin/gopher-lua in order to expose functions, types and values to the Lua scripting layer relying on a combination of reflection and static techniques in order to create the binding. See `internal/lua` for details.

## Binding
Following names are defined at the global scope
- log: format string compatible print function that prepends timestamps
- print: format string compatible print function
- ctx: context object 
- client:
- core:

 ## Example
 ```yaml
apiVersion: scripts.scropt.io/v1
kind: LuaScript
metadata:
  name: example
spec:
  code: |
    log("hello")
    local podList = core.PodList:new()
    for k, v in pairs(podList:get()) do
        print(k, v)
    end
    local listOptions = client.ListOptions:new()
    client.List(ctx, podList, listOptions)
    local result = podList:get()
    log("result is in")
    for _, v in ipairs(result.Items) do
      print(v)
      print("")
    end
    log("all done")

 ```