package auth

import "errors"

var (
	ErrInvalidEmail        = errors.New("email is invalid")
	ErrInvalidRefreshToken = errors.New("refresh token is invalid")
)
