package handler

import "regexp"

var (
	passwordUpperRe   = regexp.MustCompile(`[A-Z]`)
	passwordLowerRe   = regexp.MustCompile(`[a-z]`)
	passwordNumberRe  = regexp.MustCompile(`[0-9]`)
	passwordSpecialRe = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func validatePassword(password string) (string, string) {
	if len(password) < 8 {
		return "password must be at least 8 characters", "password too short"
	}
	if !passwordUpperRe.MatchString(password) {
		return "password must include an uppercase letter", "password missing uppercase"
	}
	if !passwordLowerRe.MatchString(password) {
		return "password must include a lowercase letter", "password missing lowercase"
	}
	if !passwordNumberRe.MatchString(password) {
		return "password must include a number", "password missing number"
	}
	if !passwordSpecialRe.MatchString(password) {
		return "password must include a special character", "password missing special"
	}
	return "", ""
}
