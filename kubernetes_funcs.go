package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesClient Return a Kubernetes client configured to connect from inside or outside the cluster
func GetKubernetesClient(connectionMode string, kubeconfigPath string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var client *kubernetes.Clientset

	// Create configuration to connect from inside the cluster using Kubernetes mechanisms
	config, err := rest.InClusterConfig()

	// Create configuration to connect from outside the cluster, using kubectl
	if connectionMode == "kubectl" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	// Check configuration errors in both cases
	if err != nil {
		return client, err
	}

	// Construct the client
	client, err = kubernetes.NewForConfig(config)
	return client, err
}
