package repository

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"service-nest/interfaces"
	"service-nest/model"
)

type ServiceProviderRepository struct {
	Collection *dynamodb.Client
}

func (s ServiceProviderRepository) UpdateServiceProvider(provider *model.ServiceProvider) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) GetProviderByServiceID(serviceID string) (*model.ServiceProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) GetProvidersByServiceType(serviceType string) ([]model.ServiceProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) GetProviderByID(providerID string) (*model.ServiceProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) SaveServiceProvider(provider model.ServiceProvider) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) GetProviderDetailByID(providerID string, serviceId string) (*model.ServiceProviderDetails, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) SaveServiceProviderDetail(provider *model.ServiceProviderDetails, requestID string, serviceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) UpdateServiceProviderDetailByRequestID(provider *model.ServiceProviderDetails, requestID string) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) IsProviderApproved(providerID string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) UpdateProviderRating(providerID string, serviceId string, rating float64) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) GetReviewsByProviderID(providerID string, limit, offset int, serviceID string) ([]model.Review, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) AddServiceToProvider(providerID, serviceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceProviderRepository) DeleteServicesByProviderID(userID string) error {
	//TODO implement me
	panic("implement me")
}

// NewServiceProviderRepository initializes a new ServiceProviderRepository
func NewServiceProviderRepository(collection *dynamodb.Client) interfaces.ServiceProviderRepository {
	return &ServiceProviderRepository{Collection: collection}
}

func (s ServiceProviderRepository) AddReview(ctx context.Context, review model.Review) error {

}
