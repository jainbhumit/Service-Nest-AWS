package interfaces

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"service-nest/model"
	"time"
)

type HouseholderService interface {
	ViewApprovedRequests(ctx context.Context, householderID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error)
	ApproveServiceRequest(ctx context.Context, requestID string, providerID string, householderID string, serviceId string, status string) error
	AddReview(ctx context.Context, providerID, householderID, serviceID, comments string, rating float64, requestId string) error
	RescheduleServiceRequest(ctx context.Context, requestID string, newTime time.Time, householderID string, status string) error
	CancelServiceRequest(ctx context.Context, requestID string, householderID string, status string) error
	RequestService(ctx context.Context, houseHolderRequest model.ServiceRequest) (string, error)
	GetServicesByCategoryId(ctx context.Context, categoryId string) ([]model.Service, error)
	ViewStatus(ctx context.Context, householderID string, limit int, lastEvaluatedKey map[string]types.AttributeValue, status string) ([]model.ServiceRequest, map[string]types.AttributeValue, error)
	GetAllServiceCategory(ctx context.Context) ([]model.Category, error)
}
