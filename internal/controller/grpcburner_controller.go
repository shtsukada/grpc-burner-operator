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
	"reflect"

	grpcv1alpha1 "github.com/shtsukada/grpc-burner-operator/api/v1alpha1"
	"github.com/shtsukada/grpc-burner-operator/internal/metrics"
	"go.opentelemetry.io/otel"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const grpcburnerFinalizer = "grpcburner.grpc.burner.dev/finalizer"

type GrpcBurnerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=grpc.burner.dev,resources=grpcburners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=grpc.burner.dev,resources=grpcburners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=grpc.burner.dev,resources=grpcburners/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GrpcBurner object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *GrpcBurnerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues(
		"controller", "grpcburner",
		"name", req.Name,
		"namespace", req.Namespace,
	)

	log.Info("Starting reconciliation")

	metrics.ReconcileTotal.WithLabelValues("grpcburner").Inc()

	tracer := otel.Tracer("controller.grpcburner")
	ctx, span := tracer.Start(ctx, "Reconcile")
	defer span.End()

	var grpcburner grpcv1alpha1.GrpcBurner
	if err := r.Get(ctx, req.NamespacedName, &grpcburner); err != nil {
		log.Error(err, "Failed to fetch GrpcBurner resource")
		metrics.ReconcileErrors.WithLabelValues("grpcburner").Inc()
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log = log.WithValues("phase", grpcburner.Status.Phase)

	deployName := grpcburner.Name + "-burner"
	var deploy appsv1.Deployment
	_ = r.Get(ctx, types.NamespacedName{Name: deployName, Namespace: grpcburner.Namespace}, &deploy)

	if grpcburner.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&grpcburner, grpcburnerFinalizer) {
			log.Info("Adding finalizer")

			if err := r.Get(ctx, req.NamespacedName, &grpcburner); err != nil {
				log.Error(err, "Failed to re-fetch GrpcBurner before adding finalizer")
				return ctrl.Result{}, err
			}

			controllerutil.AddFinalizer(&grpcburner, grpcburnerFinalizer)
			if err := r.Update(ctx, &grpcburner); err != nil {
				log.Error(err, "Failed to update GrpcBurner with finalizer")
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&grpcburner, grpcburnerFinalizer) {
			deployName := fmt.Sprintf("%s-burner", grpcburner.Name)
			deploy := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deployName,
					Namespace: grpcburner.Namespace,
				},
			}

			log.Info("Running finalizer: deleting Deployment", "deployment", deployName)

			if err := r.Delete(ctx, &deploy); client.IgnoreNotFound(err) != nil {
				log.Error(err, "Failed to delete Deployment during finalization", "deployment", deployName)
				metrics.ReconcileErrors.WithLabelValues("grpcburner").Inc()
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&grpcburner, grpcburnerFinalizer)
			if err := r.Update(ctx, &grpcburner); err != nil {
				log.Error(err, "Failed to remove finalizer from GrpcBurner")
				metrics.ReconcileErrors.WithLabelValues("grpcburner").Inc()
				return ctrl.Result{}, err
			}

			log.Info("Finalizer cleanup complete")
		}
	}
	err := r.Get(ctx, types.NamespacedName{Name: deployName, Namespace: grpcburner.Namespace}, &deploy)
	if errors.IsNotFound(err) {
		log.Info("Deployment not found, creating new one")
		deploy = generateDeployment(&grpcburner)
		if err := r.Create(ctx, &deploy); err != nil {
			log.Error(err, "Failed to create Deployment")
			metrics.ReconcileErrors.WithLabelValues("grpcburner").Inc()
			return ctrl.Result{}, err
		}
		log.Info("Deployment created successfully", "deployment", deploy.Name)
	} else if err != nil {
		log.Error(err, "Failed to fetch Deployment")
		metrics.ReconcileErrors.WithLabelValues("grpcburner").Inc()
		return ctrl.Result{}, err
	} else {
		desired := generateDeployment(&grpcburner)
		if !isDeploymentSpecEqual(deploy.Spec, desired.Spec) {
			log.Info("Updating Deployment spec to match desired state")

			if err := r.Get(ctx, types.NamespacedName{Name: deployName, Namespace: grpcburner.Namespace}, &deploy); err != nil {
				log.Error(err, "Failed to re-fetch Deployment before update")
				return ctrl.Result{}, err
			}

			deploy.Spec = desired.Spec
			if err := r.Update(ctx, &deploy); err != nil {
				log.Error(err, "Failed to update Deployment spec")
				return ctrl.Result{}, err
			}
			log.Info("Deployment spec updated", "deployment", deploy.Name)
		}
	}

	phase := "Pending"
	if deploy.Status.ReadyReplicas == grpcburner.Spec.Replicas {
		phase = "Running"
	} else if deploy.Status.UnavailableReplicas > 0 {
		phase = "Failed"
	}

	updated := false
	if grpcburner.Status.ReadyReplicas != deploy.Status.ReadyReplicas {
		log.Info("Updating status.ReadyReplicas", "old", grpcburner.Status.ReadyReplicas, "new", deploy.Status.ReadyReplicas)
		grpcburner.Status.ReadyReplicas = deploy.Status.ReadyReplicas
		updated = true
	}
	if grpcburner.Status.Phase != phase {
		log.Info("Updating status.Phase", "old", grpcburner.Status.Phase, "new", phase)
		grpcburner.Status.Phase = phase
		grpcburner.Status.LastRunTime = metav1.Now()
		updated = true
	}

	if updated {
		var fresh grpcv1alpha1.GrpcBurner
		if err := r.Get(ctx, req.NamespacedName, &fresh); err != nil {
			log.Error(err, "Failed to re-fetch GrpcBurner before status update")
			metrics.ReconcileErrors.WithLabelValues("grpcburner").Inc()
			return ctrl.Result{}, err
		}

		fresh.Status.ReadyReplicas = grpcburner.Status.ReadyReplicas
		fresh.Status.Phase = grpcburner.Status.Phase
		fresh.Status.LastRunTime = grpcburner.Status.LastRunTime

		if err := r.Status().Update(ctx, &fresh); err != nil {
			log.Error(err, "Failed to update GrpcBurner status")
			metrics.ReconcileErrors.WithLabelValues("grpcburner").Inc()
			return ctrl.Result{}, err
		}
		log.Info("GrpcBurner status updated")
	}

	log.Info("Reconciliation complete")
	return ctrl.Result{}, nil
}

func (r *GrpcBurnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grpcv1alpha1.GrpcBurner{}).
		Complete(r)
}

func generateDeployment(grpcburner *grpcv1alpha1.GrpcBurner) appsv1.Deployment {
	labels := map[string]string{
		"app": grpcburner.Name,
	}

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      grpcburner.Name + "-burner",
			Namespace: grpcburner.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &grpcburner.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "grpc-burner",
							Image: "stsukada/grpc-burner-demo:latest",
							Args: []string{
								"--mode", grpcburner.Spec.Mode,
								"--message-size", fmt.Sprintf("%d", grpcburner.Spec.MessageSize),
								"--qps", fmt.Sprintf("%d", grpcburner.Spec.QPS),
								"--duration", grpcburner.Spec.Duration,
							},
							Resources: grpcburner.Spec.Resources,
						},
					},
				},
			},
		},
	}
}

func isDeploymentSpecEqual(a, b appsv1.DeploymentSpec) bool {
	return a.Replicas != nil && b.Replicas != nil &&
		*a.Replicas == *b.Replicas &&
		reflect.DeepEqual(a.Template.Spec.Containers, b.Template.Spec.Containers)
}
