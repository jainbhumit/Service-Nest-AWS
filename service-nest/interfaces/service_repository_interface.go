package interfaces

import (
	"context"
	"service-nest/model"
)

type ServiceRepository interface {
	RemoveCategory(ctx context.Context, serviceID string) error
	SaveService(ctx context.Context, service model.Service) error
	GetAllServices(limit, offset int) ([]model.Service, error)
	GetServiceByID(serviceID string) (*model.Service, error)
	GetServiceIdByCategory(category string) (*string, error)
	GetServiceByProviderID(ctx context.Context, providerID string) ([]model.Service, error)
	UpdateService(ctx context.Context, providerID string, updatedService model.Service) error
	RemoveServiceByProviderID(ctx context.Context, providerID string, serviceID string) error
	CategoryExists(categoryName string) (bool, error)
	GetAllCategory(ctx context.Context) ([]model.Category, error)
	AddCategory(ctx context.Context, category *model.Category) error
	GetServicesByCategoryId(ctx context.Context, categoryId string) ([]model.Service, error)
	GetProviderByServiceId(ctx context.Context, providerID string, serviceId string) (*model.Service, error)
}
