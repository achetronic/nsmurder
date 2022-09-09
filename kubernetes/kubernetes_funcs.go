package kubernetes

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"

	// Kubernetes clients
	"k8s.io/client-go/discovery"          // Ref: https://pkg.go.dev/k8s.io/client-go/discovery
	"k8s.io/client-go/dynamic"            // Ref: https://pkg.go.dev/k8s.io/client-go/dynamic
	ctrl "sigs.k8s.io/controller-runtime" // Ref: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config

	// Kubernetes types
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const (
	StatusConditionTypeAvailable = "Available"
)

// SetClients configure the clients needed to perform requests to Kubernetes API
func (c *ConnectionClientsSpec) SetClients() (err error) {

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	// Create the clients to do requests to out friend: Kubernetes
	c.Dynamic, err = dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	c.Discovery, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}

	return err
}

// GetNamespacedApiResources return a list with essential data about all namespaced resource types in the cluster
func GetNamespacedApiResources(client *discovery.DiscoveryClient) (resources []ResourceTypeSpec, err error) {

	extendedApiResource := &ResourceTypeSpec{}

	// Ask the API for all preferred namespaced resources
	apiResourceLists, err := client.ServerPreferredNamespacedResources()
	if err != nil {
		return resources, err
	}

	// Store only useful information about retrieved resources
	for _, apiResourceList := range apiResourceLists {

		for _, apiResource := range apiResourceList.APIResources {

			// Assume there is no Group but only Version
			extendedApiResource.GVK.Group = ""
			extendedApiResource.GVK.Version = apiResourceList.GroupVersion

			// Separate Group and Version
			groupVersion := strings.Split(apiResourceList.GroupVersion, "/")
			if len(groupVersion) == 2 {
				extendedApiResource.GVK.Group = groupVersion[0]
				extendedApiResource.GVK.Version = groupVersion[1]
			}

			// Fill the rest with data about the resource
			extendedApiResource.GVK.Kind = apiResource.Kind
			extendedApiResource.Name = apiResource.Name
			extendedApiResource.SingularName = apiResource.SingularName

			resources = append(resources, *extendedApiResource)
		}
	}

	return resources, err
}

// GetResources find resources of a certain type in the cluster
func GetResources(ctx context.Context, client dynamic.Interface, resourceType ResourceTypeSpec, namespace string) (
	[]unstructured.Unstructured, error) {

	resourceId := schema.GroupVersionResource{
		Group:    resourceType.GVK.Group,
		Version:  resourceType.GVK.Version,
		Resource: resourceType.Name,
	}

	list, err := client.
		Resource(resourceId).
		Namespace(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

// DeleteResource delete a resource from the cluster
func DeleteResource(ctx context.Context, client dynamic.Interface, resource ResourceSpec) (err error) {

	err = client.
		Resource(resource.GroupVersionResource).
		Namespace(resource.Namespace).
		Delete(ctx, resource.Name, metav1.DeleteOptions{})

	return err
}

// GetNamespaces get a list of all namespaces existing in the cluster
func GetNamespaces(ctx context.Context, client dynamic.Interface) (namespaces []unstructured.Unstructured, err error) {

	namespacesType := ResourceTypeSpec{}
	namespacesType.GVK.Group = ""
	namespacesType.GVK.Version = "v1"
	namespacesType.Name = "namespaces"

	namespaceList, err := GetResources(ctx, client, namespacesType, "")

	if err != nil {
		return namespaces, err
	}

	return namespaceList, err
}

// GetTerminatingNamespaces get a list of 'Terminating' namespaces
func GetTerminatingNamespaces(ctx context.Context, client dynamic.Interface) (namespaces []string, err error) {

	namespaceList, err := GetNamespaces(ctx, client)

	if err != nil {
		return namespaces, err
	}

	// Add namespaces with deletion timestamp, which indicates deletion
	for _, namespace := range namespaceList {

		deletionTimestamp := namespace.GetDeletionTimestamp()

		if !deletionTimestamp.IsZero() {
			namespaces = append(namespaces, namespace.GetName())
		}
	}

	return namespaces, err
}

// DeleteNamespaces schedule namespaces for deletion
func DeleteNamespaces(ctx context.Context, client dynamic.Interface, namespaces []string) (err error) {

	resource := ResourceSpec{}

	for _, namespaceName := range namespaces {

		//log.Printf("Trying to delete namespace: %s\n", namespaceName)

		resource.Group = ""
		resource.Version = "v1"
		resource.Resource = "namespaces"
		resource.Namespace = ""
		resource.Name = namespaceName

		err = DeleteResource(ctx, client, resource)

		if err != nil {

			// IsNotFound is not an error. The function is trying to delete
			if errors.IsNotFound(err) {
				err = nil
				continue
			}
			break
		}
	}

	return err
}

// GetOrphanApiServices get a list of all APIServices existing in the cluster
func GetOrphanApiServices(ctx context.Context, client dynamic.Interface) (apiServices []string, err error) {

	var currentStatus StatusSpec

	apiServicesType := ResourceTypeSpec{}
	apiServicesType.Name = "apiservices"
	apiServicesType.GVK.Group = "apiregistration.k8s.io"
	apiServicesType.GVK.Version = "v1"

	// Get all the APIService resources
	apiServiceList, err := GetResources(ctx, client, apiServicesType, "")

	if err != nil {
		return apiServices, err
	}

	// Add APIServices to the list when not available
	for _, apiService := range apiServiceList {

		// Check if the ApiService is orphan
		// TODO: Look for a better way to do this
		apiServiceJson, _ := json.Marshal(apiService.Object["status"])
		err = json.Unmarshal(apiServiceJson, &currentStatus)
		if err != nil {
			return apiServices, err
		}

		// Append it to the list when orphaned
		if meta.IsStatusConditionFalse(currentStatus.Conditions, StatusConditionTypeAvailable) {
			apiServices = append(apiServices, apiService.GetName())
		}
	}

	return apiServices, err
}

// DeleteOrphanApiServices delete all APIService resources which are not 'Available'
func DeleteOrphanApiServices(ctx context.Context, client dynamic.Interface) (err error) {

	orphanApiServices, err := GetOrphanApiServices(ctx, client)
	if err != nil {
		return err
	}

	resource := ResourceSpec{}
	resource.Group = "apiregistration.k8s.io"
	resource.Version = "v1"
	resource.Resource = "apiservices"
	resource.Namespace = ""

	// Remove APIService resources
	for _, orphanApiService := range orphanApiServices {

		resource.Name = orphanApiService

		err = DeleteResource(ctx, client, resource)

		if err != nil {
			if errors.IsNotFound(err) {
				err = nil
				continue
			}
			break
		}
	}

	return err
}

// DeleteResourceFinalizers delete finalizers from the given resource
func DeleteResourceFinalizers(ctx context.Context, client dynamic.Interface, resource ResourceSpec) (err error) {

	patchBytes := []byte(`[{"op":"remove","path":"/metadata/finalizers"}]`)
	patchOptions := metav1.PatchOptions{}

	_, err = client.
		Resource(resource.GroupVersionResource).
		Namespace(resource.Namespace).
		Patch(ctx, resource.Name, types.JSONPatchType, patchBytes, patchOptions)

	if err != nil {
		return err
	}

	return err
}
