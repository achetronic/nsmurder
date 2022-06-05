package main

import (
	"context"
	"log"

	flagsPkg "nsmurder/flags" // Flags for this CLI
	"nsmurder/operations"     // Operations against Kubernetes API

	"k8s.io/client-go/discovery"          // Ref: https://pkg.go.dev/k8s.io/client-go/discovery#DiscoveryClient
	"k8s.io/client-go/dynamic"            // Ref: https://pkg.go.dev/k8s.io/client-go@v0.24.1/dynamic
	ctrl "sigs.k8s.io/controller-runtime" // Ref: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config
)

const (
	// General message
	GetClientMessage = "Generating the client to connect to Kubernetes"

	// Error messages
	KubernetesGetClientErrorMessage = "error connecting to Kubernetes API: %s"
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

	// Create the clients to do requests to out friend: Kubernetes
	client := operations.KubernetesClientsSpec{}

	client.Dynamic, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Printf(KubernetesGetClientErrorMessage, err)
	}

	client.Discovery, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Printf(KubernetesGetClientErrorMessage, err)
	}

	// ------------------------------------------------------------------------
	// Schedule namespaces deletion
	err = ScheduleNamespaceDeletion(ctx, client)
	if err != nil {
		log.Print(err)
	}

	// TODO: Implement a time to wait between processes to let Kubernetes to manage the situation

	// ---------------------------------------------------------------
	// Delete unavailable API services
	log.Print("Deleting orphan APIService resources")
	err = operations.DeleteOrphanApiServices(ctx, client.Dynamic)
	if err != nil {
		log.Print(err)
	}

	// ---------------------------------------------------------------
	// Delete resources on stuck namespaces
	log.Print("Cleaning resources inside stuck namespaces")
	err = CleanStuckNamespaces(ctx, client)
	if err != nil {
		log.Print(err)
	}

	// TODO: Implement a time to wait between processes to let Kubernetes to manage the situation

	// ---------------------------------------------------------------

}
