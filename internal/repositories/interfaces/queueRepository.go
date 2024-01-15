package interfaces

import (
	"context"

	"github.com/onekonsole/sys-service-provisioning/pkg/models"
)

type QueueRepository interface {
	Dequeue(ctx context.Context) (models.Order, error)
	Enqueue(ctx context.Context, order models.Order) error
}
