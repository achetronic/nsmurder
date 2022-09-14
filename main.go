package main

import (
	"context"
	"log"
	"time"

	"nsmurder/flags"      // Configuration flags for the CLI
	"nsmurder/kubernetes" // Requests against Kubernetes API
	"nsmurder/manager"    // Requests against Kubernetes API
)

const (
	// General messages
	GetClientMessage                 = "Generating the client to connect to Kubernetes"
	ScheduleNamespaceDeletionMessage = "Scheduling namespaces for deletion"
	WaitTimeMessage                  = "Waiting prudential time between strategies: %s"
	DeleteOrphanApisMessage          = "Deleting orphan APIService resources"
	CleanResourcesMessage            = "Cleaning resources inside stuck namespaces"
	DeleteNamespaceByForceMessage    = "Deleting namespaces by using force"
	DeletionCompleteMessage          = "Scheduled namespaces have been deleted"

	// Error messages
	NamespacesRequiredErrorMessage        = "No namespaces specified. Use one of the following flags: --include or --include-all"
	KubernetesGetClientErrorMessage       = "error connecting to Kubernetes API: %s"
	ScheduleNamespaceDeletionErrorMessage = "error scheduling namespaces for deletion: %s"
	DeleteOrphanApisErrorMessage          = "error deleting orphan APIs: %s"
	CleanResourcesErrorMessage            = "error cleaning stuck namespaces: %s"
)

var inputFlags flags.FlagsSpec

//
func main() {
	ctx := context.Background()

	// 1. Parse the flags from the command line
	inputFlags.ParseFlags()

	// Several flags are used by the Manager module. To keep it independent, it inherits the real flags type
	// TODO: Look for a better way to do this
	var nsManager manager.Manager
	nsManager.Duration = inputFlags.Duration
	nsManager.IncludeAll = inputFlags.IncludeAll
	nsManager.Include = inputFlags.Include
	nsManager.Ignore = inputFlags.Ignore

	if !*inputFlags.IncludeAll && len(inputFlags.Include) <= 0 {
		log.Print(NamespacesRequiredErrorMessage)
		return
	}

	// 2. Generate the Kubernetes client to modify the resources
	log.Print(GetClientMessage)
	client := kubernetes.ConnectionClientsSpec{}

	err := client.SetClients()
	if err != nil {
		log.Printf(KubernetesGetClientErrorMessage, err)
		return
	}

	// 3. Schedule namespaces deletion
	log.Print(ScheduleNamespaceDeletionMessage)
	err = nsManager.ScheduleNamespaceDeletion(ctx, client)
	if err != nil {
		log.Printf(ScheduleNamespaceDeletionErrorMessage, err)
		return
	}

	// Wait a time between strategies to allow Kubernetes try to manage the garbage
	log.Printf(WaitTimeMessage, *inputFlags.Duration)
	time.Sleep(*inputFlags.Duration)

	// 4. Delete unavailable API services
	log.Print(DeleteOrphanApisMessage)
	err = kubernetes.DeleteOrphanApiServices(ctx, client.Dynamic)
	if err != nil {
		log.Printf(DeleteOrphanApisErrorMessage, err)
		return
	}

	// Wait a time between strategies to allow Kubernetes try to manage the garbage
	log.Printf(WaitTimeMessage, *inputFlags.Duration)
	time.Sleep(*inputFlags.Duration)

	// 5. Delete resources on stuck namespaces
	log.Print(CleanResourcesMessage)
	err = manager.CleanStuckNamespaces(ctx, client)
	if err != nil {
		log.Printf(CleanResourcesErrorMessage, err)
		return
	}

	// Wait a time between strategies to allow Kubernetes try to manage the garbage
	log.Printf(WaitTimeMessage, *inputFlags.Duration)
	time.Sleep(*inputFlags.Duration)

	// 6. Delete last namespaces by forcing deletion
	log.Print(DeleteNamespaceByForceMessage)
	err = manager.DeleteTerminatingNamespacesByForce(ctx, client)
	if err != nil {
		log.Print(err)
		return
	}

	// 7. Success
	log.Print(DeletionCompleteMessage)
}
