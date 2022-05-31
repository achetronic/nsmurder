package main

import (
	"flag"
	"log"
	"path/filepath"
	//"time"

	"k8s.io/client-go/util/homedir"
)

const (
	// Messages to help the users when using flags
	ConnectionModeFlagDescription = "(optional) What type of connection to use: incluster, kubectl"
	KubeconfigFlagDescription     = "(optional) absolute path to the kubeconfig file"
	IncludeAllFlagDescription     = "Schedule deletion for all namespaces"
	IncludeFlagDescription        = "Namespaces to include in deletion list"
	IgnoreFlagDescription         = "Namespaces to ignore from deletion list"

	// SynchronizationScheduleSeconds The time in seconds between synchronizations
	SynchronizationScheduleSeconds = 20
)

type Flags struct {
	connectionMode *string    `json:"connection_mode"`
	kubeconfig     *string    `json:"kubeconfig"`
	includeAll     *string    `json:"include_all"`
	include        arrayFlags `json:"include"`
	ignore         arrayFlags `json:"ignore"`
}

var flags Flags

//func SynchronizeSecrets(client *kubernetes.Clientset, namespace string, secrets []Secret) error {
//	var err error
//	secretsClient := client.CoreV1().Secrets(namespace)
//
//	// Synchronize the Secret resources
//	for _, secret := range secrets {
//
//		// Generate a Secret structure
//		secretObject := &corev1.Secret{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      secret.Name,
//				Namespace: namespace,
//			},
//			StringData: map[string]string{
//				"tls.crt": secret.Certificate,
//			},
//		}
//
//		// Search for the secret
//		_, err = secretsClient.Get(context.Background(), secret.Name, metav1.GetOptions{})
//
//		// The Secret does NOT exist: Create it
//		if err != nil {
//			_, err = secretsClient.Create(context.Background(), secretObject.DeepCopy(), metav1.CreateOptions{})
//			if err != nil {
//				return err
//			}
//		}
//
//		// The Secret DOES exist: Update it
//		_, err = secretsClient.Update(context.Background(), secretObject, metav1.UpdateOptions{})
//		if err != nil {
//			return err
//		}
//	}
//
//	return err
//}

// ParseFlags parse the flags introduced by the user from the command line
func ParseFlags(flags *Flags) {
	// Get the values from flags
	flags.connectionMode = flag.String("connection-mode", "kubectl", ConnectionModeFlagDescription)
	flags.kubeconfig = flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), KubeconfigFlagDescription)
	flags.includeAll = flag.String("include-all", "false", IncludeAllFlagDescription)
	flag.Var(&flags.include, "include", IncludeFlagDescription)
	flag.Var(&flags.ignore, "ignore", IgnoreFlagDescription)
	flag.Parse()
}

// GetNamespacesToKill return a list with the desired namespaces to be killed already filtered
func GetNamespacesFromFlags(flags *Flags) []string {
	//
}

func main() {

	// Parse the flags from the command line
	ParseFlags(&flags)

	// Generate the Kubernetes client to modify the resources
	log.Printf("Generating the client to connect to Kubernetes")
	client, err := GetKubernetesClient(*flags.connectionMode, *flags.kubeconfig)
	if err != nil {
		log.Printf("Error connecting to Kubernetes API: %s", err)
	}

	// GET SCHEDULABLE NAMESPACES FROM FLAGS
	namespacesToKill := GetNamespacesFromFlags(&flags)

	// Schedule namespaces for deletion

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
