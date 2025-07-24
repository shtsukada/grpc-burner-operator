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
	"os"

	// "path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	grpcv1alpha1 "github.com/shtsukada/grpc-burner-operator/api/v1alpha1"
)

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestMain(m *testing.M) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())
	defer cancel()

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = grpcv1alpha1.AddToScheme(scheme)

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{"../../config/crd/bases"},
	}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	err = (&GrpcBurnerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic(err)
		}
	}()

	code := m.Run()

	_ = testEnv.Stop()
	os.Exit(code)
}

func TestGrpcBurnerE2E(t *testing.T) {
	ctx := context.Background()

	grpc := &grpcv1alpha1.GrpcBurner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-sample",
			Namespace: "default",
		},
		Spec: grpcv1alpha1.GrpcBurnerSpec{
			Replicas:    1,
			Mode:        "unary",
			MessageSize: 128,
			QPS:         5,
			Duration:    "10s",
		},
	}

	require.NoError(t, k8sClient.Create(ctx, grpc))

	var deploy appsv1.Deployment
	key := types.NamespacedName{
		Name:      "e2e-sample-burner",
		Namespace: "default",
	}

	tryUntil(t, 5*time.Second, func() bool {
		err := k8sClient.Get(ctx, key, &deploy)
		return err == nil
	}, "Deployment should be created")

	require.Equal(t, int32(1), *deploy.Spec.Replicas)

	var updated grpcv1alpha1.GrpcBurner
	tryUntil(t, 5*time.Second, func() bool {
		_ = k8sClient.Get(ctx, types.NamespacedName{
			Name:      grpc.Name,
			Namespace: grpc.Namespace,
		}, &updated)
		return updated.Status.Phase != ""
	}, "Status should be updated")

	require.Contains(t, []string{"Pending", "Running"}, updated.Status.Phase)
}

func tryUntil(t *testing.T, timeout time.Duration, condition func() bool, msg string) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("timeout exceeded: %s", msg)
}
