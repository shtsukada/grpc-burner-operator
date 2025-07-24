package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BurnerJobSpec defines the desired state of BurnerJob
type BurnerJobSpec struct {
	TargetService string `json:"targetService"`
	QPS           int    `json:"qps"`
	Duration      string `json:"duration"`
}

// BurnerJobStatus defines the observed state of BurnerJob
type BurnerJobStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// BurnerJob is the Schema for the burnerjobs API
type BurnerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BurnerJobSpec   `json:"spec,omitempty"`
	Status BurnerJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BurnerJobList contains a list of BurnerJob
type BurnerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BurnerJob `json:"items"`
}
