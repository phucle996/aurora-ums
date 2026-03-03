package domainsvc

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type MFASvcInterface interface {
	// TOTP (authenticator app)
	BeginTOTPSetup(ctx context.Context, userID uuid.UUID) (secret string, otpauthURL string, err error)
	VerifyAndEnableTOTP(ctx context.Context, userID uuid.UUID, code string) error
	VerifyTOTPCode(ctx context.Context, userID uuid.UUID, code string) error
	DisableTOTP(ctx context.Context, userID uuid.UUID) error

	// Methods
	ListEnabledMethods(ctx context.Context, userID uuid.UUID) ([]entity.MFAMethodType, error)

	// Recovery codes
	GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, count int) ([]string, error)
	ConsumeRecoveryCode(ctx context.Context, userID uuid.UUID, code string) error

	// Challenge handling (sms/email/app)
	IssueChallenge(ctx context.Context, userID uuid.UUID, method entity.MFAMethodType) (challengeID uuid.UUID, err error)
	VerifyChallenge(ctx context.Context, userID uuid.UUID, challengeID uuid.UUID, code string) error
}
