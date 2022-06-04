package main

import (
	"log"

	flagsPkg "nsmurder/flags"
	"nsmurder/operations"
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

	// Parse the flags from the command line
	flags.ParseFlags()

	// Generate the Kubernetes client to modify the resources
	log.Print(GetClientMessage)
	client, err := GetKubernetesClient(*flags.ConnectionMode, *flags.Kubeconfig)
	if err != nil {
		log.Printf(KubernetesGetClientErrorMessage, err)
	}

	// Get the namespace to terminate
	var namespaces []string

	namespaces = flags.GetNamespaces()
	if *flags.IncludeAll {
		namespaces, err = GetAllNamespaces(client)
	}

	if err != nil {
		log.Printf(GetNamespacesErrorMessage, err)
	}

	log.Print(namespaces)

	// Schedule deletion for desired namespaces
	err = operations.DeleteNamespaces(client, namespaces)
	if err != nil {
		log.Printf(DeleteNamespaceErrorMessage, err)
	}

	// TODO: Implement a time to wait between processes to let Kubernetes to manage the situation

	// Delete unavailable APIs

	// Delete stuck namespace's resources

	// Force delete stuck namespaces

	//// Update the Secrets time by time
	//for {
	//	// Build the Secret resources with the certificates content
	//	secrets, err := BuildSecrets(SecretNames, TLSHosts)
	//
	//	// Use the Kubernetes client to synchronize the resources
	//	log.Printf("Synchronizing the Secrets in the namespace: %s", *namespaceFlag)
	//	err = SynchronizeSecrets(client, *namespaceFlag, secrets)
	//	if err != nil {
	//		log.Printf("Error synchronizing the Secrets: %s", err)
	//	}
	//
	//	log.Printf("Next synchronization in %d seconds", SynchronizationScheduleSeconds)
	//	time.Sleep(SynchronizationScheduleSeconds * time.Second)
	//}
}
