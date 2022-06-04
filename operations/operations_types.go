package operations

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatusSpec represents the Status of any resource in Kubernetes
// This spec was initially added for APIService resources so it could fail with different CRDs
type StatusSpec struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
