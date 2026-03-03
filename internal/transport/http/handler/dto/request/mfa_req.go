package reqdto

type VerifyTOTPRequest struct {
	Code string `json:"code" binding:"required"`
}

type VerifyMFAChallengeRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Method     string `json:"method" binding:"required"`
	Code       string `json:"code" binding:"required"`
	MFASession string `json:"mfa_session" binding:"required"`
}
