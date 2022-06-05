package main

import (
	"context"
	"errors"
	"log"
	"nsmurder/operations"
)

const (

	// Error messages
	GetNamespacesErrorMessage             = "error getting the namespaces: %s"
	DeleteNamespaceErrorMessage           = "error deleting namespace: %s"
	GetNamespacedApiResourcesErrorMessage = "error getting namespaced API resources"
	GetTerminatingNamespacesErrorMessage  = "error getting terminating namespaces"
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

// TODO implement the logic for this
// CleanStuckNamespaces delete all resources on stuck namespaces
func CleanStuckNamespaces(ctx context.Context, client operations.KubernetesClientsSpec) (err error) {

	apiResources, err := operations.GetNamespacedApiResources(ctx, client.Discovery)
	if err != nil {
		return errors.New(GetNamespacedApiResourcesErrorMessage)
	}

	//
	terminatingNamespaces, err := operations.GetTerminatingNamespaces(ctx, client.Dynamic)
	if err != nil {
		return errors.New(GetTerminatingNamespacesErrorMessage)
	}

	// Pasar por los namespaces buscando los recursos que se pueden, y eliminarlos
	log.Print(apiResources)
	log.Print(terminatingNamespaces)

	return err

}
