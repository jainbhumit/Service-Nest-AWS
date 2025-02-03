package interfaces

import (
	"context"
	"service-nest/model"
)

type ServiceRepository interface {
	RemoveCategory(ctx context.Context, serviceID string) error
	SaveService(ctx context.Context, service model.Service) error
	GetServiceByProviderID(ctx context.Context, providerID string) ([]model.Service, error)
	UpdateService(ctx context.Context, providerID string, updatedService model.Service) error
	RemoveServiceByProviderID(ctx context.Context, providerID string, serviceID string) error
	GetAllCategory(ctx context.Context) ([]model.Category, error)
	AddCategory(ctx context.Context, category *model.Category) error
	GetServicesByCategoryId(ctx context.Context, categoryId string) ([]model.Service, error)
	GetProviderByServiceId(ctx context.Context, providerID string, serviceId string) (*model.Service, error)
	UpdateProviderRating(ctx context.Context, provider *model.Service) error
	GetAllServiceProviderService(ctx context.Context, limit int, offset int, categoryId string) ([]model.Service, error)
}
