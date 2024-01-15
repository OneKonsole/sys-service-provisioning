package interfaces

import (
	"context"

	"github.com/onekonsole/sys-service-provisioning/internal/models"
)

type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant models.Tenant) error
	FindAvailableNodePort(ctx context.Context) (int32, error)
	CreateTenantNamespace(ctx context.Context, tenant models.Tenant) error
}
