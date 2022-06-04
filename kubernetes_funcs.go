package main

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetAllNamespaces get a list of all namespaces existing in the cluster
func GetAllNamespaces(client *kubernetes.Clientset) (namespaces []string, err error) {

	namespaceList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return namespaces, err
	}

	for _, value := range namespaceList.Items {
		namespaces = append(namespaces, value.GetName())
	}

	return namespaces, err
}
