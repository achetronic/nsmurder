package manager

import (
	"context"
	"errors"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/utils/strings/slices"
	"log"
	"rodillo/flags"
	"rodillo/kubernetes"
)

const (

	// General messages
	DeletedNamespacedResourceMessage = "Deleted %s/%s resource of kind %s/%s"
	DeletedNamespaceMessage          = "Deleted namespace %s"

	// Error messages
	GetNamespacesErrorMessage             = "error getting the namespaces: %s"
	DeleteNamespaceErrorMessage           = "error deleting namespace: %s"
	GetNamespacedApiResourcesErrorMessage = "error getting namespaced API resources"
	GetTerminatingNamespacesErrorMessage  = "error getting terminating namespaces"
	CleanNamespaceErrorMessage            = "error cleaning a namespace: %s"
)

// Manager TODO
type Manager struct {
	flags.FlagsSpec
}

// GetNamespaces get a list of all namespaces existing in the cluster
func GetNamespaces(ctx context.Context, client dynamic.Interface) (namespaces []string, err error) {

	namespaceList, err := kubernetes.GetNamespaces(ctx, client)

	if err != nil {
		return namespaces, err
	}

	for _, value := range namespaceList {
		namespaces = append(namespaces, value.GetName())
	}

	return namespaces, err
}

// ScheduleNamespaceDeletion schedule deletion for all selected namespaces according to the CLI flags
func (m *Manager) ScheduleNamespaceDeletion(ctx context.Context, client kubernetes.ConnectionClientsSpec) (err error) {

	// Calculate included namespaces
	var tmpNamespaces []string

	tmpNamespaces = m.Include

	if *m.IncludeAll {
		tmpNamespaces, err = GetNamespaces(ctx, client.Dynamic)
		if err != nil {
			return errors.New(GetNamespacesErrorMessage)
		}
	}

	var namespaces []string

	// Delete ignored namespaces from list
	for _, ns := range tmpNamespaces {
		if !slices.Contains(m.Ignore, ns) {
			namespaces = append(namespaces, ns)
		}
	}

	// Schedule deletion for desired namespaces
	err = kubernetes.DeleteNamespaces(ctx, client.Dynamic, namespaces)
	if err != nil {
		return errors.New(DeleteNamespaceErrorMessage)
	}

	return err
}

// CleanNamespace delete all the resources of the given types from a namespace
func CleanNamespace(ctx context.Context, client kubernetes.ConnectionClientsSpec,
	namespace string, ResourceTypes []kubernetes.ResourceTypeSpec) (err error) {

	currentApiResourceType := &kubernetes.ResourceTypeSpec{}
	currentResource := &kubernetes.ResourceSpec{}

	// Loop over all given resource types
	for _, resourceType := range ResourceTypes {

		currentApiResourceType.GVK.Group = resourceType.GVK.Group
		currentApiResourceType.GVK.Version = resourceType.GVK.Version
		currentApiResourceType.Name = resourceType.Name

		// Get all resources of current type from the namespace
		var resources []unstructured.Unstructured
		resources, err = kubernetes.GetResources(ctx, client.Dynamic, *currentApiResourceType, namespace)

		if err != nil && !apierrors.IsMethodNotSupported(err) {
			return err
		}

		// Delete all the resources of that type from the namespace
		for _, resource := range resources {

			currentResource.Group = resourceType.GVK.Group
			currentResource.Version = resourceType.GVK.Version
			currentResource.Resource = resourceType.Name
			currentResource.Name = resource.GetName()
			currentResource.Namespace = namespace

			err = kubernetes.DeleteResource(ctx, client.Dynamic, *currentResource)
			if err != nil && !apierrors.IsMethodNotSupported(err) && !apierrors.IsNotFound(err) {
				return err
			}

			// Remove the finalizers of each deleted resource
			err = kubernetes.DeleteResourceFinalizers(ctx, client.Dynamic, *currentResource)
			if err != nil {
				return err
			}

			log.Printf(DeletedNamespacedResourceMessage,
				currentResource.Namespace,
				currentResource.Name,
				resource.GetAPIVersion(),
				resource.GetKind(),
			)
		}
	}

	return err
}

// CleanStuckNamespaces delete all resources on stuck namespaces
func CleanStuckNamespaces(ctx context.Context, client kubernetes.ConnectionClientsSpec) (err error) {

	// Get all resources able to be created into a namespace
	apiResources, err := kubernetes.GetNamespacedApiResources(client.Discovery)
	if err != nil {
		return errors.New(GetNamespacedApiResourcesErrorMessage)
	}

	// Get all namespaces in phase 'Terminating'
	terminatingNamespaces, err := kubernetes.GetTerminatingNamespaces(ctx, client.Dynamic)
	if err != nil {
		return errors.New(GetTerminatingNamespacesErrorMessage)
	}

	// Loop over the namespaces cleaning them
	for _, namespace := range terminatingNamespaces {
		err = CleanNamespace(ctx, client, namespace, apiResources)
		if err != nil {
			errorMessage := fmt.Sprintf(CleanNamespaceErrorMessage, err)
			return errors.New(errorMessage)
		}
	}

	return err
}

// DeleteTerminatingNamespacesByForce delete namespaces in 'Terminating' phase deleting finalizers by patching
func DeleteTerminatingNamespacesByForce(ctx context.Context, client kubernetes.ConnectionClientsSpec) (err error) {

	resource := &kubernetes.ResourceSpec{}
	resource.Group = ""
	resource.Version = "v1"
	resource.Resource = "namespaces"

	// Get terminating namespaces
	var namespaces []string
	namespaces, err = kubernetes.GetTerminatingNamespaces(ctx, client.Dynamic)
	if err != nil {
		return err
	}

	// Loop over namespaces patching finalizers
	for _, namespace := range namespaces {

		resource.Name = namespace

		err = kubernetes.DeleteResourceFinalizers(ctx, client.Dynamic, *resource)
		if err != nil {
			return err
		}

		log.Printf(DeletedNamespaceMessage, namespace)
	}

	return err
}
