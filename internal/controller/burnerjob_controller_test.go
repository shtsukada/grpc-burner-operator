package controller

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	burnerv1alpha1 "github.com/shtsukada/grpc-burner-operator/api/v1alpha1"
)

func TestBurnerJob_Reconcile_CreatesJob(t *testing.T) {
	ctx := context.Background()

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = burnerv1alpha1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)

	burnerJob := &burnerv1alpha1.BurnerJob{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      "test-burn",
			Namespace: "default",
		},
		Spec: burnerv1alpha1.BurnerJobSpec{
			TargetService: "grpc-service.default.svc:50051",
			QPS:           100,
			Duration:      "10s",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(burnerJob).
		Build()

	reconciler := &BurnerJobReconciler{
		Client: fakeClient,
		Log:    ctrl.Log.WithName("test"),
		Scheme: scheme,
	}

	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: clientpkg.ObjectKeyFromObject(burnerJob),
	})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	var createdJob batchv1.Job
	jobName := "test-burn-job"
	err = fakeClient.Get(ctx, clientpkg.ObjectKey{Namespace: "default", Name: jobName}, &createdJob)
	if err != nil {
		t.Fatalf("Expected Job %q to be created, but got error: %v", jobName, err)
	}
}
