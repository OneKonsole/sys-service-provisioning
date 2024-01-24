package repositories

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/onekonsole/sys-service-provisioning/internal/models"
	iRepository "github.com/onekonsole/sys-service-provisioning/internal/repositories/interfaces"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type tenantKubernetesCluster struct {
	clientset *kubernetes.Clientset
}

// NewTenantKubernetesCluster returns a new instance of the tenantKubernetesCluster struct
func NewTenantKubernetesCluster(clientset *kubernetes.Clientset) iRepository.TenantRepository {
	return &tenantKubernetesCluster{
		clientset: clientset,
	}
}

// CreateTenant creates the TenantControlPlane CRDS object on the Kubernetes cluster
func (t *tenantKubernetesCluster) CreateTenant(ctx context.Context, tenant models.Tenant) error {
	println("Creating TenantControlPlane CRDS object on the Kubernetes cluster...")
	ctx = context.Background()

	// Create the TenantControlPlane CRDS object on the Kubernetes cluster
	_, err := t.clientset.CoreV1().RESTClient().Post().
		AbsPath("/apis/kamaji.clastix.io/v1alpha1").
		Namespace(tenant.TenantControlPlane.Namespace).
		Resource("tenantcontrolplanes").
		Body(&tenant.TenantControlPlane).
		DoRaw(ctx)

	if err != nil {
		fmt.Printf("Error creating TenantControlPlane CRDS object on the Kubernetes cluster: %v", err)
		return err
	}

	return nil
}

// CreateTenantNamespace creates the namespace of the tenant on the Kubernetes cluster if it doesn't exist
func (t *tenantKubernetesCluster) CreateTenantNamespace(ctx context.Context, tenant models.Tenant) error {
	namespace := tenant.TenantControlPlane.Namespace
	_, err := t.clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		_, err = t.clientset.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}, metav1.CreateOptions{})
		if err != nil && err.Error() != fmt.Sprint("namespaces \""+namespace+"\" already exists") {
			return err
		}
	}
	return nil
}

// FindAvailableNodePort returns an available node port number
func (t *tenantKubernetesCluster) FindAvailableNodePort(ctx context.Context) (int32, error) {
	// Concurent-safe random number generator
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	ctx = context.Background()
	// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()
	for {
		// Generate a random port number in the NodePort range
		port := int32(generator.Intn(2768) + 30000)

		// Get a list of all services in all namespaces
		services, err := t.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error getting a list of all services in all namespaces: %v", err)
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
