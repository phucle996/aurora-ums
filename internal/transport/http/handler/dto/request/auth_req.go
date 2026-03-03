package reqdto

type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=6"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	RePassword string `json:"re_password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type VerifyAccountRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

type ForgotPasswdRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyResetPasswordRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

type NewPasswordRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Token      string `json:"token" binding:"required"`
	Password   string `json:"password" binding:"required,min=8"`
	RePassword string `json:"re_password" binding:"required"`
}

type UpsertProfileRequest struct {
	FullName       string `json:"full_name"`
	Company        string `json:"company"`
	ReferralSource string `json:"referral_source"`
	Phone          string `json:"phone"`
	JobFunction    string `json:"job_function"`
	Country        string `json:"country"`
	AvatarURL      string `json:"avatar_url"`
	Bio            string `json:"bio"`
}
