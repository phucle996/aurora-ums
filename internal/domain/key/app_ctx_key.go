package appctxKey

// Key defines a typed key for context values.
type Key string

const (
	KeyRequestID    Key = "request_id"
	KeyTraceParent  Key = "traceparent"
	KeyUserID       Key = "user_id"
	KeyDeviceID     Key = "device_id"
	KeyPermissions  Key = "permissions"
	KeyUserLevel    Key = "user_level"
	KeyRoles        Key = "roles"
	KeyDeviceSecret Key = "device_secret"
	KeyJWTID        Key = "jwt_id"
	KeyJWTExp       Key = "jwt_exp"
	KeyJWTDeviceID  Key = "jwt_device_id"
	KeyTenantID     Key = "tenant_id"
	KeyWorkspaceID  Key = "workspace_id"
)
