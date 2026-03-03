package domainrepo

import (
	"aurora/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceRepoInterface interface {
	CreateDevice(ctx context.Context, device *entity.UserDevice) error
	DeleteStaleDevices(ctx context.Context, before time.Time) error
	GetDeviceByDeviceID(ctx context.Context, deviceID uuid.UUID) (*entity.UserDevice, error)
}
