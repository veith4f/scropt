/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	scrv1 "github.com/veith4f/scropt/api/v1"
	lua "github.com/veith4f/scropt/internal/lua"
)

// MoonScriptReconciler reconciles a MoonScript object
type MoonScriptReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=scripts.scropt.io,resources=moonscripts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scripts.scropt.io,resources=moonscripts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=scripts.scropt.io,resources=moonscripts/finalizers,verbs=update
// +kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MoonScript object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *MoonScriptReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Fetch Script resource
	script := &scrv1.MoonScript{}
	if err := r.Get(ctx, req.NamespacedName, script); err != nil {
		log.Printf("MoonScript resource not found, ignoring")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if script.Spec.Code == "" {
		log.Printf("Ignoring empty MoonScript: %s", fqn(script.ObjectMeta))
		return ctrl.Result{}, nil
	}

	if script.Status.Output == CONTROLLER_SCRIPT_EXECUTED {
		log.Printf("Ignoring already executed MoonScript: %s", fqn(script.ObjectMeta))
		return ctrl.Result{}, nil
	}

	// Create a copy to avoid modifying the script resource
	// it can be modified while we work on it
	// and a subsequent update may fail
	// Patch only the Status field to prevent conflicts
	patch := client.MergeFrom(script)
	scriptCopy := script.DeepCopy()

	// Compile MoonScript to Lua
	log.Printf("Compiling MoonScript: %s", fqn(script.ObjectMeta))
	luascript, err := lua.CompileMoonscript(script.Spec.Code)
	if err != nil {
		log.Printf("Error compiling MoonScript: %v", err)
		return ctrl.Result{}, err
	}

	// Execute compiled MoonScript
	log.Printf("Executing MoonScript: %s", fqn(script.ObjectMeta))
	if err := lua.Exec(ctx, luascript, r.Client); err != nil {
		scriptCopy.Status.Output = fmt.Sprintf("Error: %v", err)
	} else {
		scriptCopy.Status.Output = CONTROLLER_SCRIPT_EXECUTED
	}

	if err := r.Status().Patch(ctx, script, patch); err != nil {
		log.Printf("Failed updating MoonScript status to %v: %s", err, fqn(script.ObjectMeta))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MoonScriptReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scrv1.MoonScript{}).
		Named("moonscript").
		Complete(r)
}
