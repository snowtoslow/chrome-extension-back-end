package auth

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidAccessToken = errors.New("invalid access token")
	ErrInvalidPassword    = errors.New("correct email but invalid password")
)

type ValidationErrors []validator.FieldError

func Simple(verr validator.ValidationErrors) map[string]string {
	errs := make(map[string]string)

	for _, f := range verr {
		err := f.ActualTag()
		if f.Param() != "" {
			err = fmt.Sprintf("%s=%s", err, f.Param())
		}
		errs[f.Field()] = err
	}

	return errs
}
