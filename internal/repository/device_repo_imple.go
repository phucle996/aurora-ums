package repoImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepoImple struct {
	db *pgxpool.Pool
}

func NewDeviceRepoImple(db *pgxpool.Pool) domainrepo.DeviceRepoInterface {
	return &DeviceRepoImple{db: db}
}

func (r *DeviceRepoImple) CreateDevice(ctx context.Context, device *entity.UserDevice) error {

	m := model.UserDeviceEntityToModel(*device)
	const q = `
		INSERT INTO user_devices (
			id,
			user_id,
			device_id,
			device_secret_hash,
			user_agent,
			ip_first,
			ip_last,
			revoked,
			created_at,
			last_seen,
			revoked_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		)
	`
	_, err := r.db.Exec(ctx, q,
		m.ID,
		m.UserID,
		m.DeviceID,
		m.DeviceSecretHash,
		m.UserAgent,
		m.IPFirst,
		m.IPLast,
		m.Revoked,
		m.CreatedAt,
		m.LastSeen,
		m.RevokedAt,
	)
	return err
}

func (r *DeviceRepoImple) DeleteStaleDevices(ctx context.Context, before time.Time) error {
	const q = `
		DELETE FROM user_devices
		WHERE COALESCE(last_seen, created_at) < $1
	`
	_, err := r.db.Exec(ctx, q, before)
	return err
}

func (r *DeviceRepoImple) GetDeviceByDeviceID(ctx context.Context, deviceID uuid.UUID) (*entity.UserDevice, error) {
	const q = `
		SELECT
			id,
			user_id,
			device_id,
			device_secret_hash,
			user_agent,
			ip_first,
			ip_last,
			revoked,
			created_at,
			last_seen,
			revoked_at
		FROM user_devices
		WHERE device_id = $1
		LIMIT 1
	`
	var device model.UserDevice
	if err := r.db.QueryRow(ctx, q, deviceID).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceID,
		&device.DeviceSecretHash,
		&device.UserAgent,
		&device.IPFirst,
		&device.IPLast,
		&device.Revoked,
		&device.CreatedAt,
		&device.LastSeen,
		&device.RevokedAt,
	); err != nil {
		return nil, err
	}
	deviceEntity := model.UserDeviceModelToEntity(device)
	return &deviceEntity, nil
}
