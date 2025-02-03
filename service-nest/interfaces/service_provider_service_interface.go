package interfaces

import (
	"context"
	"service-nest/model"
)

type ServiceProviderService interface {
	GetReviews(ctx context.Context, providerID string, limit, offset int, serviceID string) ([]model.Review, error)
	ViewApprovedRequestsByProvider(ctx context.Context, providerID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error)
	ViewServices(ctx context.Context, providerID string) ([]model.Service, error)
	AcceptServiceRequest(ctx context.Context, providerID, requestID string, estimatedPrice string, serviceId string, status string) error
	RemoveService(ctx context.Context, providerID, serviceID string) error
	GetAllServiceRequests(ctx context.Context, providerId string, serviceID string, limit, offset int) ([]model.ServiceRequest, error)
	UpdateService(ctx context.Context, updatedService model.Service) error
	AddService(ctx context.Context, providerID string, newService model.Service) (string, error)
}
