package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	kamajiv1alpha1 "github.com/clastix/kamaji/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Tenant struct {
	TenantControlPlane kamajiv1alpha1.TenantControlPlane `json:"tenant_control_plane"`
	HostnameManager    HostnameManager                   `json:"hostname_manager"`
}

func NewTenant(HostnameManager HostnameManager) *Tenant {
	return &Tenant{
		TenantControlPlane: kamajiv1alpha1.TenantControlPlane{},
		HostnameManager:    HostnameManager,
	}
}

// CreateTenant => Create a tenant requested by an order on the specified Kubernetes cluster
func (tenant Tenant) CreateTenant(ctx context.Context, order Order, client kubernetes.Clientset, namespace string, datastore string) error {
	// Convert UserID and OrderID to string
	userID := strconv.Itoa(order.UserID)
	orderID := strconv.Itoa(order.ID)

	// Get the client kubernetes cluster version
	// version, err := discovery.NewDiscoveryClientForConfigOrDie(client.RESTClient().Config()).ServerVersion()
	// if err != nil {
	// 	return err
	// }
	// println(version.String())

	version := "v1.28.2"

	labels := map[string]string{
		"tenant.clastix.io": order.ClusterName,
		"app":               "tenant-control-plane",
		"client":            userID,
		"order":             orderID,
	}

	annotations := map[string]string{}

	additionalMetadata := kamajiv1alpha1.AdditionalMetadata{
		Labels:      labels,
		Annotations: annotations,
	}

	// Create a metadata object for the tenant
	meta := metav1.ObjectMeta{
		Name:      order.ClusterName,
		Labels:    labels,
		Namespace: namespace,
	}

	replicas := int32(1)

	// Control plane deployment specifications
	controlPlaneComponentsResources := kamajiv1alpha1.ControlPlaneComponentsResources{
		APIServer: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
			Limits: corev1.ResourceList{},
		},
		ControllerManager: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("125m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			Limits: corev1.ResourceList{},
		},
		Scheduler: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("125m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			Limits: corev1.ResourceList{},
		},
	}

	controlPlaneDeploymentSpec := kamajiv1alpha1.DeploymentSpec{
		Replicas:           &replicas,
		AdditionalMetadata: additionalMetadata,
		Resources:          &controlPlaneComponentsResources,
	}

	controlPlaneService := kamajiv1alpha1.ServiceSpec{
		AdditionalMetadata: additionalMetadata,
		ServiceType:        kamajiv1alpha1.ServiceTypeNodePort,
	}

	controlPlaneIngress := kamajiv1alpha1.IngressSpec{
		AdditionalMetadata: additionalMetadata,
		IngressClassName:   "nginx",
		Hostname:           tenant.HostnameManager.FullSubdomain,
	}

	controlPlane := kamajiv1alpha1.ControlPlane{
		Deployment: controlPlaneDeploymentSpec,
		Service:    controlPlaneService,
		Ingress:    &controlPlaneIngress,
	}

	// Kubernetes cluster specifications
	kubernetesClusterSpec := kamajiv1alpha1.KubernetesSpec{
		Version: version,
		Kubelet: kamajiv1alpha1.KubeletSpec{
			CGroupFS: "systemd",
		},
		AdmissionControllers: []kamajiv1alpha1.AdmissionController{
			"ResourceQuota",
			"LimitRanger",
		},
	}

	// Network profile specifications
	networkProfileSpec := kamajiv1alpha1.NetworkProfileSpec{
		Address: "127.0.0.1",
		Port:    int32(32536),
		CertSANs: []string{
			tenant.HostnameManager.FullSubdomain,
		},
		ServiceCIDR: "10.96.0.0/16",
		PodCIDR:     "10.244.0.0/16",
		DNSServiceIPs: []string{
			"10.96.0.10",
		},
	}

	// Konnectivity specifications
	konnectivitySpec := kamajiv1alpha1.KonnectivitySpec{
		KonnectivityServerSpec: kamajiv1alpha1.KonnectivityServerSpec{
			Port: int32(8132),
			Resources: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{},
			},
		},
		KonnectivityAgentSpec: kamajiv1alpha1.KonnectivityAgentSpec{},
	}

	// Addons specifications
	addonsSpec := kamajiv1alpha1.AddonsSpec{
		CoreDNS:      &kamajiv1alpha1.AddonSpec{},
		KubeProxy:    &kamajiv1alpha1.AddonSpec{},
		Konnectivity: &konnectivitySpec,
	}

	// Tenant control plane specifications
	tenantControlPlaneSpec := kamajiv1alpha1.TenantControlPlaneSpec{
		DataStore:      datastore,
		ControlPlane:   controlPlane,
		Kubernetes:     kubernetesClusterSpec,
		NetworkProfile: networkProfileSpec,
		Addons:         addonsSpec,
	}

	// Create a tenant control plane object with the order's specifications
	tenant.TenantControlPlane = kamajiv1alpha1.TenantControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TenantControlPlane",
			APIVersion: "kamaji.clastix.io/v1alpha1",
		},
		ObjectMeta: meta,
		Spec:       tenantControlPlaneSpec,
		Status:     kamajiv1alpha1.TenantControlPlaneStatus{},
	}

	// Display the TenantControlPlane CRDS object in JSON format
	tenantControlPlaneJSON, err := json.MarshalIndent(tenant.TenantControlPlane, "", "    ")
	if err != nil {
		fmt.Printf("Error displaying the TenantControlPlane CRDS object in JSON format: %v", err)
		return err
	}
	fmt.Printf("TenantControlPlane CRDS object in JSON format: %v", string(tenantControlPlaneJSON))

	//TODO: Uncomment when we will know what is the path to the TenantControlPlane CRDS object
	// Create the TenantControlPlane CRDS object on the Kubernetes cluster
	_, err = client.CoreV1().RESTClient().Post().
		AbsPath("/apis/kamaji.clastix.io/v1alpha1").
		Namespace(namespace).
		Resource("tenantcontrolplanes").
		Body(&tenant.TenantControlPlane).
		DoRaw(ctx)

	if err != nil {
		fmt.Printf("Error creating TenantControlPlane CRDS object on the Kubernetes cluster: %v", err)
		return err
	}

	//fmt.Printf("TenantControlPlane CRDS object created on the Kubernetes cluster: %v", tenant.TenantControlPlane)
	return nil
}

// GetTenants => Get all client's tenant requested by a client name on the specified Kubernetes cluster
