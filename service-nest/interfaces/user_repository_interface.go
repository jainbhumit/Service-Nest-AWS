package interfaces

import (
	"context"
	"service-nest/model"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	UpdateUser(ctx context.Context, updatedUser *model.User, oldEmail string) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	DeActivateUser(ctx context.Context, userID string, email string) error
	UpdatePassword(ctx context.Context, userEmail, userId string, updatedPassword string) error
	SaveOTP(ctx context.Context, email string, otp string) error
	ValidateOTP(ctx context.Context, email string, otp string) (bool, error)
}
