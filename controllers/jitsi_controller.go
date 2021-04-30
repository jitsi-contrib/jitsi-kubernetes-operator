/*
Copyright 2021.

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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "jitsi-operator/api/v1alpha1"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

// JitsiReconciler reconciles a Jitsi object
type JitsiReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.jit.si,resources=jitsis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.jit.si,resources=jitsis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.jit.si,resources=jitsis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Jitsi object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *JitsiReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("jitsi", req.NamespacedName)

	jitsi := &appsv1alpha1.Jitsi{}
	if err := r.Client.Get(ctx, req.NamespacedName, jitsi); err != nil {
		//log.Error("unable to fetch Application", "error", err)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	jitsi.SetDefaults()

	jitsiSecretSyncer := NewJitsiSecretSyncer(jitsi, r.Client)
	if _, err := jitsiSecretSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	prosodyServiceSyncer := NewProsodyServiceSyncer(jitsi, r.Client)
	if _, err := prosodyServiceSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	prosodyDepSyncer := NewProsodyDeploymentSyncer(jitsi, r.Client)
	if _, err := prosodyDepSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	jicofoSyncer := NewJicofoDeploymentSyncer(jitsi, r.Client)
	if _, err := jicofoSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	jvbServiceSyncer := NewJVBServiceSyncer(jitsi, r.Client)
	if _, err := jvbServiceSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	jvbCMSyncer := NewJVBConfigMapSyncer(jitsi, r.Client)
	if _, err := jvbCMSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	jvbSyncer := NewJVBDeploymentSyncer(jitsi, r.Client)
	if _, err := jvbSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	webSyncer := NewWebDeploymentSyncer(jitsi, r.Client)
	if _, err := webSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	webServiceSyncer := NewWebServiceSyncer(jitsi, r.Client)
	if _, err := webServiceSyncer.Sync(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *JitsiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Jitsi{}).
		Complete(r)
}
