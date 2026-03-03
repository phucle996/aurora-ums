package handler

import (
	"aurora/internal/config"
	"aurora/internal/domain/entity"
	appctxKey "aurora/internal/domain/key"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	reqdto "aurora/internal/transport/http/handler/dto/request"
	"aurora/internal/transport/http/response"
	"aurora/pkg/logger"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	AuthSvc domainsvc.AuthSvcInterface
	Token   *config.TokenCfg
}

// contructor
func NewAuthHandler(AuthSvc domainsvc.AuthSvcInterface, Token *config.TokenCfg) *AuthHandler {
	return &AuthHandler{
		AuthSvc: AuthSvc,
		Token:   Token,
	}
}

// register
func (h *AuthHandler) SignupAccount(c *gin.Context) {
	op := "auth.SignupAccount"

	// context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// parse and validate dto iput
	var req *reqdto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	if req.Password != req.RePassword {
		response.RespondBadRequest(c, "passwords do not match")
		logger.HandlerInfo(ctx, op, "passwords do not match")
		return
	}

	if msg, logMsg := validatePassword(req.Password); msg != "" {
		response.RespondBadRequest(c, msg)
		logger.HandlerInfo(ctx, op, "%s", logMsg)
		return
	}

	// init entity from reqdto

	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// call service
	if err := h.AuthSvc.RegisterAccount(ctx, user); err != nil {
		switch {
		case errors.Is(err, errorx.ErrUserAlreadyExist):
			response.RespondConflict(c, "User Already Exist, please use another Username or Email ")
			logger.HandlerInfo(ctx, op, "User Already Exist")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}

	}
	response.RespondCreated(c, nil, "SignUp Account Successfully")
	logger.HandlerInfo(ctx, op, "SignUp Account Successfully")
}

func (h *AuthHandler) Login(c *gin.Context) {
	op := "auth.LoginAccount"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Parse & validate request
	var req reqdto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "invalid request")
		logger.HandlerInfo(ctx, op, "invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		response.RespondBadRequest(c, "invalid request")
		logger.HandlerInfo(ctx, op, "empty username")
		return
	}

	// Call service
	auth, err := h.AuthSvc.Login(ctx, username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, errorx.ErrUserNotFound),
			errors.Is(err, errorx.ErrPasswordMismatch),
			errors.Is(err, errorx.ErrInvalidHashFormat):
			response.RespondUnauthorized(c, "invalid username or password")
			logger.HandlerInfo(ctx, op, "invalid credentials: %s", err)
			return
		case errors.Is(err, errorx.ErrRoleNotFound):
			response.RespondForbidden(c, "no role assigned for this login scope")
			logger.HandlerInfo(ctx, op, "missing role in login scope")
			return
		default:
			response.RespondInternalError(c, "internal server error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	switch auth.UserState {
	case entity.UserStatusPending:
		response.RespondForbidden(c, "account verification required")
		logger.HandlerInfo(ctx, op, "account pending verification")
		return

	case entity.UserStatusSuspended:
		response.RespondForbidden(c, "account suspended")
		logger.HandlerInfo(ctx, op, "account suspended")
		return

	case entity.UserStatusDeleted:
		response.RespondForbidden(c, "account deleted")
		logger.HandlerInfo(ctx, op, "account deleted")
		return

	case entity.UserStatusActive:
		// continue
	default:
		response.RespondInternalError(c, "invalid user state")
		logger.HandlerWarn(ctx, op, "unknown user state")
		return
	}

	// ===== Only ACTIVE users reach here =====
	if auth.MFARequired {
		response.RespondSuccess(c, gin.H{
			"mfa_required":            true,
			"methods":                 auth.MFAMethods,
			"mfa_session":             auth.MFASession,
			"mfa_session_ttl_seconds": auth.MFASessionTTLSeconds,
			"user_id":                 auth.UserID.String(),
		}, "mfa required")
		logger.HandlerInfo(ctx, op, "mfa required")
		return
	}

	secure := c.Request.TLS != nil
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		secure = true
	}

	accessMaxAge := int(h.Token.AccessTTL.Seconds())
	refreshMaxAge := int(h.Token.RefreshTTL.Seconds())

	c.SetSameSite(http.SameSiteLaxMode)
	h.setAuthCookies(c, auth.AccessToken, auth.RefreshToken, accessMaxAge, refreshMaxAge, secure)
	h.setDeviceCookies(c, auth.DeviceID, auth.DeviceSecret, secure)

	response.RespondSuccess(c, nil, "login successful")
	logger.HandlerInfo(ctx, op, "login successful")
}

