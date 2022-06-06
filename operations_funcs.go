package main

import (
	"context"
	"errors"
	"fmt"

	"nsmurder/kubernetes"
)

const (

	// Error messages
	GetNamespacesErrorMessage             = "error getting the namespaces: %s"
	DeleteNamespaceErrorMessage           = "error deleting namespace: %s"
	GetNamespacedApiResourcesErrorMessage = "error getting namespaced API resources"
	GetTerminatingNamespacesErrorMessage  = "error getting terminating namespaces"
	CleanNamespaceErrorMessage            = "error cleaning a namespace: %s"
)

// ScheduleNamespaceDeletion schedule deletion for all selected namespaces according to the CLI flags
func ScheduleNamespaceDeletion(ctx context.Context, client kubernetes.ConnectionClientsSpec) (err error) {

	var namespaces []string

	namespaces = inputFlags.GetNamespaces()

	if *inputFlags.IncludeAll {
		namespaces, err = kubernetes.GetNamespaces(ctx, client.Dynamic)
	}

	if err != nil {
		return errors.New(GetNamespacesErrorMessage)
	}

	// Schedule deletion for desired namespaces
	err = kubernetes.DeleteNamespaces(ctx, client.Dynamic, namespaces)
	if err != nil {
		return errors.New(DeleteNamespaceErrorMessage)
	}

	return err
}

// CleanNamespace delete given resource types from a namespace
func CleanNamespace(ctx context.Context, client kubernetes.ConnectionClientsSpec,
	namespace string, ResourceTypes []kubernetes.ResourceTypeSpec) (err error) {

	apiResourceType := &kubernetes.ResourceTypeSpec{}
	currentResource := &kubernetes.ResourceSpec{}

	// Loop over all given resource types
	for _, resourceType := range ResourceTypes {

		apiResourceType.GVK.Group = resourceType.GVK.Group
		apiResourceType.GVK.Version = resourceType.GVK.Version
		apiResourceType.Name = resourceType.Name

		// Get all resources of current type from the namespace
		resources, err := kubernetes.GetResources(ctx, client.Dynamic, *apiResourceType, namespace)

		if err != nil {
			return err
		}

		// Delete all the resources of that type from the namespace
		for _, resource := range resources {

			currentResource.Group = resourceType.GVK.Group
			currentResource.Version = resourceType.GVK.Version
			currentResource.Resource = resourceType.Name
			currentResource.Name = resource.GetName()
			currentResource.Namespace = namespace

			err := kubernetes.DeleteResource(ctx, client.Dynamic, *currentResource)

			if err != nil {
				return err
			}
		}
	}

	return err
}

// CleanStuckNamespaces delete all resources on stuck namespaces
func CleanStuckNamespaces(ctx context.Context, client kubernetes.ConnectionClientsSpec) (err error) {

	// Get all resources able to be created into a namespace
	apiResources, err := kubernetes.GetNamespacedApiResources(ctx, client.Discovery)
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
		err := CleanNamespace(ctx, client, namespace, apiResources)
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
	namespaces, err := kubernetes.GetTerminatingNamespaces(ctx, client.Dynamic)
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
	}

	return err
}
