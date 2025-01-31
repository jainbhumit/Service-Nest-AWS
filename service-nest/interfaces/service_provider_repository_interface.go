package interfaces

import (
	"context"
	"service-nest/model"
)

type ServiceProviderRepository interface {
	UpdateServiceProvider(provider *model.ServiceProvider) error
	GetProviderByServiceID(serviceID string) (*model.ServiceProvider, error)
	GetProvidersByServiceType(serviceType string) ([]model.ServiceProvider, error)
	GetProviderByID(providerID string) (*model.ServiceProvider, error)
	SaveServiceProvider(provider model.ServiceProvider) error
	GetProviderDetailByID(providerID string, serviceId string) (*model.ServiceProviderDetails, error)
	SaveServiceProviderDetail(provider *model.ServiceProviderDetails, requestID string, serviceID string) error
	UpdateServiceProviderDetailByRequestID(provider *model.ServiceProviderDetails, requestID string) error
	IsProviderApproved(providerID string) (bool, error)
	AddReview(ctx context.Context, review model.Review) error
	UpdateProviderRating(providerID string, serviceId string, rating float64) error
	GetReviewsByProviderID(providerID string, limit, offset int, serviceID string) ([]model.Review, error)
	AddServiceToProvider(providerID, serviceID string) error
	DeleteServicesByProviderID(userID string) error
}
