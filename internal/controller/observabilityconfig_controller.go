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

	grpcv1alpha1 "github.com/shtsukada/grpc-burner-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	// logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ObservabilityConfigReconciler reconciles a ObservabilityConfig object
type ObservabilityConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=grpc.burner.dev,resources=observabilityconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=grpc.burner.dev,resources=observabilityconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=grpc.burner.dev,resources=observabilityconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ObservabilityConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *ObservabilityConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var obsConfig grpcv1alpha1.ObservabilityConfig
	if err := r.Get(ctx, req.NamespacedName, &obsConfig); err != nil {
		logger.Error(err, "unable to fetch ObservabilityConfig")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling ObservabilityConfig",
		"name", obsConfig.Name,
		"logLevel", obsConfig.Spec.LogLevel,
		"metricsEnabled", obsConfig.Spec.MetricsEnabled,
	)

	obsConfig.Status.AppliedLogLevel = obsConfig.Spec.LogLevel
	obsConfig.Status.MetricsActive = &obsConfig.Spec.MetricsEnabled
	obsConfig.Status.Message = "settings changed"

	if err := r.Status().Update(ctx, &obsConfig); err != nil {
		logger.Error(err, "Status update failed")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObservabilityConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		// For().
		Named("observabilityconfig").
		Complete(r)
}
