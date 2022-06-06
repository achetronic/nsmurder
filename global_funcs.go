package main

import (
	"context"
	"errors"
	"fmt"
	"nsmurder/operations"
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
func ScheduleNamespaceDeletion(ctx context.Context, client operations.KubernetesClientsSpec) (err error) {

	var namespaces []string

	namespaces = flags.GetNamespaces()
	if *flags.IncludeAll {
		namespaces, err = operations.GetNamespaces(ctx, client.Dynamic)
	}

	if err != nil {
		return errors.New(GetNamespacesErrorMessage)
	}

	// Schedule deletion for desired namespaces
	err = operations.DeleteNamespaces(ctx, client.Dynamic, namespaces)
	if err != nil {
		return errors.New(DeleteNamespaceErrorMessage)
	}

	return err
}

// CleanNamespace delete all resources from a namespace
func CleanNamespace(ctx context.Context, client operations.KubernetesClientsSpec,
	namespace string, namespacedApiResources []operations.ExtendedGroupVersionKindSpec) (err error) {

	// Loop over all available resource types
	for _, apiResource := range namespacedApiResources {

		// Get all resources of that type into the namespace
		resources, err := operations.GetResources(ctx, client.Dynamic,
			apiResource.Group,
			apiResource.Version,
			apiResource.Name,
			namespace)

		if err != nil {
			return err
		}

		// Delete all the resources of that type from the namespace
		for _, resource := range resources {
			err := operations.DeleteResource(ctx, client.Dynamic,
				apiResource.Group,
				apiResource.Version,
				apiResource.Name,
				resource.GetName(),
				namespace)

			if err != nil {
				return err
			}
		}
	}

	return err
}

// CleanStuckNamespaces delete all resources on stuck namespaces
func CleanStuckNamespaces(ctx context.Context, client operations.KubernetesClientsSpec) (err error) {

	// Get all resources able to be created into a namespace
	apiResources, err := operations.GetNamespacedApiResources(ctx, client.Discovery)
	if err != nil {
		return errors.New(GetNamespacedApiResourcesErrorMessage)
	}

	// Get all namespaces in phase 'Terminating'
	terminatingNamespaces, err := operations.GetTerminatingNamespaces(ctx, client.Dynamic)
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
