package user

import "errors"

var (
	UserNotFound   = errors.New("no users found")
	InvalidIdError = errors.New("invalid id")
)
