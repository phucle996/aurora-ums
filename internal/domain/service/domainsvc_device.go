package domainsvc

import (
	"aurora/internal/domain/entity"
	"context"
	"time"
)

type DeviceSvcInterface interface {
	CreateDevice(ctx context.Context, device *entity.UserDevice) error
	CleanupStaleDevices(ctx context.Context, before time.Time) error
}
