package svcImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	domainsvc "aurora/internal/domain/service"
	"context"
	"time"
)

type DeviceSvcImple struct {
	repo domainrepo.DeviceRepoInterface
}

func NewDeviceSvcImple(repo domainrepo.DeviceRepoInterface) domainsvc.DeviceSvcInterface {
	return &DeviceSvcImple{repo: repo}
}

func (s *DeviceSvcImple) CreateDevice(ctx context.Context, device *entity.UserDevice) error {
	return s.repo.CreateDevice(ctx, device)
}

func (s *DeviceSvcImple) CleanupStaleDevices(ctx context.Context, before time.Time) error {
	return s.repo.DeleteStaleDevices(ctx, before)
}
