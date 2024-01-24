package models

import "fmt"

// HostnameManager represents a struct for managing hostnames
type HostnameManager struct {
	Domain      string
	ClientName  string
	ClusterName string
	FullDomain  string // FullDomain is the full domain name of the cluster
}

// NewHostnameManager is a constructor function for HostnameManager
func NewHostnameManager(domain, clientName, clusterName string) *HostnameManager {
	fullDomain := fmt.Sprintf("%s.%s.%s", clusterName, clientName, domain)

	return &HostnameManager{
		Domain:     domain,
		ClientName: clientName,
		FullDomain: fullDomain,
	}
}
