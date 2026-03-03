package handler

import (
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	reqdto "aurora/internal/transport/http/handler/dto/request"
	"aurora/internal/transport/http/response"
	"aurora/pkg/logger"
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MfaHandler struct {
	MfaSvc domainsvc.MFASvcInterface
}

func NewMfaHandler(mfaSvc domainsvc.MFASvcInterface) *MfaHandler {
	return &MfaHandler{MfaSvc: mfaSvc}
}

func (h *MfaHandler) BeginTOTPSetup(c *gin.Context) {
	op := "mfa.BeginTOTPSetup"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rawUserID, ok := c.Get("user_id")
	if !ok {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "missing user id")
		return
	}
	userIDStr, ok := rawUserID.(string)
	if !ok || userIDStr == "" {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}

	secret, otpauthURL, err := h.MfaSvc.BeginTOTPSetup(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, errorx.ErrMFAMethodAlreadyEnabled):
			response.RespondConflict(c, "2fa already enabled")
			logger.HandlerInfo(ctx, op, "2fa already enabled")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, gin.H{
		"secret":      secret,
		"otpauth_url": otpauthURL,
	}, "totp setup created")
	logger.HandlerInfo(ctx, op, "totp setup created")
}

func (h *MfaHandler) VerifyTOTPSetup(c *gin.Context) {
	op := "mfa.VerifyTOTPSetup"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rawUserID, ok := c.Get("user_id")
	if !ok {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "missing user id")
		return
	}
	userIDStr, ok := rawUserID.(string)
	if !ok || userIDStr == "" {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}

	var req reqdto.VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	if err := h.MfaSvc.VerifyAndEnableTOTP(ctx, userID, req.Code); err != nil {
		switch {
		case errors.Is(err, errorx.ErrMFAMethodNotFound):
			response.RespondBadRequest(c, "mfa method not found")
			logger.HandlerInfo(ctx, op, "mfa method not found")
			return
		case errors.Is(err, errorx.ErrMFAMethodAlreadyEnabled):
			response.RespondConflict(c, "2fa already enabled")
			logger.HandlerInfo(ctx, op, "2fa already enabled")
			return
		case errors.Is(err, errorx.ErrMFACodeInvalid):
			response.RespondBadRequest(c, "invalid code")
			logger.HandlerInfo(ctx, op, "invalid code")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, nil, "2fa enabled")
	logger.HandlerInfo(ctx, op, "2fa enabled")
}

func (h *MfaHandler) Disable2FA(c *gin.Context) {
	op := "mfa.Disable2FA"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rawUserID, ok := c.Get("user_id")
	if !ok {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "missing user id")
		return
	}
	userIDStr, ok := rawUserID.(string)
	if !ok || userIDStr == "" {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}

	if err := h.MfaSvc.DisableTOTP(ctx, userID); err != nil {
		switch {
		case errors.Is(err, errorx.ErrMFAMethodNotFound):
			response.RespondBadRequest(c, "mfa method not found")
			logger.HandlerInfo(ctx, op, "mfa method not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, nil, "2fa disabled")
	logger.HandlerInfo(ctx, op, "2fa disabled")
}

func (h *MfaHandler) GenerateRecoveryCodes(c *gin.Context) {
	op := "mfa.GenerateRecoveryCodes"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rawUserID, ok := c.Get("user_id")
	if !ok {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "missing user id")
		return
	}
	userIDStr, ok := rawUserID.(string)
	if !ok || userIDStr == "" {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}

	codes, err := h.MfaSvc.GenerateRecoveryCodes(ctx, userID, 10)
	if err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}

	response.RespondSuccess(c, gin.H{"codes": codes}, "recovery codes created")
	logger.HandlerInfo(ctx, op, "recovery codes created")
}

func (h *MfaHandler) ListMethods(c *gin.Context) {
	op := "mfa.ListMethods"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rawUserID, ok := c.Get("user_id")
	if !ok {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "missing user id")
		return
	}
	userIDStr, ok := rawUserID.(string)
	if !ok || userIDStr == "" {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondUnauthorized(c, "unauthorized")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}

	methods, err := h.MfaSvc.ListEnabledMethods(ctx, userID)
	if err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}

	response.RespondSuccess(c, gin.H{"methods": methods}, "ok")
	logger.HandlerInfo(ctx, op, "list methods success")
}

// func (h *MfaHandler) IssueChallenge(c *gin.Context) {
// 	op := "mfa.IssueChallenge"
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
// 	defer cancel()
// 	response.RespondNotImplemented(c, "not implemented")
// 	logger.HandlerInfo(ctx, op, "not implemented")
// }

// func (h *MfaHandler) VerifyChallenge(c *gin.Context) {
// 	op := "mfa.VerifyChallenge"
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
// 	defer cancel()
// 	response.RespondNotImplemented(c, "not implemented")
// 	logger.HandlerInfo(ctx, op, "not implemented")
// }
