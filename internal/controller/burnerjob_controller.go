package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	burnerv1alpha1 "github.com/shtsukada/grpc-burner-operator/api/v1alpha1"
)

type BurnerJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=grpc.burner.dev,resources=burnerjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=grpc.burner.dev,resources=burnerjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=grpc.burner.dev,resources=burnerjobs/finalizers,verbs=update

func (r *BurnerJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues(
		"controller", "burnerjob",
		"name", req.Name,
		"namespace", req.Namespace,
	)

	tracer := otel.Tracer("controller.burnerjob")
	ctx, span := tracer.Start(ctx, "Reconcile")
	defer span.End()

	burnerjob := &burnerv1alpha1.BurnerJob{}
	if err := r.Get(ctx, req.NamespacedName, burnerjob); err != nil {
		log.Error(err, "unable to fetch BurnerJob")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("target", burnerjob.Spec.TargetService)
	log.Info("Reconciling BurnerJob")

	jobName := req.Name + "-job"
	var existingJob batchv1.Job
	err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: req.Namespace}, &existingJob)
	if err == nil {
		log.Info("Job already exists, skipping creation", "job", jobName)
	} else if !apierrors.IsNotFound(err) {
		log.Error(err, "Failed to get existing Job")
		return ctrl.Result{}, err
	} else {
		log.Info("Creating new Job", "job", jobName)

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: burnerjob.Namespace,
				Labels:    map[string]string{"burnerjob": burnerjob.Name},
			},
			Spec: batchv1.JobSpec{
				TTLSecondsAfterFinished: ptr.To(int32(30)),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "grpc-burner",
								Image: "stsukada/grpc-burner-demo:latest",
								Args: []string{
									"--target", burnerjob.Spec.TargetService,
									"--qps", fmt.Sprintf("%d", burnerjob.Spec.QPS),
									"--duration", burnerjob.Spec.Duration,
								},
							},
						},
						RestartPolicy: corev1.RestartPolicyNever,
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(burnerjob, job, r.Scheme); err != nil {
			log.Error(err, "Failed to set controller reference")
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "Failed to create Job", "job", job.Name)
			return ctrl.Result{}, err
		}
		log.Info("Created new Job", "job", job.Name)
	}

	var phase string
	var message string

	for _, c := range existingJob.Status.Conditions {
		switch c.Type {
		case batchv1.JobComplete:
			if c.Status == corev1.ConditionTrue {
				phase = "Succeeded"
				message = c.Message
			}
		case batchv1.JobFailed:
			if c.Status == corev1.ConditionTrue {
				phase = "Failed"
				message = c.Message
			}
		}
	}

	updated := false
	if burnerjob.Status.Phase != phase {
		log.Info("Updating phase", "old", burnerjob.Status.Phase, "new", phase)
		burnerjob.Status.Phase = phase
		updated = true
	}
	if burnerjob.Status.Message != message {
		log.Info("Updating message", "old", burnerjob.Status.Message, "new", message)
		burnerjob.Status.Message = message
		updated = true
	}

	if updated {
		if err := r.Status().Update(ctx, burnerjob); err != nil {
			log.Error(err, "Failed to update BurnerJob status")
			return ctrl.Result{}, err
		}
		log.Info("Updated BurnerJob status", "phase", phase)
	}

	log.Info("Reconciliation complete")
	return ctrl.Result{}, nil
}

func (r *BurnerJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&burnerv1alpha1.BurnerJob{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
