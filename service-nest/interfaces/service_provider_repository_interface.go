package interfaces

import (
	"context"
	"service-nest/model"
)

type ServiceProviderRepository interface {
	AddReview(ctx context.Context, review model.Review) error
	GetReviewsByProviderID(ctx context.Context, providerID string, limit, offset int, serviceID string) ([]model.Review, error)
}
