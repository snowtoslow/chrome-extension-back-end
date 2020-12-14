package user

import (
	"chrome-extension-back-end/models"
	"context"
)

type UseCase interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmailAndPassword(ctx context.Context, email, password string) (*models.User, error)
	UpdateUser(ctx context.Context, dto *models.PatchDTO) (err error)
	GetUserById(ctx context.Context, id string) (user *models.User, err error)
}
