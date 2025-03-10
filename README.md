```text
 _______ _______  ______  _____   _____  _______
 |______ |       |_____/ |     | |_____]    |   
 ______| |_____  |    \_ |_____| |          |   
```
 Script Operator for Kubernetes


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