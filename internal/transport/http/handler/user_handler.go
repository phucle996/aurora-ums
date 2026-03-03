package handler

import (
	"aurora/internal/domain/entity"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	reqdto "aurora/internal/transport/http/handler/dto/request"
	"aurora/internal/transport/http/response"
	"aurora/pkg/logger"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	UserSvc domainsvc.UserSvcInterface
}

func NewUserHandler(UserSvc domainsvc.UserSvcInterface) *UserHandler {
	return &UserHandler{
		UserSvc: UserSvc,
	}
}

func (h *UserHandler) Me(c *gin.Context) {
	op := "user.Me"

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

	currentUser, err := h.UserSvc.GetCurrentUser(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, errorx.ErrUserNotFound):
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, op, "user not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, gin.H{
		"id":          currentUser.User.ID,
		"email":       currentUser.User.Email,
		"status":      currentUser.User.Status,
		"on_boarding": currentUser.User.OnBoarding,
		"roles":       currentUser.Roles,
		"permissions": currentUser.Permissions,
		"profile": gin.H{
			"full_name":       currentUser.Profile.FullName,
			"company":         currentUser.Profile.Company,
			"referral_source": currentUser.Profile.ReferralSource,
			"phone":           currentUser.Profile.Phone,
			"job_function":    currentUser.Profile.JobFunction,
			"country":         currentUser.Profile.Country,
			"avatar_url":      currentUser.Profile.AvatarURL,
			"bio":             currentUser.Profile.Bio,
		},
	}, "ok")
	logger.HandlerInfo(ctx, op, "get profile success")
}

func (h *UserHandler) UpsertProfile(c *gin.Context) {
	op := "user.UpsertProfile"

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

	var req reqdto.UpsertProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	now := time.Now().UTC()
	profile := &entity.Profile{
		ID:             uuid.New(),
		UserID:         userID,
		FullName:       stringPtr(req.FullName),
		Company:        stringPtr(req.Company),
		ReferralSource: stringPtr(req.ReferralSource),
		Phone:          stringPtr(req.Phone),
		JobFunction:    stringPtr(req.JobFunction),
		Country:        stringPtr(req.Country),
		AvatarURL:      stringPtr(req.AvatarURL),
		Bio:            stringPtr(req.Bio),
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	if err := h.UserSvc.UpsertProfile(ctx, profile); err != nil {
		switch {
		case errors.Is(err, errorx.ErrUserNotFound):
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, op, "user not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, nil, "profile updated")
	logger.HandlerInfo(ctx, op, "profile updated")
}

func stringPtr(input string) *string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func readStringContextValue(c *gin.Context, key string) string {
	raw, ok := c.Get(key)
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func uuidPtrToString(v *uuid.UUID) *string {
	if v == nil || *v == uuid.Nil {
		return nil
	}
	str := v.String()
	return &str
}

func (h *UserHandler) ActiveAccount(c *gin.Context) {
	op := "user.ActiveAccount"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req reqdto.VerifyAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	if err := h.UserSvc.ActiveAccount(ctx, userID, req.Token); err != nil {
		switch {
		case errors.Is(err, errorx.ErrOttNotFound),
			errors.Is(err, errorx.ErrUserNotFound):
			response.RespondBadRequest(c, "Invalid or expired activation token")
			logger.HandlerInfo(ctx, op, "Invalid or expired activation token: %s", err)
			return
		case errors.Is(err, errorx.ErrAccountAlreadyActivated):
			response.RespondConflict(c, "Account already activated")
			logger.HandlerInfo(ctx, op, "Account already activated")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, nil, "Active Account Successfully")
	logger.HandlerInfo(ctx, op, "Active Account Successfully")
}

func (h *UserHandler) ForgotPasswd(c *gin.Context) {
	op := "user.ForgotPasswd"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req reqdto.ForgotPasswdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	if err := h.UserSvc.ForgotPasswd(ctx, req.Email); err != nil {
		switch {
		case errors.Is(err, errorx.ErrUserNotFound):
			response.RespondSuccess(c, nil, "If the account exists, a reset link has been sent")
			logger.HandlerInfo(ctx, op, "forgot password: user not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, nil, "If the account exists, a reset link has been sent")
	logger.HandlerInfo(ctx, op, "forgot password accepted")
}

func (h *UserHandler) VerifyResetPassword(c *gin.Context) {
	op := "user.VerifyResetPassword"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req reqdto.VerifyResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	if err := h.UserSvc.VerifyResetPassword(ctx, userID, req.Token); err != nil {
		switch {
		case errors.Is(err, errorx.ErrOttNotFound),
			errors.Is(err, errorx.ErrUserNotFound):
			response.RespondBadRequest(c, "Invalid or expired reset token")
			logger.HandlerInfo(ctx, op, "invalid reset token")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, nil, "Reset token is valid")
	logger.HandlerInfo(ctx, op, "reset token verified")
}

func (h *UserHandler) NewPassword(c *gin.Context) {
	op := "user.NewPassword"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req reqdto.NewPasswordRequest
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

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	if err := h.UserSvc.NewPassword(ctx, userID, req.Token, req.Password); err != nil {
		switch {
		case errors.Is(err, errorx.ErrOttNotFound),
			errors.Is(err, errorx.ErrUserNotFound):
			response.RespondBadRequest(c, "Invalid or expired reset token")
			logger.HandlerInfo(ctx, op, "invalid reset token")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondSuccess(c, nil, "Password updated successfully")
	logger.HandlerInfo(ctx, op, "password updated")
}
