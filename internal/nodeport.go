package internal

import (
	"context"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetAvailableNodePort returns an available node port number
func GetAvailableNodePort(clientset *kubernetes.Clientset) (int32, error) {
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		// Generate a random port number in the NodePort range
		port := int32(rand.Intn(2768) + 30000)

		// Get a list of all services in all namespaces
		services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return 0, err
		}

		// Check if the generated port number is used by any service
		portUsed := false
		for _, service := range services.Items {
			for _, servicePort := range service.Spec.Ports {
				if servicePort.NodePort == port {
					portUsed = true
					break
				}
			}
			if portUsed {
				break
			}
		}

		// If the port number is not used, return it
		if !portUsed {
			return port, nil
		}
	}
}
