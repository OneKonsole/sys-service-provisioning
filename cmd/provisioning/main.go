package provisioning

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/onekonsole/sys-service-provisioning/pkg/models"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Run(order models.Order) bool {

	// Create a new order
	// order := models.Order{
	// 	ID:              1,
	// 	UserID:          1,
	// 	ClusterName:     "test",
	// 	HasControlPlane: true,
	// 	HasMonitoring:   true,
	// 	HasAlerting:     true,
	// 	StorageSize:     10,
	// }

	//TODO: Get information from the cli
	// Create a new hostname manager
	hostnameManager := models.NewHostnameManager("emetral.fr", order.ClusterName, "onekonsole")

	// Create a new tenant
	tenant := models.NewTenant(*hostnameManager)

	// Create a new Kubernetes client
	client, err := GetKubernetesClientset()
	if err != nil {
		return false
	}

	//Check if the namespace exists and create it if it doesn't
	namespace := strconv.Itoa(order.UserID)
	_, err = client.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		_, err = client.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil && err.Error() != fmt.Sprint("namespaces \""+namespace+"\" already exists") {
			return false
		}
	}

	// Create a new tenant
	err = tenant.CreateTenant(context.Background(), order, *client, namespace, "default")
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

// GetKubernetesClientset returns a Kubernetes clientset using the kubeconfig file at the default location.
func GetKubernetesClientset() (*kubernetes.Clientset, error) {
	// Get the kubeconfig file path from the default location
	kubeconfig := os.Getenv("HOME") + "/.kube/config"

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	return clientset, nil
}
