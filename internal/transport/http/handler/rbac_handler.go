package handler

import (
	"aurora/internal/domain/entity"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	"aurora/internal/transport/http/response"
	"aurora/pkg/logger"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RbacHandler struct {
	RbacSvc domainsvc.RBACSvcInterface
}

func NewRbacHandler(rbacSvc domainsvc.RBACSvcInterface) *RbacHandler {
	return &RbacHandler{RbacSvc: rbacSvc}
}

func (h *RbacHandler) ListRoles(c *gin.Context) {
	op := "rbac.ListRoles"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	roles, err := h.RbacSvc.ListRoles(ctx)
	if err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}
	response.RespondSuccess(c, roles, "ok")
	logger.HandlerInfo(ctx, op, "list roles")
}

func (h *RbacHandler) CreateRole(c *gin.Context) {
	op := "rbac.CreateRole"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondBadRequest(c, "bad request")
		logger.HandlerInfo(ctx, op, "bad request")
		return
	}

	role := &entity.Role{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
	}

	if err := h.RbacSvc.CreateRole(ctx, role); err != nil {
		switch {
		case errors.Is(err, errorx.ErrRoleAlreadyExist):
			response.RespondConflict(c, "role already exists")
			logger.HandlerInfo(ctx, op, "role already exists")
			return
		case errors.Is(err, errorx.ErrInvalidArgument):
			response.RespondBadRequest(c, "invalid name")
			logger.HandlerInfo(ctx, op, "invalid name")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}

	response.RespondCreated(c, role, "created")
	logger.HandlerInfo(ctx, op, "role created")
}

