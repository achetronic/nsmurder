package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

// ConnectionClientsSpec represents the group of the connectors needed by this CLI
// to call the Kubernetes API on any scenario
type ConnectionClientsSpec struct {
	Discovery *discovery.DiscoveryClient // Discovery operations about API resources
	Dynamic   dynamic.Interface          // Unstructured objects and operations
}

// ResourceTypeSpec represents a resource type inside Kubernetes
type ResourceTypeSpec struct {
	Name         string
	SingularName string
	GVK          schema.GroupVersionKind // Group Version Kind
}

// ResourceSpec represents a resource inside Kubernetes
type ResourceSpec struct {
	types.NamespacedName
	schema.GroupVersionResource
}

// StatusSpec represents the Status of any resource in Kubernetes
// This spec was initially added to find orphan APIService resources
// and not all the resources include conditions
type StatusSpec struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
