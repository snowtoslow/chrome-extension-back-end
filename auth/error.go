package auth

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidAccessToken = errors.New("invalid access token")
	ErrInvalidPassword    = errors.New("correct email but invalid password")
)
