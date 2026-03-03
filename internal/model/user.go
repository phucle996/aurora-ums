package model

import (
	"time"

	"aurora/internal/domain/entity"

	"github.com/google/uuid"
)

// UserDB mirrors users table with db tags for scanning.
type User struct {
	ID         uuid.UUID         `db:"id"`
	Username   string            `db:"username"`
	Email      string            `db:"email"`
	Password   string            `db:"password_hash"`
	Status     entity.UserStatus `db:"status"`
	UserLevel  int32             `db:"user_level"`
	OnBoarding bool              `db:"on_boarding"`
	CreatedAt  time.Time         `db:"created_at"`
	UpdatedAt  time.Time         `db:"updated_at"`
}

func UserModelToEntity(m User) entity.User {
	// model -> entity
	return entity.User{
		ID:         m.ID,
		Username:   m.Username,
		Email:      m.Email,
		Password:   m.Password,
		Status:     m.Status,
		UserLevel:  m.UserLevel,
		OnBoarding: m.OnBoarding,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func UserEntityToModel(e entity.User) User {
	return User{
		ID:         e.ID,
		Username:   e.Username,
		Email:      e.Email,
		Password:   e.Password,
		Status:     e.Status,
		UserLevel:  e.UserLevel,
		OnBoarding: e.OnBoarding,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}

// RefreshTokenDB mirrors refresh_tokens table.
type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	DeviceID  uuid.UUID `db:"device_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

func RefreshTokenModelToEntity(m RefreshToken) entity.RefreshToken {
	return entity.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		DeviceID:  m.DeviceID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}
}

func RefreshTokenEntityToModel(e entity.RefreshToken) RefreshToken {
	return RefreshToken{
		ID:        e.ID,
		UserID:    e.UserID,
		DeviceID:  e.DeviceID,
		TokenHash: e.TokenHash,
		ExpiresAt: e.ExpiresAt,
		CreatedAt: e.CreatedAt,
	}
}

// UserDeviceDB mirrors user_devices table.
type UserDevice struct {
	ID               uuid.UUID  `db:"id"`
	UserID           uuid.UUID  `db:"user_id"`
	DeviceID         uuid.UUID  `db:"device_id"`
	DeviceSecretHash string     `db:"device_secret_hash"`
	UserAgent        *string    `db:"user_agent"`
	IPFirst          *string    `db:"ip_first"`
	IPLast           *string    `db:"ip_last"`
	Revoked          bool       `db:"revoked"`
	CreatedAt        *time.Time `db:"created_at"`
	LastSeen         *time.Time `db:"last_seen"`
	RevokedAt        *time.Time `db:"revoked_at"`
}

func UserDeviceModelToEntity(m UserDevice) entity.UserDevice {
	return entity.UserDevice{
		ID:               m.ID,
		UserID:           m.UserID,
		DeviceID:         m.DeviceID,
		DeviceSecretHash: m.DeviceSecretHash,
		UserAgent:        m.UserAgent,
		IPFirst:          m.IPFirst,
		IPLast:           m.IPLast,
		Revoked:          m.Revoked,
		CreatedAt:        m.CreatedAt,
		LastSeen:         m.LastSeen,
		RevokedAt:        m.RevokedAt,
	}
}

func UserDeviceEntityToModel(e entity.UserDevice) UserDevice {
	return UserDevice{
		ID:               e.ID,
		UserID:           e.UserID,
		DeviceID:         e.DeviceID,
		DeviceSecretHash: e.DeviceSecretHash,
		UserAgent:        e.UserAgent,
		IPFirst:          e.IPFirst,
		IPLast:           e.IPLast,
		Revoked:          e.Revoked,
		CreatedAt:        e.CreatedAt,
		LastSeen:         e.LastSeen,
		RevokedAt:        e.RevokedAt,
	}
}

// Profile mirrors profiles table.
type Profile struct {
	ID             uuid.UUID  `db:"id"`
	UserID         uuid.UUID  `db:"user_id"`
	FullName       *string    `db:"full_name"`
	Company        *string    `db:"company"`
	ReferralSource *string    `db:"referral_source"`
	Phone          *string    `db:"phone"`
	JobFunction    *string    `db:"job_function"`
	Country        *string    `db:"country"`
	AvatarURL      *string    `db:"avatar_url"`
	Bio            *string    `db:"bio"`
	CreatedAt      *time.Time `db:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at"`
}

func ProfileModelToEntity(m Profile) entity.Profile {
	return entity.Profile{
		ID:             m.ID,
		UserID:         m.UserID,
		FullName:       m.FullName,
		Company:        m.Company,
		ReferralSource: m.ReferralSource,
		Phone:          m.Phone,
		JobFunction:    m.JobFunction,
		Country:        m.Country,
		AvatarURL:      m.AvatarURL,
		Bio:            m.Bio,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func ProfileEntityToModel(e entity.Profile) Profile {
	return Profile{
		ID:             e.ID,
		UserID:         e.UserID,
		FullName:       e.FullName,
		Company:        e.Company,
		ReferralSource: e.ReferralSource,
		Phone:          e.Phone,
		JobFunction:    e.JobFunction,
		Country:        e.Country,
		AvatarURL:      e.AvatarURL,
		Bio:            e.Bio,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
