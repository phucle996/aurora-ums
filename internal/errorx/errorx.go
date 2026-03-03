package errorx

import "errors"

var (
	ErrEntityNil = errors.New("Entity is nil")

	// ErrInvalidHashFormat is returned when the stored hash doesn't match expected format.
	ErrInvalidHashFormat = errors.New("invalid password hash format")
	// ErrPasswordMismatch is returned when password verification fails.
	ErrPasswordMismatch = errors.New("password does not match")
)

// user account

var (
	ErrUserAlreadyExist = errors.New("User Already exist in system")

	ErrUserNotFound    = errors.New("User not found")
	ErrUnexpectedRows  = errors.New("unexpected affected rows")
	ErrUserIsPending   = errors.New("User is pending account")
	ErrUserIsSuspended = errors.New("This account has been suspended ")
	ErrUserIsDeleted   = errors.New("This account has been deleted")
	ErrProfileNotFound = errors.New("Profile not found")
)

// one-time token
var (
	ErrOttNotFound = errors.New("One-time token not found")
)

// mfa
var (
	ErrMFAMethodNotFound       = errors.New("MFA method not found")
	ErrMFAChallengeNotFound    = errors.New("MFA challenge not found")
	ErrMFAMethodAlreadyEnabled = errors.New("MFA method already enabled")
	ErrMFACodeInvalid          = errors.New("MFA code invalid")
)

// token
var (
	ErrTokenInvalid   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token expired")
	ErrTokenNotActive = errors.New("token not active")
)

// common
var (
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrPermissionDenied = errors.New("permission denied")
)

// activation account token
var (
	ErrAccountAlreadyActivated = errors.New("Account is already activated")
)

// rbac
var (
	ErrRoleNotFound           = errors.New("Role not found")
	ErrRoleAlreadyExist       = errors.New("Role already exists")
	ErrPermissionNotFound     = errors.New("Permission not found")
	ErrRolePermissionNotFound = errors.New("Role permission not found")
	ErrUserRoleNotFound       = errors.New("User role not found")
)
