package operations

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeleteNamespaces schedule namespaces for deletion
func DeleteNamespaces(client *kubernetes.Clientset, namespaces []string) (err error) {

	// Zero grace period
	gracePeriod := int64(0)

	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	for _, namespace := range namespaces {
		err = client.CoreV1().Namespaces().Delete(context.TODO(), namespace, deleteOptions)
		if err != nil {
			break
		}
	}

	return err
}
