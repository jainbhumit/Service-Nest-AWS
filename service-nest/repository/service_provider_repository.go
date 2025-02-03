package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"service-nest/config"
	"time"

	"service-nest/interfaces"
	"service-nest/model"
)

type ServiceProviderRepository struct {
	client *dynamodb.Client
}

// NewServiceProviderRepository initializes a new ServiceProviderRepository
func NewServiceProviderRepository(client *dynamodb.Client) interfaces.ServiceProviderRepository {
	return &ServiceProviderRepository{client: client}
}

func (s *ServiceProviderRepository) AddReview(ctx context.Context, review model.Review) error {
	item, err := attributevalue.MarshalMap(review)
	if err != nil {
		return fmt.Errorf("failed to marshal category: %v", err)
	}

	item["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", review.ProviderID)}
	item["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("review:%s:%s:%s", review.ServiceID, review.HouseholderID, review.RequestId)}
	item["review_date"] = &types.AttributeValueMemberS{Value: review.ReviewDate.Format(time.RFC3339)}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(config.TABLENAME),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}

	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return fmt.Errorf("review already exists")
		}
		return fmt.Errorf("failed to add review: %v", err)
	}

	return nil
}

func (s ServiceProviderRepository) GetReviewsByProviderID(ctx context.Context, providerID string, limit, offset int, serviceID string) ([]model.Review, error) {
	var input *dynamodb.QueryInput
	if serviceID == "" {
		input = &dynamodb.QueryInput{
			TableName:              aws.String(config.TABLENAME),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", providerID)},
				":skPrefix": &types.AttributeValueMemberS{Value: "review:"},
			},
		}
	} else {
		input = &dynamodb.QueryInput{
			TableName:              aws.String(config.TABLENAME),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", providerID)},
				":skPrefix": &types.AttributeValueMemberS{Value: fmt.Sprintf("review:%s:", serviceID)},
			},
		}
	}

	if limit > 0 {
		input.Limit = aws.Int32(int32(limit))
	}

	// If offset is provided, we need to implement pagination
	var startKey map[string]types.AttributeValue
	if offset > 0 {
		// First, we need to get the item at the offset position
		tempInput := *input
		tempInput.Limit = aws.Int32(int32(offset))
		tempResult, err := s.client.Query(ctx, &tempInput)
		if err != nil {
			return nil, fmt.Errorf("failed to query for offset: %v", err)
		}

		// If we have results and reached the offset
		if len(tempResult.Items) == int(offset) {
			startKey = tempResult.LastEvaluatedKey
			input.ExclusiveStartKey = startKey
		}
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.Review{}, nil
	}

	reviews := make([]model.Review, 0, len(result.Items))
	for _, item := range result.Items {
		var review model.Review
		err = attributevalue.UnmarshalMap(item, &review)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal review: %v", err)
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}
