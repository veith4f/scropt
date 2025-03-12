package lua

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"

	libs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeConfig() (*rest.Config, error) {
	// Check if running inside a cluster
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster {
		return rest.InClusterConfig() // In-cluster config
	}

	// Running outside cluster - use kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG") // Try KUBECONFIG env var
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = home + "/.kube/config" // Default kubeconfig path
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func Exec(ctx context.Context, code string, cli client.Client) error {

	config, err := getKubeConfig()
	if err != nil {
		panic(err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}

	L := lua.NewState()
	defer L.Close()

	libs.Preload(L)

	// add project assets
	if err := L.DoString(`package.path = package.path .. 
			";modules/lua/?.lua;modules/lua/?/init.lua"`); err != nil {
		return err
	}

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

	/*
		_scheme := addNamespace(L, "scheme")
		addObject(L, _scheme, scheme.Scheme)
	*/

	return L.DoString(code)
}
