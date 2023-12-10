package models

import "fmt"

// HostnameManager represents a struct for managing hostnames
type HostnameManager struct {
	Domain        string
	ClientName    string
	Subdomain     string
	FullDomain    string // Combination of Subdomain and Domain
	FullSubdomain string // Combination of ClientName, Subdomain, and Domain
}

// NewHostnameManager is a constructor function for HostnameManager
func NewHostnameManager(domain, clientName, subdomain string) *HostnameManager {
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)
	fullSubdomain := fmt.Sprintf("%s.%s.%s", clientName, subdomain, domain)

	return &HostnameManager{
		Domain:        domain,
		ClientName:    clientName,
		Subdomain:     subdomain,
		FullDomain:    fullDomain,
		FullSubdomain: fullSubdomain,
	}
}
