package interfaces

import (
	"context"

	"github.com/onekonsole/sys-service-provisioning/pkg/models"
)

type Tenant interface {
	CreateTenant(ctx context.Context, order models.Order, namespace string, datastore string) error
}
