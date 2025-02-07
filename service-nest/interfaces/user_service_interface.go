package interfaces

import (
	"context"
	"service-nest/model"
)

type UserService interface {
	CreateUser(ctx context.Context, user *model.User) error
	CheckUserExists(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, userID string, newEmail, newPassword, newAddress, newPhone *string) error
	ViewProfileByID(ctx context.Context, userID string) (*model.User, error)
	GenerateOtp(ctx context.Context, email string) error
	VerifyAndUpdatePassword(ctx context.Context, email, password string, otp string) error
}
