package interfaces

import (
	"context"
	"service-nest/model"
)

type ServiceRequestRepository interface {
	//SaveAllServiceRequests(serviceRequests []model.ServiceRequest) error
	GetAllServiceRequests(limit, offset int) ([]model.ServiceRequest, error)
	UpdateServiceRequest(ctx context.Context, updatedRequest *model.ServiceRequest, status string) error
	GetServiceRequestsByHouseholderID(ctx context.Context, householderID string, limit, offset int, status string) ([]model.ServiceRequest, error)
	GetServiceRequestByID(ctx context.Context, requestID string, householderId string, status string) (*model.ServiceRequest, error)
	SaveServiceRequest(ctx context.Context, request model.ServiceRequest) error
	GetApproveServiceRequestsByProviderID(ctx context.Context, providerID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error)
	GetServiceProviderByRequestID(requestID, providerID string) (*model.ServiceRequest, error)
	GetApproveServiceRequestsByHouseholderID(ctx context.Context, householderID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error)
	GetAllPendingRequestsByProvider(ctx context.Context, providerId string, serviceID string, limit, offset int) ([]model.ServiceRequest, error)
	CancelServiceRequest(ctx context.Context, request *model.ServiceRequest, prevStatus string) error
	GetServiceRequestByProvider(ctx context.Context, requestID string, serviceId string) (*model.ServiceRequest, error)
	AcceptServiceRequestByProvider(ctx context.Context, updatedRequest *model.ServiceRequest, status string) error
	ApproveServiceRequest(ctx context.Context, serviceRequest *model.ServiceRequest, status string, serviceId string) error
}
