```text
███████  ██████ ██████   ██████  ██████  ████████ 
██      ██      ██   ██ ██    ██ ██   ██    ██    
███████ ██      ██████  ██    ██ ██████     ██    
     ██ ██      ██   ██ ██    ██ ██         ██    
███████  ██████ ██   ██  ██████  ██         ██    
Script Operator for Kubernetes
```

## Overview
`ScrOpt` adds scripting to kubernetes by exposing api methods available to the controller to `Lua` (https://www.lua.org/) or `Moonscript` (https://moonscript.org/). It defines Custom Resources `LuaScript` and `Moonscript` that can be created as resources and executes them in the respective controller's reconcile method. 

`ScrOpt` uses https://github.com/yuin/gopher-lua in order to expose functions, types and values to the `Lua` scripting layer relying on a combination of reflection and static techniques in order to create the binding. See `internal/lua/binding.go` for details. 

`Moonscript` is a convenience language that compiles to `Lua` and any Custom Resource `Moonscript` ultimately executes as `Lua` code in the very same environment.

## Binding
Following names are currently defined at the global scope. 
- log: format string compatible print function that prepends timestamps
- print: format string compatible print function
- ctx: methods and types from controller's context.Context object ("context")
- discovery: methods and types from discovery.DiscoveryClient ("k8s.io/client-go/discovery")
- client: methods and types from controllers's client.Client object ("sigs.k8s.io/controller-runtime/pkg/client") 
- core: types from core package ("k8s.io/api/core/v1")

Exposing `Go` constructs to `Lua` is mostly automated and requires very little effort.
```golang
addFunction(L, nil, "print", reflect.ValueOf(fmt.Println))
addFunction(L, nil, "log", reflect.ValueOf(log.Printf))

if err := addObject(L, "ctx", reflect.ValueOf(ctx)); err != nil {
  return err
}

if err := addObject(L, "discovery", reflect.ValueOf(discoveryClient)); err != nil {
  return err
}

if err := addObject(L, "client", reflect.ValueOf(cli)); err != nil {
  return err
}

coreNs := addNamespace(L, "core")
addTypes(L, coreNs, "k8s.io/api/core/v1")
```

 ## Examples
```yaml
apiVersion: scripts.scropt.io/v1
kind: MoonScript
metadata:
  name: example
spec:
  code: |
    class Thing
      name: "unknown"

    class Person extends Thing
      say_name: => print "Hello, I am #{@name}!"

    with Person!
      .name = "MoonScript"
      \say_name!
```
 
```yaml
apiVersion: scripts.scropt.io/v1
kind: LuaScript
metadata:
  name: example
spec:
  code: |
    log("Hello, I am Lua!")
    for _, p in pairs({ctx, client, core}) do
      if p["__NAME__"] ~= nil then
          print(p["__NAME__"])
      elseif p["__TYPE__"] ~= nil then
          print(p["__TYPE__"])
      else
        print("")
      end
      print("-------------------------------------------")
      for k, v in pairs(p) do 
        print(k, v)
      end
      print("")
    end
    --[[
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
    --]]
```

## Getting started
- `make install`: reate the CRDs in your cluster
- `kubectl apply -f examples`: Create example scripts in your cluster
- `make build`: build local
- `bin/scropt`: reconcile the managers, i.e. compile/execute any deployed scripts
- `make docker-build`: build docker containers
- `make docker-push`: push docker containers. Edit IMG in Makefile to set repository.

 ## List of exposed tables
 ```text
 context.valueCtx
-------------------------------------------
__PTR__ context.Background.WithCancel.WithValue(logr.contextKey, logr.Logger).WithValue(controller.reconcileIDKey, types.UID)
__TYPE__ context.valueCtx
Deadline function: 0x1400089a000
Done function: 0x1400089a080
Err function: 0x1400089a0c0
String function: 0x1400089a100
Value function: 0x1400089a140
AfterFunc function: 0x14001c4fa40
Background function: 0x14001c4fac0
Cause function: 0x14001c4fb40
TODO function: 0x14001c4fbc0
WithCancel function: 0x14001c4fc40
WithCancelCause function: 0x14001c4fd00
WithDeadline function: 0x14001c4fdc0
WithDeadlineCause function: 0x14001c4fe40
WithTimeout function: 0x14001c4ff00
WithTimeoutCause function: 0x14000ab2000
WithValue function: 0x14000ab2080
WithoutCancel function: 0x14000ab2100

discovery.DiscoveryClient
-------------------------------------------
__PTR__ &{0x140009b0320 /api false}
__TYPE__ discovery.DiscoveryClient
GroupsAndMaybeResources function: 0x140022aa1c0
OpenAPISchema function: 0x140022aa240
OpenAPIV3 function: 0x140022aa280
RESTClient function: 0x140022aa2c0
ServerGroups function: 0x140022aa300
ServerGroupsAndResources function: 0x140022aa3c0
ServerPreferredNamespacedResources function: 0x140022aa400
ServerPreferredResources function: 0x140022aa440
ServerResourcesForGroupVersion function: 0x140022aa480
ServerVersion function: 0x140022aa4c0
WithLegacy function: 0x140022aa500

client.client
-------------------------------------------
__PTR__ &{{0x14000755710 0x14000714380} {0x14000755710 {}} {0x14000886ef8 0x140006fce10} 0x14000377dc0 0x140006fce10 0x1400059f308 map[] false}
__TYPE__ client.client
Create function: 0x14000ab2280
Delete function: 0x14000ab2380
DeleteAllOf function: 0x14000ab2580
Get function: 0x14000ab2640
GroupVersionKindFor function: 0x14000ab2700
IsObjectNamespaced function: 0x14000ab2740
List function: 0x14000ab2800
Patch function: 0x14000ab2840
RESTMapper function: 0x14000ab2880
Scheme function: 0x14000ab28c0
Status function: 0x14000ab2900
SubResource function: 0x14000ab2940
Update function: 0x14000ab2a00
CacheOptions map[get:function: 0x140025b81c0 new:function: 0x140025b8240 set:function: 0x140025b8200]
CreateOptions map[get:function: 0x140025b82c0 new:function: 0x140025b8340 set:function: 0x140025b8300]
DeleteAllOfOptions map[get:function: 0x140025b83c0 new:function: 0x140025b8440 set:function: 0x140025b8400]
DeleteOptions map[get:function: 0x140025b84c0 new:function: 0x140025b8580 set:function: 0x140025b8500]
GetOptions map[get:function: 0x140025b8600 new:function: 0x140025b8680 set:function: 0x140025b8640]
IgnoreAlreadyExists function: 0x140025b8740
IgnoreNotFound function: 0x140025b8780
ListOptions map[get:function: 0x140025b87c0 new:function: 0x140025b8840 set:function: 0x140025b8800]
MatchingFieldsSelector map[get:function: 0x140025b88c0 new:function: 0x140025b89c0 set:function: 0x140025b8900]
MatchingLabelsSelector map[get:function: 0x140025b8a40 new:function: 0x140025b8ac0 set:function: 0x140025b8a80]
MergeFrom function: 0x140025b8b80
MergeFromOptions map[get:function: 0x140025b8bc0 new:function: 0x140025b8c40 set:function: 0x140025b8c00]
MergeFromWithOptimisticLock map[get:function: 0x140025b8cc0 new:function: 0x140025b8d80 set:function: 0x140025b8d00]
MergeFromWithOptions function: 0x140025b8e40
New function: 0x140025b8ec0
NewDryRunClient function: 0x140025b8f40
NewNamespacedClient function: 0x140025b8fc0
NewWithWatch function: 0x140025b9040
ObjectKeyFromObject function: 0x140025b90c0
Options map[get:function: 0x140025b9100 new:function: 0x140025b9180 set:function: 0x140025b9140]
PatchOptions map[get:function: 0x140025b9200 new:function: 0x140025b9280 set:function: 0x140025b9240]
Preconditions map[get:function: 0x140025b9300 new:function: 0x140025b9380 set:function: 0x140025b9340]
RawPatch function: 0x140025b9440
StrategicMergeFrom function: 0x140025b9480
SubResourceCreateOptions map[get:function: 0x140025b94c0 new:function: 0x140025b95c0 set:function: 0x140025b9580]
SubResourceGetOptions map[get:function: 0x140025b9640 new:function: 0x140025b96c0 set:function: 0x140025b9680]
SubResourcePatchOptions map[get:function: 0x140025b9740 new:function: 0x140025b97c0 set:function: 0x140025b9780]
SubResourceUpdateOptions map[get:function: 0x140025b9840 new:function: 0x140025b98c0 set:function: 0x140025b9880]
UpdateOptions map[get:function: 0x140025b9a00 new:function: 0x140025b9a80 set:function: 0x140025b9a40]
WithFieldOwner function: 0x140025b9b00
WithFieldValidation function: 0x140025b9b80
WithSubResourceBody function: 0x140025b9c40

core
-------------------------------------------
__NAME__ core
__TYPE__ namespace
AWSElasticBlockStoreVolumeSource map[get:function: 0x14000b6b200 new:function: 0x14000b6b280 set:function: 0x14000b6b240]
Affinity map[get:function: 0x14000b6b300 new:function: 0x14000b6b380 set:function: 0x14000b6b340]
AppArmorProfile map[get:function: 0x14000b6b400 new:function: 0x14000b6b480 set:function: 0x14000b6b440]
AttachedVolume map[get:function: 0x14000b6b500 new:function: 0x14000b6b580 set:function: 0x14000b6b540]
AvoidPods map[get:function: 0x14000b6b600 new:function: 0x14000b6b680 set:function: 0x14000b6b640]
AzureDiskVolumeSource map[get:function: 0x14000b6b700 new:function: 0x14000b6b780 set:function: 0x14000b6b740]
AzureFilePersistentVolumeSource map[get:function: 0x14000b6b800 new:function: 0x14000b6b880 set:function: 0x14000b6b840]
AzureFileVolumeSource map[get:function: 0x14000b6b900 new:function: 0x14000b6b980 set:function: 0x14000b6b940]
Binding map[get:function: 0x14000b6ba00 new:function: 0x14000b6ba80 set:function: 0x14000b6ba40]
CSIPersistentVolumeSource map[get:function: 0x14000b6bb00 new:function: 0x14000b6bb80 set:function: 0x14000b6bb40]
CSIVolumeSource map[get:function: 0x14000b6bc00 new:function: 0x14000b6bc80 set:function: 0x14000b6bc40]
Capabilities map[get:function: 0x14000b6bd00 new:function: 0x14000b6bd80 set:function: 0x14000b6bd40]
CephFSPersistentVolumeSource map[get:function: 0x14000b6be00 new:function: 0x14000b6be80 set:function: 0x14000b6be40]
CephFSVolumeSource map[get:function: 0x14000b6bf00 new:function: 0x14000b6e000 set:function: 0x14000b6bf40]
CinderPersistentVolumeSource map[get:function: 0x14000b6e080 new:function: 0x14000b6e100 set:function: 0x14000b6e0c0]
CinderVolumeSource map[get:function: 0x14000b6e180 new:function: 0x14000b6e200 set:function: 0x14000b6e1c0]
ClientIPConfig map[get:function: 0x14000b6e280 new:function: 0x14000b6e300 set:function: 0x14000b6e2c0]
ClusterTrustBundleProjection map[get:function: 0x14000b6e380 new:function: 0x14000b6e400 set:function: 0x14000b6e3c0]
ComponentCondition map[get:function: 0x14000b6e480 new:function: 0x14000b6e500 set:function: 0x14000b6e4c0]
ComponentStatus map[get:function: 0x14000b6e580 new:function: 0x14000b6e600 set:function: 0x14000b6e5c0]
ComponentStatusList map[get:function: 0x14000b6e680 new:function: 0x14000b6e700 set:function: 0x14000b6e6c0]
ConfigMap map[get:function: 0x14000b6e780 new:function: 0x14000b6e800 set:function: 0x14000b6e7c0]
ConfigMapEnvSource map[get:function: 0x14000b6e880 new:function: 0x14000b6e900 set:function: 0x14000b6e8c0]
ConfigMapKeySelector map[get:function: 0x14000b6e980 new:function: 0x14000b6ea00 set:function: 0x14000b6e9c0]
ConfigMapList map[get:function: 0x14000b6ea80 new:function: 0x14000b6eb00 set:function: 0x14000b6eac0]
ConfigMapNodeConfigSource map[get:function: 0x14000b6eb80 new:function: 0x14000b6ec00 set:function: 0x14000b6ebc0]
ConfigMapProjection map[get:function: 0x14000b6ec80 new:function: 0x14000b6ed00 set:function: 0x14000b6ecc0]
ConfigMapVolumeSource map[get:function: 0x14000b6ed80 new:function: 0x14000b6ee00 set:function: 0x14000b6edc0]
Container map[get:function: 0x14000b6ee80 new:function: 0x14000b6ef00 set:function: 0x14000b6eec0]
ContainerImage map[get:function: 0x14000b6ef80 new:function: 0x14000b6f000 set:function: 0x14000b6efc0]
ContainerPort map[get:function: 0x14000b6f080 new:function: 0x14000b6f100 set:function: 0x14000b6f0c0]
ContainerResizePolicy map[get:function: 0x14000b6f180 new:function: 0x14000b6f200 set:function: 0x14000b6f1c0]
ContainerState map[get:function: 0x14000b6f280 new:function: 0x14000b6f300 set:function: 0x14000b6f2c0]
ContainerStateRunning map[get:function: 0x14000b6f380 new:function: 0x14000b6f400 set:function: 0x14000b6f3c0]
ContainerStateTerminated map[get:function: 0x14000b6f480 new:function: 0x14000b6f500 set:function: 0x14000b6f4c0]
ContainerStateWaiting map[get:function: 0x14000b6f580 new:function: 0x14000b6f600 set:function: 0x14000b6f5c0]
ContainerStatus map[get:function: 0x14000b6f680 new:function: 0x14000b6f700 set:function: 0x14000b6f6c0]
ContainerUser map[get:function: 0x14000b6f780 new:function: 0x14000b6f800 set:function: 0x14000b6f7c0]
DaemonEndpoint map[get:function: 0x14000b6f880 new:function: 0x14000b6f900 set:function: 0x14000b6f8c0]
DownwardAPIProjection map[get:function: 0x14000b6f980 new:function: 0x14000b6fa00 set:function: 0x14000b6f9c0]
DownwardAPIVolumeFile map[get:function: 0x14000b6fa80 new:function: 0x14000b6fb00 set:function: 0x14000b6fac0]
DownwardAPIVolumeSource map[get:function: 0x14000b6fb80 new:function: 0x14000b6fc00 set:function: 0x14000b6fbc0]
EmptyDirVolumeSource map[get:function: 0x14000b6fc80 new:function: 0x14000b6fd00 set:function: 0x14000b6fcc0]
EndpointAddress map[get:function: 0x14000b6fd80 new:function: 0x14000b6fe00 set:function: 0x14000b6fdc0]
EndpointPort map[get:function: 0x14000b6fe80 new:function: 0x14000b6ff00 set:function: 0x14000b6fec0]
EndpointSubset map[get:function: 0x14000b72000 new:function: 0x14000b72080 set:function: 0x14000b72040]
Endpoints map[get:function: 0x14000b72100 new:function: 0x14000b72180 set:function: 0x14000b72140]
EndpointsList map[get:function: 0x14000b72200 new:function: 0x14000b72280 set:function: 0x14000b72240]
EnvFromSource map[get:function: 0x14000b72300 new:function: 0x14000b72380 set:function: 0x14000b72340]
EnvVar map[get:function: 0x14000b72400 new:function: 0x14000b72480 set:function: 0x14000b72440]
EnvVarSource map[get:function: 0x14000b72500 new:function: 0x14000b72580 set:function: 0x14000b72540]
EphemeralContainer map[get:function: 0x14000b72600 new:function: 0x14000b72680 set:function: 0x14000b72640]
EphemeralContainerCommon map[get:function: 0x14000b72700 new:function: 0x14000b72780 set:function: 0x14000b72740]
EphemeralVolumeSource map[get:function: 0x14000b72800 new:function: 0x14000b72880 set:function: 0x14000b72840]
Event map[get:function: 0x14000b72900 new:function: 0x14000b72980 set:function: 0x14000b72940]
EventList map[get:function: 0x14000b72a00 new:function: 0x14000b72a80 set:function: 0x14000b72a40]
EventSeries map[get:function: 0x14000b72b00 new:function: 0x14000b72b80 set:function: 0x14000b72b40]
EventSource map[get:function: 0x14000b72c00 new:function: 0x14000b72c80 set:function: 0x14000b72c40]
ExecAction map[get:function: 0x14000b72d00 new:function: 0x14000b72d80 set:function: 0x14000b72d40]
FCVolumeSource map[get:function: 0x14000b72e00 new:function: 0x14000b72e80 set:function: 0x14000b72e40]
FlexPersistentVolumeSource map[get:function: 0x14000b72f00 new:function: 0x14000b72f80 set:function: 0x14000b72f40]
FlexVolumeSource map[get:function: 0x14000b73000 new:function: 0x14000b73080 set:function: 0x14000b73040]
FlockerVolumeSource map[get:function: 0x14000b73100 new:function: 0x14000b73180 set:function: 0x14000b73140]
GCEPersistentDiskVolumeSource map[get:function: 0x14000b73200 new:function: 0x14000b73280 set:function: 0x14000b73240]
GRPCAction map[get:function: 0x14000b73300 new:function: 0x14000b73380 set:function: 0x14000b73340]
GitRepoVolumeSource map[get:function: 0x14000b73400 new:function: 0x14000b73480 set:function: 0x14000b73440]
GlusterfsPersistentVolumeSource map[get:function: 0x14000b73500 new:function: 0x14000b73580 set:function: 0x14000b73540]
GlusterfsVolumeSource map[get:function: 0x14000b73600 new:function: 0x14000b73680 set:function: 0x14000b73640]
HTTPGetAction map[get:function: 0x14000b73700 new:function: 0x14000b73780 set:function: 0x14000b73740]
HTTPHeader map[get:function: 0x14000b73800 new:function: 0x14000b73880 set:function: 0x14000b73840]
HostAlias map[get:function: 0x14000b73900 new:function: 0x14000b73980 set:function: 0x14000b73940]
HostIP map[get:function: 0x14000b73a00 new:function: 0x14000b73a80 set:function: 0x14000b73a40]
HostPathVolumeSource map[get:function: 0x14000b73b00 new:function: 0x14000b73b80 set:function: 0x14000b73b40]
ISCSIPersistentVolumeSource map[get:function: 0x14000b73c00 new:function: 0x14000b73c80 set:function: 0x14000b73c40]
ISCSIVolumeSource map[get:function: 0x14000b73d00 new:function: 0x14000b73d80 set:function: 0x14000b73d40]
ImageVolumeSource map[get:function: 0x14000b73e00 new:function: 0x14000b73e80 set:function: 0x14000b73e40]
KeyToPath map[get:function: 0x14000b73f00 new:function: 0x14000b80000 set:function: 0x14000b73f40]
Lifecycle map[get:function: 0x14000b80080 new:function: 0x14000b80100 set:function: 0x14000b800c0]
LifecycleHandler map[get:function: 0x14000b80180 new:function: 0x14000b80200 set:function: 0x14000b801c0]
LimitRange map[get:function: 0x14000b80280 new:function: 0x14000b80300 set:function: 0x14000b802c0]
LimitRangeItem map[get:function: 0x14000b80380 new:function: 0x14000b80400 set:function: 0x14000b803c0]
LimitRangeList map[get:function: 0x14000b80480 new:function: 0x14000b80500 set:function: 0x14000b804c0]
LimitRangeSpec map[get:function: 0x14000b80580 new:function: 0x14000b80600 set:function: 0x14000b805c0]
LinuxContainerUser map[get:function: 0x14000b80680 new:function: 0x14000b80700 set:function: 0x14000b806c0]
List map[get:function: 0x14000b80780 new:function: 0x14000b80800 set:function: 0x14000b807c0]
LoadBalancerIngress map[get:function: 0x14000b80880 new:function: 0x14000b80900 set:function: 0x14000b808c0]
LoadBalancerStatus map[get:function: 0x14000b80980 new:function: 0x14000b80a00 set:function: 0x14000b809c0]
LocalObjectReference map[get:function: 0x14000b80a80 new:function: 0x14000b80b00 set:function: 0x14000b80ac0]
LocalVolumeSource map[get:function: 0x14000b80b80 new:function: 0x14000b80c00 set:function: 0x14000b80bc0]
ModifyVolumeStatus map[get:function: 0x14000b80c80 new:function: 0x14000b80d00 set:function: 0x14000b80cc0]
NFSVolumeSource map[get:function: 0x14000b80d80 new:function: 0x14000b80e00 set:function: 0x14000b80dc0]
Namespace map[get:function: 0x14000b80e80 new:function: 0x14000b80f00 set:function: 0x14000b80ec0]
NamespaceCondition map[get:function: 0x14000b80f80 new:function: 0x14000b81000 set:function: 0x14000b80fc0]
NamespaceList map[get:function: 0x14000b81080 new:function: 0x14000b81100 set:function: 0x14000b810c0]
NamespaceSpec map[get:function: 0x14000b81180 new:function: 0x14000b81200 set:function: 0x14000b811c0]
NamespaceStatus map[get:function: 0x14000b81280 new:function: 0x14000b81300 set:function: 0x14000b812c0]
Node map[get:function: 0x14000b81380 new:function: 0x14000b81400 set:function: 0x14000b813c0]
NodeAddress map[get:function: 0x14000b81480 new:function: 0x14000b81500 set:function: 0x14000b814c0]
NodeAffinity map[get:function: 0x14000b81580 new:function: 0x14000b81600 set:function: 0x14000b815c0]
NodeCondition map[get:function: 0x14000b81680 new:function: 0x14000b81700 set:function: 0x14000b816c0]
NodeConfigSource map[get:function: 0x14000b81780 new:function: 0x14000b81800 set:function: 0x14000b817c0]
NodeConfigStatus map[get:function: 0x14000b81880 new:function: 0x14000b81900 set:function: 0x14000b818c0]
NodeDaemonEndpoints map[get:function: 0x14000b81980 new:function: 0x14000b81a00 set:function: 0x14000b819c0]
NodeFeatures map[get:function: 0x14000b81a80 new:function: 0x14000b81b00 set:function: 0x14000b81ac0]
NodeList map[get:function: 0x14000b81b80 new:function: 0x14000b81c00 set:function: 0x14000b81bc0]
NodeProxyOptions map[get:function: 0x14000b81c80 new:function: 0x14000b81d00 set:function: 0x14000b81cc0]
NodeRuntimeHandler map[get:function: 0x14000b81d80 new:function: 0x14000b81e00 set:function: 0x14000b81dc0]
NodeRuntimeHandlerFeatures map[get:function: 0x14000b81e80 new:function: 0x14000b81f00 set:function: 0x14000b81ec0]
NodeSelector map[get:function: 0x14000b96000 new:function: 0x14000b96080 set:function: 0x14000b96040]
NodeSelectorRequirement map[get:function: 0x14000b96100 new:function: 0x14000b96180 set:function: 0x14000b96140]
NodeSelectorTerm map[get:function: 0x14000b96200 new:function: 0x14000b96280 set:function: 0x14000b96240]
NodeSpec map[get:function: 0x14000b96300 new:function: 0x14000b96380 set:function: 0x14000b96340]
NodeStatus map[get:function: 0x14000b96400 new:function: 0x14000b96480 set:function: 0x14000b96440]
NodeSystemInfo map[get:function: 0x14000b96500 new:function: 0x14000b96580 set:function: 0x14000b96540]
ObjectFieldSelector map[get:function: 0x14000b96600 new:function: 0x14000b96680 set:function: 0x14000b96640]
ObjectReference map[get:function: 0x14000b96700 new:function: 0x14000b96780 set:function: 0x14000b96740]
PersistentVolume map[get:function: 0x14000b96800 new:function: 0x14000b96880 set:function: 0x14000b96840]
PersistentVolumeClaim map[get:function: 0x14000b96900 new:function: 0x14000b96980 set:function: 0x14000b96940]
PersistentVolumeClaimCondition map[get:function: 0x14000b96a00 new:function: 0x14000b96a80 set:function: 0x14000b96a40]
PersistentVolumeClaimList map[get:function: 0x14000b96b00 new:function: 0x14000b96b80 set:function: 0x14000b96b40]
PersistentVolumeClaimSpec map[get:function: 0x14000b96c00 new:function: 0x14000b96c80 set:function: 0x14000b96c40]
PersistentVolumeClaimStatus map[get:function: 0x14000b96d00 new:function: 0x14000b96d80 set:function: 0x14000b96d40]
PersistentVolumeClaimTemplate map[get:function: 0x14000b96e00 new:function: 0x14000b96e80 set:function: 0x14000b96e40]
PersistentVolumeClaimVolumeSource map[get:function: 0x14000b96f00 new:function: 0x14000b96f80 set:function: 0x14000b96f40]
PersistentVolumeList map[get:function: 0x14000b97000 new:function: 0x14000b97080 set:function: 0x14000b97040]
PersistentVolumeSource map[get:function: 0x14000b97100 new:function: 0x14000b97180 set:function: 0x14000b97140]
PersistentVolumeSpec map[get:function: 0x14000b97200 new:function: 0x14000b97280 set:function: 0x14000b97240]
PersistentVolumeStatus map[get:function: 0x14000b97300 new:function: 0x14000b97380 set:function: 0x14000b97340]
PhotonPersistentDiskVolumeSource map[get:function: 0x14000b97400 new:function: 0x14000b97480 set:function: 0x14000b97440]
Pod map[get:function: 0x14000b97500 new:function: 0x14000b97580 set:function: 0x14000b97540]
PodAffinity map[get:function: 0x14000b97600 new:function: 0x14000b97680 set:function: 0x14000b97640]
PodAffinityTerm map[get:function: 0x14000b97700 new:function: 0x14000b97780 set:function: 0x14000b97740]
PodAntiAffinity map[get:function: 0x14000b97800 new:function: 0x14000b97880 set:function: 0x14000b97840]
PodAttachOptions map[get:function: 0x14000b97900 new:function: 0x14000b97980 set:function: 0x14000b97940]
PodCondition map[get:function: 0x14000b97a00 new:function: 0x14000b97a80 set:function: 0x14000b97a40]
PodDNSConfig map[get:function: 0x14000b97b00 new:function: 0x14000b97b80 set:function: 0x14000b97b40]
PodDNSConfigOption map[get:function: 0x14000b97c00 new:function: 0x14000b97c80 set:function: 0x14000b97c40]
PodExecOptions map[get:function: 0x14000b97d00 new:function: 0x14000b97d80 set:function: 0x14000b97d40]
PodIP map[get:function: 0x14000b97e00 new:function: 0x14000b97e80 set:function: 0x14000b97e40]
PodList map[get:function: 0x14000b97f00 new:function: 0x14000baa000 set:function: 0x14000b97f40]
PodLogOptions map[get:function: 0x14000baa080 new:function: 0x14000baa100 set:function: 0x14000baa0c0]
PodOS map[get:function: 0x14000baa180 new:function: 0x14000baa200 set:function: 0x14000baa1c0]
PodPortForwardOptions map[get:function: 0x14000baa280 new:function: 0x14000baa300 set:function: 0x14000baa2c0]
PodProxyOptions map[get:function: 0x14000baa380 new:function: 0x14000baa400 set:function: 0x14000baa3c0]
PodReadinessGate map[get:function: 0x14000baa480 new:function: 0x14000baa500 set:function: 0x14000baa4c0]
PodResourceClaim map[get:function: 0x14000baa580 new:function: 0x14000baa600 set:function: 0x14000baa5c0]
PodResourceClaimStatus map[get:function: 0x14000baa680 new:function: 0x14000baa700 set:function: 0x14000baa6c0]
PodSchedulingGate map[get:function: 0x14000baa780 new:function: 0x14000baa800 set:function: 0x14000baa7c0]
PodSecurityContext map[get:function: 0x14000baa880 new:function: 0x14000baa900 set:function: 0x14000baa8c0]
PodSignature map[get:function: 0x14000baa980 new:function: 0x14000baaa00 set:function: 0x14000baa9c0]
PodSpec map[get:function: 0x14000baaa80 new:function: 0x14000baab00 set:function: 0x14000baaac0]
PodStatus map[get:function: 0x14000baab80 new:function: 0x14000baac00 set:function: 0x14000baabc0]
PodStatusResult map[get:function: 0x14000baac80 new:function: 0x14000baad00 set:function: 0x14000baacc0]
PodTemplate map[get:function: 0x14000baad80 new:function: 0x14000baae00 set:function: 0x14000baadc0]
PodTemplateList map[get:function: 0x14000baae80 new:function: 0x14000baaf00 set:function: 0x14000baaec0]
PodTemplateSpec map[get:function: 0x14000baaf80 new:function: 0x14000bab000 set:function: 0x14000baafc0]
PortStatus map[get:function: 0x14000bab080 new:function: 0x14000bab100 set:function: 0x14000bab0c0]
PortworxVolumeSource map[get:function: 0x14000bab180 new:function: 0x14000bab200 set:function: 0x14000bab1c0]
Preconditions map[get:function: 0x14000bab280 new:function: 0x14000bab300 set:function: 0x14000bab2c0]
PreferAvoidPodsEntry map[get:function: 0x14000bab380 new:function: 0x14000bab400 set:function: 0x14000bab3c0]
PreferredSchedulingTerm map[get:function: 0x14000bab480 new:function: 0x14000bab500 set:function: 0x14000bab4c0]
Probe map[get:function: 0x14000bab580 new:function: 0x14000bab600 set:function: 0x14000bab5c0]
ProbeHandler map[get:function: 0x14000bab680 new:function: 0x14000bab700 set:function: 0x14000bab6c0]
ProjectedVolumeSource map[get:function: 0x14000bab780 new:function: 0x14000bab800 set:function: 0x14000bab7c0]
QuobyteVolumeSource map[get:function: 0x14000bab880 new:function: 0x14000bab900 set:function: 0x14000bab8c0]
RBDPersistentVolumeSource map[get:function: 0x14000bab980 new:function: 0x14000baba00 set:function: 0x14000bab9c0]
RBDVolumeSource map[get:function: 0x14000baba80 new:function: 0x14000babb00 set:function: 0x14000babac0]
RangeAllocation map[get:function: 0x14000babb80 new:function: 0x14000babc00 set:function: 0x14000babbc0]
ReplicationController map[get:function: 0x14000babc80 new:function: 0x14000babd00 set:function: 0x14000babcc0]
ReplicationControllerCondition map[get:function: 0x14000babd80 new:function: 0x14000babe00 set:function: 0x14000babdc0]
ReplicationControllerList map[get:function: 0x14000babe80 new:function: 0x14000babf00 set:function: 0x14000babec0]
ReplicationControllerSpec map[get:function: 0x14000bc2000 new:function: 0x14000bc2080 set:function: 0x14000bc2040]
ReplicationControllerStatus map[get:function: 0x14000bc2100 new:function: 0x14000bc2180 set:function: 0x14000bc2140]
ResourceClaim map[get:function: 0x14000bc2200 new:function: 0x14000bc2280 set:function: 0x14000bc2240]
ResourceFieldSelector map[get:function: 0x14000bc2300 new:function: 0x14000bc2380 set:function: 0x14000bc2340]
ResourceHealth map[get:function: 0x14000bc2400 new:function: 0x14000bc2480 set:function: 0x14000bc2440]
ResourceQuota map[get:function: 0x14000bc2500 new:function: 0x14000bc2580 set:function: 0x14000bc2540]
ResourceQuotaList map[get:function: 0x14000bc2600 new:function: 0x14000bc2680 set:function: 0x14000bc2640]
ResourceQuotaSpec map[get:function: 0x14000bc2700 new:function: 0x14000bc2780 set:function: 0x14000bc2740]
ResourceQuotaStatus map[get:function: 0x14000bc2800 new:function: 0x14000bc2880 set:function: 0x14000bc2840]
ResourceRequirements map[get:function: 0x14000bc2900 new:function: 0x14000bc2980 set:function: 0x14000bc2940]
ResourceStatus map[get:function: 0x14000bc2a00 new:function: 0x14000bc2a80 set:function: 0x14000bc2a40]
SELinuxOptions map[get:function: 0x14000bc2b00 new:function: 0x14000bc2b80 set:function: 0x14000bc2b40]
ScaleIOPersistentVolumeSource map[get:function: 0x14000bc2c00 new:function: 0x14000bc2c80 set:function: 0x14000bc2c40]
ScaleIOVolumeSource map[get:function: 0x14000bc2d00 new:function: 0x14000bc2d80 set:function: 0x14000bc2d40]
ScopeSelector map[get:function: 0x14000bc2e00 new:function: 0x14000bc2e80 set:function: 0x14000bc2e40]
ScopedResourceSelectorRequirement map[get:function: 0x14000bc2f00 new:function: 0x14000bc2f80 set:function: 0x14000bc2f40]
SeccompProfile map[get:function: 0x14000bc3000 new:function: 0x14000bc3080 set:function: 0x14000bc3040]
Secret map[get:function: 0x14000bc3100 new:function: 0x14000bc3180 set:function: 0x14000bc3140]
SecretEnvSource map[get:function: 0x14000bc3200 new:function: 0x14000bc3280 set:function: 0x14000bc3240]
SecretKeySelector map[get:function: 0x14000bc3300 new:function: 0x14000bc3380 set:function: 0x14000bc3340]
SecretList map[get:function: 0x14000bc3400 new:function: 0x14000bc3480 set:function: 0x14000bc3440]
SecretProjection map[get:function: 0x14000bc3500 new:function: 0x14000bc3580 set:function: 0x14000bc3540]
SecretReference map[get:function: 0x14000bc3600 new:function: 0x14000bc3680 set:function: 0x14000bc3640]
SecretVolumeSource map[get:function: 0x14000bc3700 new:function: 0x14000bc3780 set:function: 0x14000bc3740]
SecurityContext map[get:function: 0x14000bc3800 new:function: 0x14000bc3880 set:function: 0x14000bc3840]
SerializedReference map[get:function: 0x14000bc3900 new:function: 0x14000bc3980 set:function: 0x14000bc3940]
Service map[get:function: 0x14000bc3a00 new:function: 0x14000bc3a80 set:function: 0x14000bc3a40]
ServiceAccount map[get:function: 0x14000bc3b00 new:function: 0x14000bc3b80 set:function: 0x14000bc3b40]
ServiceAccountList map[get:function: 0x14000bc3c00 new:function: 0x14000bc3c80 set:function: 0x14000bc3c40]
ServiceAccountTokenProjection map[get:function: 0x14000bc3d00 new:function: 0x14000bc3d80 set:function: 0x14000bc3d40]
ServiceList map[get:function: 0x14000bc3e00 new:function: 0x14000bc3e80 set:function: 0x14000bc3e40]
ServicePort map[get:function: 0x14000bc3f00 new:function: 0x14000bdc000 set:function: 0x14000bc3f40]
ServiceProxyOptions map[get:function: 0x14000bdc080 new:function: 0x14000bdc100 set:function: 0x14000bdc0c0]
ServiceSpec map[get:function: 0x14000bdc180 new:function: 0x14000bdc200 set:function: 0x14000bdc1c0]
ServiceStatus map[get:function: 0x14000bdc280 new:function: 0x14000bdc300 set:function: 0x14000bdc2c0]
SessionAffinityConfig map[get:function: 0x14000bdc380 new:function: 0x14000bdc400 set:function: 0x14000bdc3c0]
SleepAction map[get:function: 0x14000bdc480 new:function: 0x14000bdc500 set:function: 0x14000bdc4c0]
StorageOSPersistentVolumeSource map[get:function: 0x14000bdc580 new:function: 0x14000bdc600 set:function: 0x14000bdc5c0]
StorageOSVolumeSource map[get:function: 0x14000bdc680 new:function: 0x14000bdc700 set:function: 0x14000bdc6c0]
Sysctl map[get:function: 0x14000bdc780 new:function: 0x14000bdc800 set:function: 0x14000bdc7c0]
TCPSocketAction map[get:function: 0x14000bdc880 new:function: 0x14000bdc900 set:function: 0x14000bdc8c0]
Taint map[get:function: 0x14000bdc980 new:function: 0x14000bdca00 set:function: 0x14000bdc9c0]
Toleration map[get:function: 0x14000bdca80 new:function: 0x14000bdcb00 set:function: 0x14000bdcac0]
TopologySelectorLabelRequirement map[get:function: 0x14000bdcb80 new:function: 0x14000bdcc00 set:function: 0x14000bdcbc0]
TopologySelectorTerm map[get:function: 0x14000bdcc80 new:function: 0x14000bdcd00 set:function: 0x14000bdccc0]
TopologySpreadConstraint map[get:function: 0x14000bdcd80 new:function: 0x14000bdce00 set:function: 0x14000bdcdc0]
TypedLocalObjectReference map[get:function: 0x14000bdce80 new:function: 0x14000bdcf00 set:function: 0x14000bdcec0]
TypedObjectReference map[get:function: 0x14000bdcf80 new:function: 0x14000bdd000 set:function: 0x14000bdcfc0]
Volume map[get:function: 0x14000bdd080 new:function: 0x14000bdd100 set:function: 0x14000bdd0c0]
VolumeDevice map[get:function: 0x14000bdd180 new:function: 0x14000bdd200 set:function: 0x14000bdd1c0]
VolumeMount map[get:function: 0x14000bdd280 new:function: 0x14000bdd300 set:function: 0x14000bdd2c0]
VolumeMountStatus map[get:function: 0x14000bdd380 new:function: 0x14000bdd400 set:function: 0x14000bdd3c0]
VolumeNodeAffinity map[get:function: 0x14000bdd480 new:function: 0x14000bdd500 set:function: 0x14000bdd4c0]
VolumeProjection map[get:function: 0x14000bdd580 new:function: 0x14000bdd600 set:function: 0x14000bdd5c0]
VolumeResourceRequirements map[get:function: 0x14000bdd680 new:function: 0x14000bdd700 set:function: 0x14000bdd6c0]
VolumeSource map[get:function: 0x14000bdd780 new:function: 0x14000bdd800 set:function: 0x14000bdd7c0]
VsphereVirtualDiskVolumeSource map[get:function: 0x14000bdd880 new:function: 0x14000bdd900 set:function: 0x14000bdd8c0]
WeightedPodAffinityTerm map[get:function: 0x14000bdd980 new:function: 0x14000bdda00 set:function: 0x14000bdd9c0]
WindowsSecurityContextOptions map[get:function: 0x14000bdda80 new:function: 0x14000bddb00 set:function: 0x14000bddac0]
```