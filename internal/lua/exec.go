package lua

import (
	"context"
	"fmt"
	"log"
	"reflect"

	lua "github.com/yuin/gopher-lua"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Exec(ctx context.Context, code string, cli client.Client) error {

	L := lua.NewState()
	defer L.Close()
	//withLoader(L)

	addFunction(L, nil, "print", reflect.ValueOf(fmt.Println))
	addFunction(L, nil, "log", reflect.ValueOf(log.Printf))

	if err := addObject(L, "ctx", reflect.ValueOf(ctx)); err != nil {
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
