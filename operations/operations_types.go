package operations

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StatusSpec represents the Status of any resource in Kubernetes
// This spec was initially added for APIService resources
type StatusSpec struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ExtendedGroupVersionKindSpec represents the minimal definition
// to be able to find a resource inside Kubernetes
type ExtendedGroupVersionKindSpec struct {
	Name         string `json:"name,omitempty"`
	SingularName string `json:"singular_name,omitempty"`
	schema.GroupVersionKind
}
