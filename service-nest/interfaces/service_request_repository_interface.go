package interfaces

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"service-nest/model"
)

type ServiceRequestRepository interface {
	UpdateServiceRequest(ctx context.Context, updatedRequest *model.ServiceRequest, status string) error
	GetServiceRequestsByHouseholderID(ctx context.Context, householderID string, limit int, lastEvaluatedKey map[string]types.AttributeValue, status string) ([]model.ServiceRequest, map[string]types.AttributeValue, error)
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
