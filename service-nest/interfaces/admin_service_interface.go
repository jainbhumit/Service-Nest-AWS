package interfaces

import (
	"context"
	"service-nest/model"
)

type AdminService interface {
	GetAllService(limit, offset int) ([]model.Service, error)
	DeactivateAccount(userID string) error
	DeleteService(ctx context.Context, serviceID string) error
	ViewReports(limit, offset int) ([]model.ServiceRequest, error)
	AddService(ctx context.Context, category *model.Category) error
	GetUserByEmail(ctx context.Context, userEmail string) (*model.User, error)
}
