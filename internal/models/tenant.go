package models

import (
	kamajiv1alpha1 "github.com/clastix/kamaji/api/v1alpha1"
	models "github.com/onekonsole/sys-service-provisioning/pkg/models"
)

type Tenant struct {
	TenantControlPlane kamajiv1alpha1.TenantControlPlane `json:"tenant_control_plane"`
	HostnameManager    models.HostnameManager            `json:"hostname_manager"`
}

func NewTenant(HostnameManager models.HostnameManager) *Tenant {
	return &Tenant{
		TenantControlPlane: kamajiv1alpha1.TenantControlPlane{},
		HostnameManager:    HostnameManager,
	}
}
