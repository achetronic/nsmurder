package operations

import (
	"context"
	"encoding/json"
	"strings"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

const (
	StatusConditionTypeAvailable = "Available"
)

// Global references
// Ref: https://caiorcferreira.github.io/post/the-kubernetes-dynamic-client/
// Ref: https://pkg.go.dev/k8s.io/client-go@v0.24.1/dynamic
// Ref: https://itnext.io/generically-working-with-kubernetes-resources-in-go-53bce678f887
// Ref: https://pkg.go.dev/k8s.io/client-go/discovery#DiscoveryClient

// GetResources ---
func GetResources(ctx context.Context, client dynamic.Interface,
	group string, version string, kind string, namespace string) (
	[]unstructured.Unstructured, error) {

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: kind,
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

// DeleteResources ---
func DeleteResource(ctx context.Context, client dynamic.Interface,
	group string, version string, kind string, name string, namespace string) error {

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: kind,
	}

	err := client.
		Resource(resourceId).
		Namespace(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})

	return err
}

// getNamespaces get a list of all namespaces existing in the cluster
func getNamespaces(ctx context.Context, client dynamic.Interface) (namespaces []unstructured.Unstructured, err error) {

	namespaceList, err := GetResources(ctx, client, "", "v1", "namespaces", "")

	if err != nil {
		return namespaces, err
	}

	return namespaceList, err
}

// GetNamespaces get a list of all namespaces existing in the cluster
func GetNamespaces(ctx context.Context, client dynamic.Interface) (namespaces []string, err error) {

	namespaceList, err := getNamespaces(ctx, client)

	if err != nil {
		return namespaces, err
	}

	for _, value := range namespaceList {
		namespaces = append(namespaces, value.GetName())
	}

	return namespaces, err
}

// GetStuckNamespaces get a list of 'Terminating' namespaces
func GetTerminatingNamespaces(ctx context.Context, client dynamic.Interface) (namespaces []string, err error) {

	namespaceList, err := getNamespaces(ctx, client)

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

	for _, namespaceName := range namespaces {
		err := DeleteResource(ctx, client, "", "v1", "namespaces", namespaceName, "")

		if err != nil {
			break
		}
	}

	return err
}

// GetOrphanApiServices get a list of all APIServices existing in the cluster
func GetOrphanApiServices(ctx context.Context, client dynamic.Interface) (apiServices []string, err error) {

	var currentStatus StatusSpec

	// Get all the APIService resources
	apiServiceList, err := GetResources(ctx, client, "apiregistration.k8s.io", "v1", "apiservices", "")

	if err != nil {
		return apiServices, err
	}

	// Add APIServices to the list when not available
	for _, apiService := range apiServiceList {

		// Check if the ApiService is orphan
		apiServiceJson, _ := json.Marshal(apiService.Object["status"])
		err := json.Unmarshal(apiServiceJson, &currentStatus)
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

	// Remove APIService resources
	for _, orphanApiService := range orphanApiServices {
		err = DeleteResource(ctx, client, "apiregistration.k8s.io", "v1", "apiservices", orphanApiService, "")
		if err != nil {
			return err
		}
	}

	return err
}

// GetNamespacedApiResources return a list with essential data about all namespaced resource types in the cluster
func GetNamespacedApiResources(ctx context.Context, client *discovery.DiscoveryClient) (
	namespacedApiResources []ExtendedGroupVersionKindSpec, err error) {

	extendedApiResource := &ExtendedGroupVersionKindSpec{}

	// Ask the API for all preferred namespaced resources
	apiResourceLists, err := client.ServerPreferredNamespacedResources()
	if err != nil {
		return namespacedApiResources, err
	}

	// Store only useful information about retrieved resources
	for _, apiResourceList := range apiResourceLists {

		for _, apiResource := range apiResourceList.APIResources {

			// Assume there is no Group but only Version
			extendedApiResource.Group = ""
			extendedApiResource.Version = apiResourceList.GroupVersion

			// Separate Group and Version
			groupVersion := strings.Split(apiResourceList.GroupVersion, "/")
			if len(groupVersion) == 2 {
				extendedApiResource.Group = groupVersion[0]
				extendedApiResource.Version = groupVersion[1]
			}

			// Fill the rest of data about the resource
			extendedApiResource.Kind = apiResource.Kind
			extendedApiResource.Name = apiResource.Name
			extendedApiResource.SingularName = apiResource.SingularName

			namespacedApiResources = append(namespacedApiResources, *extendedApiResource)
		}
	}

	return namespacedApiResources, err
}

//
// func CleanNamespacesResources(ctx context.Context, client *discovery.DiscoveryClient, namespaces []string) (err error) {

// 	namespacedApiResources, err = GetNamespacedApiResources(ctx, client)
// 	if err != nil {
// 		log.Print("lets see")
// 	}

// 	//
// 	algo, err := GetTerminatingNamespaces(ctx, client)
// 	if err != nil {
// 		log.Print(algo)
// 	}

// 	// Pillar los namespaces atascados

// }
