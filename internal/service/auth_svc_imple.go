package svcImple

import (
	redisinfra "aurora/infra/redis"
	"aurora/internal/cache"
	"aurora/internal/config"
	"aurora/internal/domain/entity"
	appctxKey "aurora/internal/domain/key"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	"aurora/internal/security"
	"context"
	"crypto/subtle"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthSvcImple struct {
	UserSvc     domainsvc.UserSvcInterface
	OttSvc      domainsvc.OttSvcInterface
	RefreshSvc  domainsvc.RefreshSvcInterface
	DeviceSvc   domainsvc.DeviceSvcInterface
	RbacSvc     domainsvc.RBACSvcInterface
	MfaSvc      domainsvc.MFASvcInterface
	Token       *config.TokenCfg
	Publisher   redisinfra.EventPublisher
	Blacklist   *cache.JWTBlacklist
	DeviceCache *cache.DeviceSecretCache
	PermCache   *cache.UserPermissionCache
	MFASession  *cache.MFASessionCache
}

func NewAuthSvcImple(
	UserSvc domainsvc.UserSvcInterface,
	OttSvc domainsvc.OttSvcInterface,
	RefreshSvc domainsvc.RefreshSvcInterface,
	DeviceSvc domainsvc.DeviceSvcInterface,
	RbacSvc domainsvc.RBACSvcInterface,
	MfaSvc domainsvc.MFASvcInterface,
	Publisher redisinfra.EventPublisher,
	Token *config.TokenCfg,
	blacklist *cache.JWTBlacklist,
	deviceCache *cache.DeviceSecretCache,
	permCache *cache.UserPermissionCache,
	mfaSession *cache.MFASessionCache,
) domainsvc.AuthSvcInterface {
	return &AuthSvcImple{
		UserSvc:     UserSvc,
		OttSvc:      OttSvc,
		RefreshSvc:  RefreshSvc,
		DeviceSvc:   DeviceSvc,
		RbacSvc:     RbacSvc,
		MfaSvc:      MfaSvc,
		Publisher:   Publisher,
		Token:       Token,
		Blacklist:   blacklist,
		DeviceCache: deviceCache,
		PermCache:   permCache,
		MFASession:  mfaSession,
	}
}

func (s *AuthSvcImple) RegisterAccount(ctx context.Context, user *entity.User) error {
	user.ID = uuid.New()
	user.Status = entity.UserStatusPending
	user.UserLevel = 4
	user.OnBoarding = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.UserSvc.CreateUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (s *AuthSvcImple) buildAccessToken(ctx context.Context, user entity.User, deviceID uuid.UUID) (string, []string, error) {
	// get all role
	roleNames := make([]string, 0)
	roles, err := s.RbacSvc.ListUserRoles(ctx, user.ID)
	if err != nil && !errors.Is(err, errorx.ErrUserRoleNotFound) {
		return "", nil, err
	}

	roleNames = make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	permNames, err := s.collectPermissionNamesByRoles(ctx, roles)
	if err != nil {
		return "", nil, err
	}
	if err := s.PermCache.Set(ctx, user.ID, permNames, 60*time.Minute); err != nil {
		return "", nil, err
	}

	claims := map[string]any{
		"username":   user.Username,
		"roles":      roleNames,
		"device_id":  deviceID.String(),
		"jti":        uuid.NewString(),
		"user_level": user.UserLevel,
	}

	accessToken, err := security.GenerateJWT(
		user.ID.String(),
		s.Token.GetAccessSecret(),
		s.Token.AccessTTL,
		claims,
	)
	if err != nil {
		return "", nil, err
	}
	return accessToken, roleNames, nil
}

func (s *AuthSvcImple) collectPermissionNamesByRoles(ctx context.Context, roles []entity.Role) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}

	permSet := make(map[string]struct{})
	for _, role := range roles {
		perms, err := s.RbacSvc.ListRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, perm := range perms {
			name := strings.TrimSpace(perm.Name)
			if name == "" {
				continue
			}
			permSet[name] = struct{}{}
		}
	}

	permNames := make([]string, 0, len(permSet))
	for name := range permSet {
		permNames = append(permNames, name)
	}
	sort.Strings(permNames)
	return permNames, nil
}

