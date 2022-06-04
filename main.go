package main

import (
	"context"
	"log"

	flagsPkg "nsmurder/flags" // Flags for this CLI
	"nsmurder/operations"     // Operations against Kubernetes API

	"k8s.io/client-go/dynamic"            // Ref: https://pkg.go.dev/k8s.io/client-go@v0.24.1/dynamic
	ctrl "sigs.k8s.io/controller-runtime" // Ref: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config
)

const (
	// General message
	GetClientMessage = "Generating the client to connect to Kubernetes"

	// Error messages
	KubernetesGetClientErrorMessage = "Error connecting to Kubernetes API: %s"
	GetNamespacesErrorMessage       = "Error getting the namespaces: %s"
	DeleteNamespaceErrorMessage     = "Error deleting namespace: %s"
)

var flags flagsPkg.FlagsSpec

//
func main() {
	ctx := context.Background()

	// Parse the flags from the command line
	flags.ParseFlags()

	// Generate the Kubernetes client to modify the resources
	log.Print(GetClientMessage)
	config, err := ctrl.GetConfig()
	if err != nil {
		log.Printf(KubernetesGetClientErrorMessage, err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Printf(KubernetesGetClientErrorMessage, err)
	}

	// Get the namespace to terminate
	var namespaces []string

	namespaces = flags.GetNamespaces()
	if *flags.IncludeAll {
		namespaces, err = operations.GetAllNamespaces(ctx, client)
	}

	if err != nil {
		log.Printf(GetNamespacesErrorMessage, err)
	}

	log.Print(namespaces)

	// Schedule deletion for desired namespaces
	err = operations.DeleteNamespaces(ctx, client, namespaces)
	if err != nil {
		log.Printf(DeleteNamespaceErrorMessage, err)
	}

	// TODO: Implement a time to wait between processes to let Kubernetes to manage the situation

	// Delete unavailable APIs
	apiServices, err := operations.GetOrphanApiServices(ctx, client)
	if err != nil {
		log.Print("lets see")
	}

	log.Print(apiServices)

	// Delete stuck namespace's resources

	// TODO: Implement a time to wait between processes to let Kubernetes to manage the situation

}
