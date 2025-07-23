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
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	grpcv1alpha1 "github.com/shtsukada/grpc-burner-operator/api/v1alpha1"
)

func TestGrpcBurner_Reconcile_CreatesDeployment(t *testing.T) {
	ctx := context.Background()

	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = grpcv1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	_ = corev1.AddToScheme(s)

	client := fake.NewClientBuilder().WithScheme(s).Build()

	grpc := &grpcv1alpha1.GrpcBurner{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      "test-grpcburner",
			Namespace: "default",
		},
		Spec: grpcv1alpha1.GrpcBurnerSpec{
			Replicas:    1,
			Mode:        "unary",
			MessageSize: 100,
			QPS:         10,
			Duration:    "10s",
		},
	}

	require.NoError(t, client.Create(ctx, grpc))

	reconciler := &GrpcBurnerReconciler{
		Client: client,
		Scheme: s,
	}
	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      grpc.Name,
			Namespace: grpc.Namespace,
		},
	})
	require.NoError(t, err)

	var deploy appsv1.Deployment
	err = client.Get(ctx, types.NamespacedName{
		Name:      "test-grpcburner-burner",
		Namespace: "default",
	}, &deploy)
	require.NoError(t, err)
	require.Equal(t, int32(1), *deploy.Spec.Replicas)

	var updated grpcv1alpha1.GrpcBurner
	err = client.Get(ctx, types.NamespacedName{
		Name:      grpc.Name,
		Namespace: grpc.Namespace,
	}, &updated)
	require.NoError(t, err)
	require.Contains(t, updated.Finalizers, grpcburnerFinalizer)
}
