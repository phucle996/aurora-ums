package svcImple

import (
	"aurora/internal/config"
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/security"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OttSvcImple struct {
	Token   *config.TokenCfg
	OttRepo domainrepo.OttRepoInterface
}

func NewOttSvcImple(Token *config.TokenCfg, OttRepo domainrepo.OttRepoInterface) domainsvc.OttSvcInterface {
	return &OttSvcImple{
		Token:   Token,
		OttRepo: OttRepo,
	}
}

func (s *OttSvcImple) CreateToken(ctx context.Context, userID uuid.UUID, tokenType entity.OttType) (string, error) {

	// gen token
	plainToken, err := security.GenerateToken(128)
	if err != nil {
		return "", err
	}
	// hash token
	hashToken, err := security.HashToken(plainToken, s.Token.OttSecret)
	if err != nil {
		return "", err
	}

	// init entity
	ott := &entity.OneTimeToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashToken,
		Purpose:   tokenType,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.Token.OttTTL),
	}

	if err := s.OttRepo.Create(ctx, ott); err != nil {
		return "", err
	}
	return plainToken, nil
}

func (s *OttSvcImple) ValidateToken(ctx context.Context, userID uuid.UUID, plainToken string, tokenType entity.OttType) error {
	tokenHash, err := security.HashToken(plainToken, s.Token.OttSecret)
	if err != nil {
		return err
	}
	ott := &entity.OneTimeToken{
		UserID:    userID,
		TokenHash: tokenHash,
		Purpose:   tokenType,
	}
	return s.OttRepo.Validate(ctx, ott)
}

func (s *OttSvcImple) ConsumTokenTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, plainToken string, tokenType entity.OttType) error {
	tokenHash, err := security.HashToken(plainToken, s.Token.OttSecret)
	if err != nil {
		return err
	}
	ott := &entity.OneTimeToken{
		UserID:    userID,
		TokenHash: tokenHash,
		Purpose:   tokenType,
	}
	return s.OttRepo.ConsumTx(ctx, tx, ott)
}