func (h *RbacHandler) GetRole(c *gin.Context) {
	op := "rbac.GetRole"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid role id")
		logger.HandlerInfo(ctx, op, "invalid role id")
		return
	}
	role, err := h.RbacSvc.GetRoleByID(ctx, roleID)
	if err != nil {
		switch {
		case errors.Is(err, errorx.ErrRoleNotFound):
			response.RespondNotFound(c, "role not found")
			logger.HandlerInfo(ctx, op, "role not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}
	response.RespondSuccess(c, role, "ok")
	logger.HandlerInfo(ctx, op, "get role")
}

func (h *RbacHandler) DeleteRole(c *gin.Context) {
	op := "rbac.DeleteRole"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid role id")
		logger.HandlerInfo(ctx, op, "invalid role id")
		return
	}
	if err := h.RbacSvc.DeleteRole(ctx, roleID); err != nil {
		switch {
		case errors.Is(err, errorx.ErrRoleNotFound):
			response.RespondNotFound(c, "role not found")
			logger.HandlerInfo(ctx, op, "role not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}
	response.RespondSuccess(c, nil, "deleted")
	logger.HandlerInfo(ctx, op, "role deleted")
}

func (h *RbacHandler) ListPermissions(c *gin.Context) {
	op := "rbac.ListPermissions"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	perms, err := h.RbacSvc.ListPermissions(ctx)
	if err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}
	response.RespondSuccess(c, perms, "ok")
	logger.HandlerInfo(ctx, op, "list permissions")
}

func (h *RbacHandler) GetPermission(c *gin.Context) {
	op := "rbac.GetPermission"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	permID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid permission id")
		logger.HandlerInfo(ctx, op, "invalid permission id")
		return
	}
	perm, err := h.RbacSvc.GetPermissionByID(ctx, permID)
	if err != nil {
		switch {
		case errors.Is(err, errorx.ErrPermissionNotFound):
			response.RespondNotFound(c, "permission not found")
			logger.HandlerInfo(ctx, op, "permission not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}
	response.RespondSuccess(c, perm, "ok")
	logger.HandlerInfo(ctx, op, "get permission")
}

func (h *RbacHandler) AddPermissionToRole(c *gin.Context) {
	op := "rbac.AddPermissionToRole"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid role id")
		logger.HandlerInfo(ctx, op, "invalid role id")
		return
	}
	permID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid permission id")
		logger.HandlerInfo(ctx, op, "invalid permission id")
		return
	}
	if err := h.RbacSvc.AddPermissionToRole(ctx, roleID, permID); err != nil {
		switch {
		case errors.Is(err, errorx.ErrRoleNotFound):
			response.RespondNotFound(c, "role not found")
			logger.HandlerInfo(ctx, op, "role not found")
			return
		case errors.Is(err, errorx.ErrPermissionNotFound):
			response.RespondNotFound(c, "permission not found")
			logger.HandlerInfo(ctx, op, "permission not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}
	response.RespondSuccess(c, nil, "permission added to role")
	logger.HandlerInfo(ctx, op, "permission added to role")
}

func (h *RbacHandler) RemovePermissionFromRole(c *gin.Context) {
	op := "rbac.RemovePermissionFromRole"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid role id")
		logger.HandlerInfo(ctx, op, "invalid role id")
		return
	}
	permID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid permission id")
		logger.HandlerInfo(ctx, op, "invalid permission id")
		return
	}
	if err := h.RbacSvc.RemovePermissionFromRole(ctx, roleID, permID); err != nil {
		switch {
		case errors.Is(err, errorx.ErrRolePermissionNotFound):
			response.RespondNotFound(c, "role permission not found")
			logger.HandlerInfo(ctx, op, "role permission not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}
	response.RespondSuccess(c, nil, "permission removed from role")
	logger.HandlerInfo(ctx, op, "permission removed from role")
}

func (h *RbacHandler) ListRolePermissions(c *gin.Context) {
	op := "rbac.ListRolePermissions"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid role id")
		logger.HandlerInfo(ctx, op, "invalid role id")
		return
	}
	perms, err := h.RbacSvc.ListRolePermissions(ctx, roleID)
	if err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}
	response.RespondSuccess(c, perms, "ok")
	logger.HandlerInfo(ctx, op, "list role permissions")
}

func (h *RbacHandler) AssignRoleToUser(c *gin.Context) {
	op := "rbac.AssignRoleToUser"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid user id")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid role id")
		logger.HandlerInfo(ctx, op, "invalid role id")
		return
	}
	if err := h.RbacSvc.AssignRoleToUser(ctx, userID, roleID); err != nil {
		switch {
		case errors.Is(err, errorx.ErrRoleNotFound):
			response.RespondNotFound(c, "role not found")
			logger.HandlerInfo(ctx, op, "role not found")
			return
		case errors.Is(err, errorx.ErrUserNotFound):
			response.RespondNotFound(c, "user not found")
			logger.HandlerInfo(ctx, op, "user not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}
	response.RespondSuccess(c, nil, "role assigned")
	logger.HandlerInfo(ctx, op, "role assigned")
}

func (h *RbacHandler) RemoveRoleFromUser(c *gin.Context) {
	op := "rbac.RemoveRoleFromUser"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid user id")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid role id")
		logger.HandlerInfo(ctx, op, "invalid role id")
		return
	}
	if err := h.RbacSvc.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		switch {
		case errors.Is(err, errorx.ErrUserRoleNotFound):
			response.RespondNotFound(c, "user role not found")
			logger.HandlerInfo(ctx, op, "user role not found")
			return
		default:
			response.RespondInternalError(c, "Internal Server Error")
			logger.HandlerWarn(ctx, op, "%s", err.Error())
			return
		}
	}
	response.RespondSuccess(c, nil, "role removed")
	logger.HandlerInfo(ctx, op, "role removed")
}

func (h *RbacHandler) ListUserRoles(c *gin.Context) {
	op := "rbac.ListUserRoles"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid user id")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	roles, err := h.RbacSvc.ListUserRoles(ctx, userID)
	if err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}
	response.RespondSuccess(c, roles, "ok")
	logger.HandlerInfo(ctx, op, "list user roles")
}

func (h *RbacHandler) ListUserPermissions(c *gin.Context) {
	op := "rbac.ListUserPermissions"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.RespondBadRequest(c, "invalid user id")
		logger.HandlerInfo(ctx, op, "invalid user id")
		return
	}
	perms, err := h.RbacSvc.ListUserPermissions(ctx, userID)
	if err != nil {
		response.RespondInternalError(c, "Internal Server Error")
		logger.HandlerWarn(ctx, op, "%s", err.Error())
		return
	}
	response.RespondSuccess(c, perms, "ok")
	logger.HandlerInfo(ctx, op, "list user permissions")
}
