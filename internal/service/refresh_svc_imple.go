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
)

type RefreshSvcImple struct {
	RefreshRepo domainrepo.RefreshRepoInterface
	Token       *config.TokenCfg
}

func NewRefreshSvcImple(RefreshRepo domainrepo.RefreshRepoInterface, Token *config.TokenCfg) domainsvc.RefreshSvcInterface {
	return &RefreshSvcImple{
		RefreshRepo: RefreshRepo,
		Token:       Token,
	}

}

func (s *RefreshSvcImple) Create(ctx context.Context, refresh *entity.RefreshToken, plainToken string) error {
	//hash token
	hashToken, err := security.HashToken(plainToken, s.Token.GetRefreshSecret())
	if err != nil {
		return err
	}
	refresh.TokenHash = hashToken
	refresh.ExpiresAt = time.Now().Add(s.Token.RefreshTTL)
	refresh.CreatedAt = time.Now()
	//call svc
	if err := s.RefreshRepo.Create(ctx, refresh); err != nil {
		return err
	}
	return nil
}

func (s *RefreshSvcImple) GetRefreshTokenByDevice(ctx context.Context, refresh string, deviceID uuid.UUID, deviceSecretHash string) (*entity.RefreshSession, error) {
	hashToken, err := security.HashToken(refresh, s.Token.GetRefreshSecret())
	if err != nil {
		return nil, err
	}

	return s.RefreshRepo.GetTokenByDevice(ctx, hashToken, deviceID, deviceSecretHash)
}
