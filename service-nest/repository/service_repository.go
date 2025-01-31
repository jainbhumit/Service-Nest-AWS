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
	"service-nest/errs"
	"service-nest/interfaces"
	"service-nest/model"
	"strconv"
	"strings"
)

type ServiceRepository struct {
	client *dynamodb.Client
}

func (s ServiceRepository) GetAllServices(limit, offset int) ([]model.Service, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceRepository) GetServiceByID(serviceID string) (*model.Service, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceRepository) GetServiceIdByCategory(category string) (*string, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceRepository) CategoryExists(categoryName string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

// NewServiceRepository creates a new instance of ServiceRepository for MySQL
func NewServiceRepository(client *dynamodb.Client) interfaces.ServiceRepository {
	return &ServiceRepository{client: client}
}

func (s *ServiceRepository) GetAllCategory(ctx context.Context) ([]model.Category, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(config.TABLENAME),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "service"},
		},
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.Category{}, nil
	}

	categories := make([]model.Category, 0, len(result.Items))
	for _, item := range result.Items {
		var category model.Category
		err = attributevalue.UnmarshalMap(item, &category)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal category: %v", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (s *ServiceRepository) AddCategory(ctx context.Context, category *model.Category) error {
	if category.ID == "" {
		return fmt.Errorf("service ID is required")
	}

	item, err := attributevalue.MarshalMap(category)
	if err != nil {
		return fmt.Errorf("failed to marshal category: %v", err)
	}

	// Set PK and SK
	item["PK"] = &types.AttributeValueMemberS{Value: "service"}
	item["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", category.ID)}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(config.TABLENAME),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}

	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return fmt.Errorf("category already exists with ID %s", category.ID)
		}
		return fmt.Errorf("failed to add category: %v", err)
	}

	return nil
}

func (s *ServiceRepository) RemoveCategory(ctx context.Context, serviceID string) error {
	deleteCategoryInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "service"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", serviceID)},
		},
		ConditionExpression: aws.String("attribute_exists(SK)"),
	}
	_, err := s.client.DeleteItem(ctx, deleteCategoryInput)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf("serviceId not exist")
		}
		return fmt.Errorf("fail to delete category: %v", err)
	}
	return nil
}
func (s ServiceRepository) SaveService(ctx context.Context, service model.Service) error {
	serviceItem, err := attributevalue.MarshalMap(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %v", err)
	}

	serviceItem["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", service.ID)}
	serviceItem["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("service_provider:%s", service.ProviderID)}

	providerItem, err := attributevalue.MarshalMap(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %v", err)
	}
	providerItem["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", service.ProviderID)}
	providerItem["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", service.ID)}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(config.TABLENAME),
		Item:                serviceItem,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}
	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return fmt.Errorf("service already exists with ID %s", service.ID)
		}
		return fmt.Errorf("failed to add service: %v", err)
	}
	input = &dynamodb.PutItemInput{
		TableName:           aws.String(config.TABLENAME),
		Item:                providerItem,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}
	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return fmt.Errorf("service already exists with ID %s", service.ID)
		}
		return fmt.Errorf("failed to add service: %v", err)
	}
	return nil
}

func (s ServiceRepository) GetServiceByProviderID(ctx context.Context, providerID string) ([]model.Service, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(config.TABLENAME),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", providerID)},
			":skPrefix": &types.AttributeValueMemberS{Value: "service:"},
		},
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.Service{}, nil
	}

	services := make([]model.Service, 0, len(result.Items))
	for _, item := range result.Items {
		var service model.Service
		err = attributevalue.UnmarshalMap(item, &service)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal service: %v", err)
		}
		services = append(services, service)
	}

	return services, nil
}

func (s *ServiceRepository) UpdateService(ctx context.Context, providerID string, updatedService model.Service) error {
	priceStr := strconv.FormatFloat(updatedService.Price, 'f', -1, 64)
	updateMainInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", providerID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", updatedService.ID)},
		},
		UpdateExpression: aws.String("SET #name = :name, description = :description, price = :price"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name":        &types.AttributeValueMemberS{Value: updatedService.Name},
			":description": &types.AttributeValueMemberS{Value: updatedService.Description},
			":price":       &types.AttributeValueMemberN{Value: priceStr},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	}

	_, err := s.client.UpdateItem(ctx, updateMainInput)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf("service with service is %s does not exist", updatedService.ID)
		}
		return fmt.Errorf("error updatin request: %v", err)
	}

	updateMainInput = &dynamodb.UpdateItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", updatedService.ID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service_provider:%s", providerID)},
		},
		UpdateExpression: aws.String("SET #name = :name, description = :description, price = :price"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name":        &types.AttributeValueMemberS{Value: updatedService.Name},
			":description": &types.AttributeValueMemberS{Value: updatedService.Description},
			":price":       &types.AttributeValueMemberN{Value: priceStr},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	}

	_, err = s.client.UpdateItem(ctx, updateMainInput)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf("service with service is %s does not exist", updatedService.ID)
		}
		return fmt.Errorf("error updating request: %v", err)
	}
	return nil
}

func (s ServiceRepository) RemoveServiceByProviderID(ctx context.Context, providerID string, serviceID string) error {
	deleteCategoryInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", providerID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", serviceID)},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	}
	_, err := s.client.DeleteItem(ctx, deleteCategoryInput)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf("service with service is %s does not exist", serviceID)
		}
		return fmt.Errorf("fail to delete service: %v", err)
	}
	deleteCategoryInput = &dynamodb.DeleteItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", serviceID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service_provider:%s", providerID)},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	}
	_, err = s.client.DeleteItem(ctx, deleteCategoryInput)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf("service with service is %s does not exist", serviceID)
		}
		return fmt.Errorf("fail to delete service: %v", err)
	}
	return nil
}

func (s ServiceRepository) GetServicesByCategoryId(ctx context.Context, categoryId string) ([]model.Service, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(config.TABLENAME),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", categoryId)},
			":skPrefix": &types.AttributeValueMemberS{Value: "service_provider:"},
		},
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.Service{}, nil
	}

	services := make([]model.Service, 0, len(result.Items))
	for _, item := range result.Items {
		var service model.Service
		err = attributevalue.UnmarshalMap(item, &service)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal services: %v", err)
		}
		services = append(services, service)
	}

	return services, nil
}

func (s ServiceRepository) GetProviderByServiceId(ctx context.Context, providerID string, serviceId string) (*model.Service, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service:%s", serviceId)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service_provider:%s", providerID)},
		},
	}

	result, err := s.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %v", err)
	}

	if len(result.Item) == 0 {
		return nil, errors.New(errs.ProviderNotFound)
	}

	var service *model.Service
	err = attributevalue.UnmarshalMap(result.Item, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal service: %v", err)
	}

	return service, nil
}
