package user

import (
	"chrome-extension-back-end/models"
	"context"
)

type UseCase interface {
	CreateUser(ctx context.Context, user *models.User) error
	//GetUsers(ctx *context.Context) (users []*models.User, err error)
	GetUserById(ctx context.Context, id string) (user *models.User, err error)
}
