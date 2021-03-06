package usecase

import (
	"chrome-extension-back-end/models"
	"chrome-extension-back-end/user"
	"context"
)

type UserUseCase struct {
	userRepository user.Repository
}

func NewUserUseCase(userRepo user.Repository) *UserUseCase {
	return &UserUseCase{
		userRepository: userRepo,
	}
}

func (u *UserUseCase) CreateUser(ctx context.Context, user *models.User) (err error) {
	return u.userRepository.CreateUser(ctx, user)
}

func (u *UserUseCase) GetUserByEmailAndPassword(ctx context.Context, email, password string) (*models.User, error) {
	return u.userRepository.GetUserByEmailAndPassword(ctx, email, password)
}

func (u *UserUseCase) GetUserById(ctx context.Context, id string) (user *models.User, err error) {
	return u.userRepository.GetUserById(ctx, id)
}

func (u *UserUseCase) UpdateUser(ctx context.Context, dto *models.PatchDTO) (err error) {
	return u.userRepository.UpdateUser(ctx, dto)
}