func (s *AuthSvcImple) issueSession(ctx context.Context, user entity.User) (*entity.LoginWithPasswd, error) {
	deviceID := uuid.New()
	deviceSecret, err := security.GenerateToken(32)
	if err != nil {
		return nil, err
	}
	deviceSecretHash, err := security.HashToken(deviceSecret, s.Token.GetDeviceSecret())
	if err != nil {
		return nil, err
	}

	accessToken, _, err := s.buildAccessToken(ctx, user, deviceID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := security.GenerateToken(128)
	if err != nil {
		return nil, err
	}

	ref := &entity.RefreshToken{
		ID:       uuid.New(),
		UserID:   user.ID,
		DeviceID: deviceID,
	}
	if err := s.RefreshSvc.Create(ctx, ref, refreshToken); err != nil {
		return nil, err
	}

	now := time.Now()
	device := &entity.UserDevice{
		ID:               uuid.New(),
		UserID:           user.ID,
		DeviceID:         deviceID,
		DeviceSecretHash: deviceSecretHash,
		Revoked:          false,
		CreatedAt:        &now,
		LastSeen:         &now,
	}
	if err := s.DeviceSvc.CreateDevice(ctx, device); err != nil {
		return nil, err
	}

	if err := s.DeviceCache.Set(ctx, deviceID.String(), deviceSecretHash, s.Token.RefreshTTL); err != nil {
		return nil, err
	}

	return &entity.LoginWithPasswd{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserState:    user.Status,
		UserID:       user.ID,
		DeviceID:     deviceID,
		DeviceSecret: deviceSecret,
	}, nil
}

func (s *AuthSvcImple) Login(ctx context.Context, username, passwd string) (*entity.LoginWithPasswd, error) {

	user, err := s.UserSvc.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := security.ComparePassword(user.Password, passwd); err != nil {
		return nil, err
	}

	if user.Status == entity.UserStatusPending {
		plainToken, err := s.OttSvc.CreateToken(ctx, user.ID, entity.OttAccountVerify)
		if err != nil {
			return nil, err
		}
		_ = s.Publisher.Publish(ctx, "auth.account_verify", map[string]any{
			"user_id": user.ID.String(),
			"email":   user.Email,
			"token":   plainToken,
		})

		return &entity.LoginWithPasswd{UserState: entity.UserStatusPending}, nil
	}
	if user.Status == entity.UserStatusSuspended {
		return &entity.LoginWithPasswd{UserState: entity.UserStatusSuspended}, nil
	}
	if user.Status == entity.UserStatusDeleted {
		return &entity.LoginWithPasswd{UserState: entity.UserStatusDeleted}, nil
	}

	methods, err := s.MfaSvc.ListEnabledMethods(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if len(methods) > 0 {
		if s.MFASession == nil {
			return nil, errors.New("mfa session cache is nil")
		}
		mfaTTL := 150 * time.Second
		mfaToken, err := security.GenerateToken(32)
		if err != nil {
			return nil, err
		}
		if err := s.MFASession.Set(ctx, user.ID.String(), mfaToken, mfaTTL); err != nil {
			return nil, err
		}
		return &entity.LoginWithPasswd{
			UserState:            entity.UserStatusActive,
			UserID:               user.ID,
			MFARequired:          true,
			MFAMethods:           methods,
			MFASession:           mfaToken,
			MFASessionTTLSeconds: int(mfaTTL.Seconds()),
		}, nil

	}

	return s.issueSession(ctx, *user)
}

func (s *AuthSvcImple) VerifyMFAChallenge(ctx context.Context, userID uuid.UUID, sessionToken string,
	method entity.MFAMethodType, code string) (*entity.LoginWithPasswd, error) {

	if userID == uuid.Nil || strings.TrimSpace(sessionToken) == "" || strings.TrimSpace(code) == "" {
		return nil, errorx.ErrInvalidArgument
	}
	if s.MFASession == nil {
		return nil, errors.New("mfa session cache is nil")
	}

	cachedToken, err := s.MFASession.Get(ctx, userID.String())
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	if subtle.ConstantTimeCompare([]byte(cachedToken), []byte(sessionToken)) != 1 {
		return nil, errorx.ErrTokenInvalid
	}

	switch entity.MFAMethodType(strings.ToLower(strings.TrimSpace(string(method)))) {
	case entity.MFAMethodTOTP, entity.MFAMethodType("authenticator"):
		if err := s.MfaSvc.VerifyTOTPCode(ctx, userID, code); err != nil {
			return nil, err
		}
	default:
		return nil, errorx.ErrInvalidArgument
	}

	_ = s.MFASession.Delete(ctx, userID.String())

	userWithProfile, err := s.UserSvc.GetUserWithProfileByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userWithProfile == nil {
		return nil, errorx.ErrUserNotFound
	}

	switch userWithProfile.User.Status {
	case entity.UserStatusPending:
		return nil, errorx.ErrUserIsPending
	case entity.UserStatusSuspended:
		return nil, errorx.ErrUserIsSuspended
	case entity.UserStatusDeleted:
		return nil, errorx.ErrUserIsDeleted
	case entity.UserStatusActive:
		return s.issueSession(ctx, userWithProfile.User)
	default:
		return nil, errorx.ErrTokenInvalid
	}
}

func (s *AuthSvcImple) Refresh(ctx context.Context, refresh string) (*entity.LoginWithPasswd, error) {
	if refresh == "" {
		return nil, errorx.ErrTokenInvalid
	}

	deviceIDRaw, ok := ctx.Value(appctxKey.KeyDeviceID).(string)
	if !ok || strings.TrimSpace(deviceIDRaw) == "" {
		return nil, errorx.ErrTokenInvalid
	}
	deviceSecret, ok := ctx.Value(appctxKey.KeyDeviceSecret).(string)
	if !ok || strings.TrimSpace(deviceSecret) == "" {
		return nil, errorx.ErrTokenInvalid
	}

	deviceID, err := uuid.Parse(strings.TrimSpace(deviceIDRaw))
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	deviceSecretHash, err := security.HashToken(deviceSecret, s.Token.GetDeviceSecret())
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}

	session, err := s.RefreshSvc.GetRefreshTokenByDevice(ctx, refresh, deviceID, deviceSecretHash)
	if err != nil {
		return nil, err
	}

	switch session.Status {
	case entity.UserStatusPending:
		return nil, errorx.ErrUserIsPending
	case entity.UserStatusSuspended:
		return nil, errorx.ErrUserIsSuspended
	case entity.UserStatusDeleted:
		return nil, errorx.ErrUserIsDeleted
	case entity.UserStatusActive:
		// continue
	default:
		return nil, errorx.ErrTokenInvalid
	}

	accessToken, _, err := s.buildAccessToken(ctx, entity.User{
		ID:       session.UserID,
		Username: session.Username,
	}, deviceID)
	if err != nil {
		return nil, err
	}

	return &entity.LoginWithPasswd{
		AccessToken:  accessToken,
		RefreshToken: refresh,
		UserState:    session.Status,
	}, nil
}

func (s *AuthSvcImple) Logout(ctx context.Context) error {
	if s.Blacklist == nil {
		return nil
	}
	jti, ok := ctx.Value(appctxKey.KeyJWTID).(string)
	if !ok || jti == "" {
		return nil
	}

	ttl := s.Token.AccessTTL
	if exp, ok := ctx.Value(appctxKey.KeyJWTExp).(int64); ok && exp > 0 {
		expTime := time.Unix(exp, 0)
		if expTime.After(time.Now()) {
			ttl = time.Until(expTime)
		} else {
			ttl = time.Second
		}
	}

	return s.Blacklist.Block(ctx, jti, ttl)
}
