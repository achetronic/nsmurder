package main

import (
	"context"
	"log"
	"time"

	"nsmurder/flags"      // Configuration flags for the CLI
	"nsmurder/kubernetes" // Requests against Kubernetes API
)

const (
	// General message
	GetClientMessage                 = "Generating the client to connect to Kubernetes"
	ScheduleNamespaceDeletionMessage = "Scheduling namespaces for deletion"
	WaitTimeMessage                  = "Waiting prudential time between strategies: %s"
	DeleteOrphanApisMessage          = "Deleting orphan APIService resources"
	CleanResourcesMessage            = "Cleaning resources inside stuck namespaces"
	DeleteNamespaceByForceMessage    = "Deleting namespaces by using force"

	// Error messages
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

	// 2. Generate the Kubernetes client to modify the resources
	log.Print(GetClientMessage)
	client := kubernetes.ConnectionClientsSpec{}

	err := client.SetClients()
	if err != nil {
		log.Printf(KubernetesGetClientErrorMessage, err)
	}

	// 3. Schedule namespaces deletion
	log.Print(ScheduleNamespaceDeletionMessage)
	err = ScheduleNamespaceDeletion(ctx, client)
	if err != nil {
		log.Printf(ScheduleNamespaceDeletionErrorMessage, err)
	}

	// Wait a time between strategies to allow Kubernetes try to manage the garbage
	log.Printf(WaitTimeMessage, *inputFlags.Duration)
	time.Sleep(*inputFlags.Duration)

	// 4. Delete unavailable API services
	log.Print(DeleteOrphanApisMessage)
	err = kubernetes.DeleteOrphanApiServices(ctx, client.Dynamic)
	if err != nil {
		log.Printf(DeleteOrphanApisErrorMessage, err)
	}

	// Wait a time between strategies to allow Kubernetes try to manage the garbage
	log.Printf(WaitTimeMessage, *inputFlags.Duration)
	time.Sleep(*inputFlags.Duration)

	// 5. Delete resources on stuck namespaces
	log.Print(CleanResourcesMessage)
	err = CleanStuckNamespaces(ctx, client)
	if err != nil {
		log.Printf(CleanResourcesErrorMessage, err)
	}

	// Wait a time between strategies to allow Kubernetes try to manage the garbage
	log.Printf(WaitTimeMessage, *inputFlags.Duration)
	time.Sleep(*inputFlags.Duration)

	// 6. DeleteNamespacesByForce
	log.Print(DeleteNamespaceByForceMessage)
	err = DeleteTerminatingNamespacesByForce(ctx, client)
	if err != nil {
		log.Print(err)
	}
}
