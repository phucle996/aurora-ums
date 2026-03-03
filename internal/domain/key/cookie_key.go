package appctxKey

// Key defines a typed key for context values.
type CookieKey string

const (
	AccessToken  CookieKey = "access_token"
	RefreshToken CookieKey = "refresh_token"
)
