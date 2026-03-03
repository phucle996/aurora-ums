package entity

import "github.com/google/uuid"

type LoginWithPasswd struct {
	AccessToken          string
	RefreshToken         string
	UserState            UserStatus
	UserID               uuid.UUID
	DeviceID             uuid.UUID
	DeviceSecret         string
	MFARequired          bool
	MFAMethods           []MFAMethodType
	MFASession           string
	MFASessionTTLSeconds int
}

type RefreshSession struct {
	RefreshToken RefreshToken
	UserID       uuid.UUID
	DeviceID     uuid.UUID
	Status       UserStatus
	Username     string
}
