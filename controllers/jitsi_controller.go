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
	"github.com/presslabs/controller-util/pkg/syncer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "jitsi-operator/api/v1alpha1"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

// JitsiReconciler reconciles a Jitsi object
type JitsiReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.jit.si,resources=jitsis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.jit.si,resources=jitsis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.jit.si,resources=jitsis/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=*,verbs=*

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

	if jitsi.Spec.Suspend {
		return ctrl.Result{}, nil
	}

	if !jitsi.Spec.Jibri.Enabled {
		dep := jitsi.JibriDeployment()
		_ = r.Client.Delete(ctx, &dep)
	}

	if jitsi.Spec.JVB.Strategy.Type != appsv1alpha1.JVBStrategyAutoScaled {
		hpa := jitsi.JVBHPA()
		_ = r.Client.Delete(ctx, &hpa)
	}

	if jitsi.Spec.JVB.Strategy.Type == appsv1alpha1.JVBStrategyDaemon {
		dep := jitsi.JVBDeployment()
		_ = r.Client.Delete(ctx, &dep)
	} else {
		dep := jitsi.JVBDaemonSet()
		_ = r.Client.Delete(ctx, &dep)
	}

	syncers := []syncer.Interface{
		NewJitsiSecretSyncer(jitsi, r.Client),
		NewProsodyServiceSyncer(jitsi, r.Client),
		NewProsodyDeploymentSyncer(jitsi, r.Client),
		NewJicofoDeploymentSyncer(jitsi, r.Client),
		NewJVBConfigMapSyncer(jitsi, r.Client),
		NewWebDeploymentSyncer(jitsi, r.Client),
		NewWebServiceSyncer(jitsi, r.Client),
	}

	switch jitsi.Spec.JVB.Strategy.Type {
	case appsv1alpha1.JVBStrategyAutoScaled:
		syncers = append(syncers, NewJVBDeploymentSyncer(jitsi, r.Client))
		syncers = append(syncers, NewJVBHPASyncer(jitsi, r.Client))
	case appsv1alpha1.JVBStrategyDaemon:
		syncers = append(syncers, NewJVBDaemonSetSyncer(jitsi, r.Client))
	case appsv1alpha1.JVBStrategyStatic:
		syncers = append(syncers, NewJVBDeploymentSyncer(jitsi, r.Client))
	}

	if jitsi.Spec.Jibri.Enabled {
		syncers = append(syncers, NewJibriDeploymentSyncer(jitsi, r.Client))
	}

	if jitsi.Spec.Ingress.Enabled {
		syncers = append(syncers, NewIngressSyncer(jitsi, r.Client))
	}

	if jitsi.Spec.Metrics {
		syncers = append(syncers, NewJVBPodMonitorSyncer(jitsi, r.Client))
		syncers = append(syncers, NewJicofoServiceMonitorSyncer(jitsi, r.Client))
	}

	if err := r.sync(ctx, syncers); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *JitsiReconciler) sync(ctx context.Context, syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(ctx, s, r.recorder); err != nil {
			return err
		}
	}

	return nil
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
