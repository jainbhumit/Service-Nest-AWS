package interfaces

import (
	"context"
	"service-nest/model"
)

type AdminService interface {
	DeactivateAccount(ctx context.Context, userID string) error
	DeleteService(ctx context.Context, serviceID string) error
	ViewReports(ctx context.Context, limit, offset int, categoryId string) ([]model.Service, error)
	AddService(ctx context.Context, category *model.Category) error
	GetUserByEmail(ctx context.Context, userEmail string) (*model.User, error)
}
