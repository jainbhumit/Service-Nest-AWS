package interfaces

import (
	"context"
	"service-nest/model"
	"time"
)

type HouseholderService interface {
	ViewApprovedRequests(ctx context.Context, householderID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error)
	ApproveServiceRequest(ctx context.Context, requestID string, providerID string, householderID string, serviceId string, status string) error
	AddReview(ctx context.Context, providerID, householderID, serviceID, comments string, rating float64) error
	RescheduleServiceRequest(ctx context.Context, requestID string, newTime time.Time, householderID string, status string) error
	CancelServiceRequest(ctx context.Context, requestID string, householderID string, status string) error
	GetAvailableServices(limit, offset int) ([]model.Service, error)
	RequestService(ctx context.Context, houseHolderRequest model.ServiceRequest) (string, error)
	GetServicesByCategoryId(ctx context.Context, categoryId string) ([]model.Service, error)
	SearchService(householder *model.Householder, serviceType string) ([]model.ServiceProvider, error)
	ViewStatus(ctx context.Context, householderID string, limit, offset int, status string) ([]model.ServiceRequest, error)
	GetAllServiceCategory(ctx context.Context) ([]model.Category, error)
}
