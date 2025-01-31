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
	DeActivateUser(userID string) error
	GetSecurityAnswerByEmail(userEmail string) (*string, error)
	UpdatePassword(userEmail, updatedPassword string) error
}
