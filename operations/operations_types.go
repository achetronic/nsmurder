package operations

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

//
type KubernetesClientsSpec struct {
	Discovery *discovery.DiscoveryClient // Discovery operations about API resources
	Dynamic   dynamic.Interface          // Unstructured objects and operations
}

// StatusSpec represents the Status of any resource in Kubernetes
// This spec was initially added to find orphan APIService resources
// and not all the resources include conditions
type StatusSpec struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ExtendedGroupVersionKindSpec represents the minimal definition
// to find accurately a resource inside Kubernetes
type ExtendedGroupVersionKindSpec struct {
	Name         string
	SingularName string
	schema.GroupVersionKind
}