func (h *AuthHandler) VerifyMFAChallenge(c *gin.Context) {
	op := "auth.VerifyMFAChallenge"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req reqdto.VerifyMFAChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "invalid request")
		logger.HandlerInfo(ctx, op, "invalid request body")
		return
	}

	userID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		response.RespondBadRequest(c, "invalid request")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}

	auth, err := h.AuthSvc.VerifyMFAChallenge(
		ctx,
		userID,
		strings.TrimSpace(req.MFASession),
		entity.MFAMethodType(strings.ToLower(strings.TrimSpace(req.Method))),
		strings.TrimSpace(req.Code),
	)
	if err != nil {
		switch {
		case errors.Is(err, errorx.ErrInvalidArgument):
			response.RespondBadRequest(c, "invalid request")
			logger.HandlerInfo(ctx, op, "invalid request")
			return
		case errors.Is(err, errorx.ErrTokenInvalid),
			errors.Is(err, errorx.ErrTokenExpired),
			errors.Is(err, errorx.ErrMFAMethodNotFound),
			errors.Is(err, errorx.ErrMFACodeInvalid):
			response.RespondUnauthorized(c, "invalid mfa challenge")
			logger.HandlerInfo(ctx, op, "invalid mfa challenge")
			return
		case errors.Is(err, errorx.ErrUserIsPending),
			errors.Is(err, errorx.ErrUserIsSuspended),
			errors.Is(err, errorx.ErrUserIsDeleted):
			response.RespondForbidden(c, "account is not active")
			logger.HandlerInfo(ctx, op, "account not active")
			return
		default:
			response.RespondInternalError(c, "internal server error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	secure := c.Request.TLS != nil
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		secure = true
	}

	accessMaxAge := int(h.Token.AccessTTL.Seconds())
	refreshMaxAge := int(h.Token.RefreshTTL.Seconds())

	c.SetSameSite(http.SameSiteLaxMode)
	h.setAuthCookies(c, auth.AccessToken, auth.RefreshToken, accessMaxAge, refreshMaxAge, secure)
	h.setDeviceCookies(c, auth.DeviceID, auth.DeviceSecret, secure)

	response.RespondSuccess(c, nil, "mfa verified")
	logger.HandlerInfo(ctx, op, "mfa verified")
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	op := "user.Refresh"

	// context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	refreshToken, err := c.Cookie(string(appctxKey.RefreshToken))
	if err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}
	if refreshToken == "" {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	auth, err := h.AuthSvc.Refresh(ctx, refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, errorx.ErrTokenInvalid),
			errors.Is(err, errorx.ErrTokenExpired),
			errors.Is(err, errorx.ErrUserNotFound):
			response.RespondUnauthorized(c, "invalid or expired refresh token")
			logger.HandlerInfo(ctx, op, "invalid refresh token")
			return
		case errors.Is(err, errorx.ErrUserIsPending),
			errors.Is(err, errorx.ErrUserIsSuspended),
			errors.Is(err, errorx.ErrUserIsDeleted):
			response.RespondForbidden(c, "account is not active")
			logger.HandlerInfo(ctx, op, "account not active")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	secure := c.Request.TLS != nil
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		secure = true
	}

	accessMaxAge := int(h.Token.AccessTTL.Seconds())
	refreshMaxAge := int(h.Token.RefreshTTL.Seconds())

	c.SetSameSite(http.SameSiteLaxMode)
	h.setAuthCookies(c, auth.AccessToken, auth.RefreshToken, accessMaxAge, refreshMaxAge, secure)

	response.RespondSuccess(c, nil, "refresh successful")
	logger.HandlerInfo(ctx, op, "refresh successful")

}

func (h *AuthHandler) Logout(c *gin.Context) {
	op := "auth.Logout"

	ctx := c.Request.Context()

	if jti, ok := c.Get("jti"); ok {
		if jtiStr, ok := jti.(string); ok && strings.TrimSpace(jtiStr) != "" {
			ctx = context.WithValue(ctx, appctxKey.KeyJWTID, strings.TrimSpace(jtiStr))
		}
	}
	if exp, ok := c.Get("jwt_exp"); ok {
		if expVal, ok := exp.(int64); ok && expVal > 0 {
			ctx = context.WithValue(ctx, appctxKey.KeyJWTExp, expVal)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.AuthSvc.Logout(ctx); err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}

	secure := c.Request.TLS != nil
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		secure = true
	}

	c.SetSameSite(http.SameSiteLaxMode)
	clearCookie(c, "access_token", secure)
	clearCookie(c, "refresh_token", secure)
	clearCookie(c, "device_id", secure)
	clearCookie(c, "device_secret", secure)

	response.RespondSuccess(c, nil, "logout successful")
	logger.HandlerInfo(ctx, op, "logout successful")
}

func clearCookie(c *gin.Context, name string, secure bool) {
	c.SetCookie(name, "", -1, "/", "", secure, true)
}

func (h *AuthHandler) setAuthCookies(
	c *gin.Context,
	accessToken string,
	refreshToken string,
	accessMaxAge int,
	refreshMaxAge int,
	secure bool,
) {
	accessCookieMaxAge := accessMaxAge
	if refreshMaxAge > accessCookieMaxAge {
		accessCookieMaxAge = refreshMaxAge
	}
	c.SetCookie("access_token", accessToken, accessCookieMaxAge, "/", "", secure, true)
	c.SetCookie("refresh_token", refreshToken, refreshMaxAge, "/", "", secure, true)
}

func (h *AuthHandler) setDeviceCookies(
	c *gin.Context,
	deviceID uuid.UUID,
	deviceSecret string,
	secure bool,
) {
	if deviceID == uuid.Nil || strings.TrimSpace(deviceSecret) == "" {
		return
	}
	c.SetCookie("device_id", deviceID.String(), 60*24*60*60, "/", "", secure, true)
	c.SetCookie("device_secret", deviceSecret, 7*24*60*60, "/", "", secure, true)
}
