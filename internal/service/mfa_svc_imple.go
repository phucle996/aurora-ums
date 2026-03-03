package svcImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	"aurora/internal/security"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MFASvcImple struct {
	repo domainrepo.MFARepoInterface
}

func NewMFASvcImple(repo domainrepo.MFARepoInterface) domainsvc.MFASvcInterface {
	return &MFASvcImple{repo: repo}
}

func (s *MFASvcImple) BeginTOTPSetup(ctx context.Context, userID uuid.UUID) (string, string, error) {
	existing, err := s.repo.GetMethodByUserAndType(ctx, userID, entity.MFAMethodTOTP)
	if err == nil {
		if existing.VerifiedAt != nil {
			return "", "", errorx.ErrMFAMethodAlreadyEnabled
		}
		if existing.Secret != nil && *existing.Secret != "" {
			return *existing.Secret, security.BuildOTPAuthURL(userID.String(), *existing.Secret), nil
		}
		_ = s.repo.DeleteMethod(ctx, userID, entity.MFAMethodTOTP)
	} else if !errors.Is(err, errorx.ErrMFAMethodNotFound) {
		return "", "", err
	}

	secret, err := security.GenerateTOTPSecret()
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	method := &entity.MFAMethod{
		ID:        uuid.New(),
		UserID:    userID,
		Method:    entity.MFAMethodTOTP,
		Secret:    &secret,
		CreatedAt: &now,
	}
	if err := s.repo.CreateMethod(ctx, method); err != nil {
		return "", "", err
	}

	return secret, security.BuildOTPAuthURL(userID.String(), secret), nil
}

func (s *MFASvcImple) VerifyAndEnableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	method, err := s.repo.GetMethodByUserAndType(ctx, userID, entity.MFAMethodTOTP)
	if err != nil {
		return err
	}
	if method.VerifiedAt != nil {
		return errorx.ErrMFAMethodAlreadyEnabled
	}
	if method.Secret == nil || *method.Secret == "" {
		return errorx.ErrTokenInvalid
	}
	if !security.ValidateTOTP(*method.Secret, code, 1) {
		return errorx.ErrMFACodeInvalid
	}
	return s.repo.UpdateMethodVerifiedAt(ctx, method.ID, time.Now())
}

func (s *MFASvcImple) VerifyTOTPCode(ctx context.Context, userID uuid.UUID, code string) error {
	method, err := s.repo.GetMethodByUserAndType(ctx, userID, entity.MFAMethodTOTP)
	if err != nil {
		return err
	}
	if method.Secret == nil || *method.Secret == "" {
		return errorx.ErrTokenInvalid
	}
	if method.VerifiedAt == nil {
		return errorx.ErrMFAMethodNotFound
	}
	if !security.ValidateTOTP(*method.Secret, code, 1) {
		return errorx.ErrMFACodeInvalid
	}
	return nil
}

func (s *MFASvcImple) DisableTOTP(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteMethod(ctx, userID, entity.MFAMethodTOTP)
}

func (s *MFASvcImple) ListEnabledMethods(ctx context.Context, userID uuid.UUID) ([]entity.MFAMethodType, error) {
	methods, err := s.repo.ListMethodsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]entity.MFAMethodType, 0, len(methods))
	for _, m := range methods {
		out = append(out, m.Method)
	}
	return out, nil
}

func (s *MFASvcImple) GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	codes := make([]string, 0, count)
	entities := make([]entity.MFARecoveryCode, 0, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		code, err := generateRecoveryCode()
		if err != nil {
			return nil, err
		}
		hash := hashRecoveryCode(code)
		codes = append(codes, code)
		entities = append(entities, entity.MFARecoveryCode{
			ID:        uuid.New(),
			UserID:    userID,
			CodeHash:  hash,
			CreatedAt: &now,
		})
	}

	if err := s.repo.DeleteRecoveryCodesByUser(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.repo.CreateRecoveryCodes(ctx, entities); err != nil {
		return nil, err
	}

	return codes, nil
}

func (s *MFASvcImple) ConsumeRecoveryCode(ctx context.Context, userID uuid.UUID, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return errorx.ErrMFACodeInvalid
	}
	hash := hashRecoveryCode(code)

	codes, err := s.repo.ListRecoveryCodesByUser(ctx, userID)
	if err != nil {
		return err
	}

	found := false
	for _, c := range codes {
		if c.CodeHash == hash {
			found = true
			continue
		}
	}
	if !found {
		return errorx.ErrMFACodeInvalid
	}

	if err := s.repo.DeleteRecoveryCodesByUser(ctx, userID); err != nil {
		return err
	}

	remaining := make([]entity.MFARecoveryCode, 0, len(codes)-1)
	now := time.Now()
	for _, c := range codes {
		if c.CodeHash == hash {
			continue
		}
		remaining = append(remaining, entity.MFARecoveryCode{
			ID:        c.ID,
			UserID:    c.UserID,
			CodeHash:  c.CodeHash,
			CreatedAt: &now,
		})
	}
	if len(remaining) > 0 {
		if err := s.repo.CreateRecoveryCodes(ctx, remaining); err != nil {
			return err
		}
	}
	return nil
}

func (s *MFASvcImple) IssueChallenge(ctx context.Context, userID uuid.UUID, method entity.MFAMethodType) (uuid.UUID, error) {
	return uuid.Nil, errors.New("not implemented")
}

func (s *MFASvcImple) VerifyChallenge(ctx context.Context, userID uuid.UUID, challengeID uuid.UUID, code string) error {
	return errors.New("not implemented")
}

func generateRecoveryCode() (string, error) {
	const size = 10
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	code := enc.EncodeToString(buf)
	if len(code) > size {
		code = code[:size]
	}
	return code, nil
}

func hashRecoveryCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
