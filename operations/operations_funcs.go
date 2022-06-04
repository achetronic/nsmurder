package operations

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetResources ---
// Ref: https://pkg.go.dev/k8s.io/client-go@v0.24.1/dynamic
// Ref: https://itnext.io/generically-working-with-kubernetes-resources-in-go-53bce678f887
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
// Ref: https://pkg.go.dev/k8s.io/client-go@v0.24.1/dynamic
// Ref: https://itnext.io/generically-working-with-kubernetes-resources-in-go-53bce678f887
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

// GetAllNamespaces get a list of all namespaces existing in the cluster
func GetAllNamespaces(ctx context.Context, client dynamic.Interface) (namespaces []string, err error) {

	namespaceList, err := GetResources(ctx, client, "", "v1", "namespaces", "")

	if err != nil {
		return namespaces, err
	}

	for _, value := range namespaceList {
		namespaces = append(namespaces, value.GetName())
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

	apiServiceList, err := GetResources(ctx, client, "apiregistration.k8s.io", "v1", "apiservices", "")

	if err != nil {
		return apiServices, err
	}

	for _, apiService := range apiServiceList {
		// Check if the ApiService is orphan

		// Append it to the list
		apiServices = append(apiServices, apiService.GetName())
	}

	return apiServices, err
}
