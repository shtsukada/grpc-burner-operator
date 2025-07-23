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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrpcBurnerSpec defines the desired state of GrpcBurner
type GrpcBurnerSpec struct {
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`

	// +kubebuilder:validation:Enum=unary;server-streaming;client-streaming;bidirectional-streaming
	Mode string `json:"mode"`

	// +kubebuilder:validation:Minimum=1
	MessageSize int32 `json:"messageSize"`

	// +kubebuilder:validation:Minimum=1
	QPS int32 `json:"qps"`

	// +kubebuilder:validation:Pattern=^\d+[smh]$
	Duration string `json:"duration"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// GrpcBurnerStatus defines the observed state of GrpcBurner
type GrpcBurnerStatus struct {
	ReadyReplicas int32       `json:"readyReplicas,omitempty"`
	Phase         string      `json:"phase,omitempty"`
	LastRunTime   metav1.Time `json:"lastRunTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GrpcBurner is the Schema for the grpcburners API
type GrpcBurner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrpcBurnerSpec   `json:"spec,omitempty"`
	Status GrpcBurnerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GrpcBurnerList contains a list of GrpcBurner
type GrpcBurnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrpcBurner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrpcBurner{}, &GrpcBurnerList{})
}
