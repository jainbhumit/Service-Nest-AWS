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
	"service-nest/util"
	"strings"
	"time"

	"service-nest/interfaces"
	"service-nest/model"
)

type ServiceRequestRepository struct {
	client *dynamodb.Client
}

func (s ServiceRequestRepository) GetServiceProviderByRequestID(requestID, providerID string) (*model.ServiceRequest, error) {
	//TODO implement me
	panic("implement me")
}

// NewServiceRequestRepository initializes a new ServiceRequestRepository with MySQL
func NewServiceRequestRepository(db *dynamodb.Client) interfaces.ServiceRequestRepository {
	return &ServiceRequestRepository{client: db}
}
func (s *ServiceRequestRepository) SaveServiceRequest(ctx context.Context, request model.ServiceRequest) error {
	requestItem, err := attributevalue.MarshalMap(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	requestItem["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", *request.HouseholderID)}
	requestItem["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("request:pen:%s", request.ID)}
	requestItem["requested_time"] = &types.AttributeValueMemberS{Value: request.RequestedTime.Format(time.RFC3339)}
	requestItem["scheduled_time"] = &types.AttributeValueMemberS{Value: request.ScheduledTime.Format(time.RFC3339)}

	serviceRequestItem, err := attributevalue.MarshalMap(request)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %v", err)
	}
	serviceRequestItem["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("service_request")}
	serviceRequestItem["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", request.ServiceID, request.ID)}
	serviceRequestItem["requested_time"] = &types.AttributeValueMemberS{Value: request.RequestedTime.Format(time.RFC3339)}
	serviceRequestItem["scheduled_time"] = &types.AttributeValueMemberS{Value: request.ScheduledTime.Format(time.RFC3339)}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(config.TABLENAME),
		Item:                requestItem,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}
	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return fmt.Errorf("request already exists with ID %s", request.ID)
		}
		return fmt.Errorf("failed to add service: %v", err)
	}
	input = &dynamodb.PutItemInput{
		TableName:           aws.String(config.TABLENAME),
		Item:                serviceRequestItem,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}
	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return fmt.Errorf("request already exists with ID %s", request.ID)
		}
		return fmt.Errorf("failed to add service: %v", err)
	}
	return nil
}

func (s *ServiceRequestRepository) UpdateServiceRequest(ctx context.Context, updatedRequest *model.ServiceRequest, status string) error {
	updateMainInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", *updatedRequest.HouseholderID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", status, updatedRequest.ID)},
		},
		UpdateExpression: aws.String("SET scheduled_time = :scheduled_time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":scheduled_time": &types.AttributeValueMemberS{Value: updatedRequest.ScheduledTime.Format(time.RFC3339)},
		},
	}

	_, err := s.client.UpdateItem(ctx, updateMainInput)
	if err != nil {
		return fmt.Errorf("fail to update the request: %v", err)
	}

	updateMainInput = &dynamodb.UpdateItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service_request")},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", updatedRequest.ServiceID, updatedRequest.ID)},
		},
		UpdateExpression: aws.String("SET scheduled_time = :scheduled_time"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":scheduled_time": &types.AttributeValueMemberS{Value: updatedRequest.ScheduledTime.Format(time.RFC3339)},
		},
	}

	_, err = s.client.UpdateItem(ctx, updateMainInput)
	if err != nil {
		return fmt.Errorf("fail to update the request: %v", err)
	}
	return nil
}

func (s *ServiceRequestRepository) CancelServiceRequest(ctx context.Context, request *model.ServiceRequest, status string) error {
	// First handle the conditional deletes using TransactWriteItems
	transactItems := []types.TransactWriteItem{
		{
			Delete: &types.Delete{
				TableName: aws.String(config.TABLENAME),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", *request.HouseholderID)},
					"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", status, request.ID)},
				},
				ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
			},
		},
		{
			Delete: &types.Delete{
				TableName: aws.String(config.TABLENAME),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "service_request"},
					"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", request.ServiceID, request.ID)},
				},
			},
		},
	}

	// Execute conditional deletes
	_, err := s.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	})
	if err != nil {
		var txErr *types.TransactionCanceledException
		if errors.As(err, &txErr) {
			return fmt.Errorf("request with ID %s does not exist", request.ID)
		}
		return fmt.Errorf("failed to delete requests: %v", err)
	}

	// Marshal the request for the new cancelled item
	requestItem, err := attributevalue.MarshalMap(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create base cancelled item with all request data
	cancelledItem := make(map[string]types.AttributeValue)
	for k, v := range requestItem {
		cancelledItem[k] = v
	}
	cancelledItem["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", *request.HouseholderID)}
	cancelledItem["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("request:cancl:%s", request.ID)}
	cancelledItem["status"] = &types.AttributeValueMemberS{Value: request.Status}
	cancelledItem["requested_time"] = &types.AttributeValueMemberS{Value: request.RequestedTime.Format(time.RFC3339)}
	cancelledItem["scheduled_time"] = &types.AttributeValueMemberS{Value: request.ScheduledTime.Format(time.RFC3339)}

	// Prepare batch write items
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: cancelledItem,
			},
		},
	}

	// Add update items if request was approved
	if request.ApproveStatus || request.Status == "Approved" {
		// Add updates for both householder and provider
		updateItems := []struct {
			PK string
			SK string
		}{
			{
				PK: fmt.Sprintf("user:%s", *request.HouseholderID),
				SK: fmt.Sprintf("approve:%s:%s", request.ScheduledTime.Format(time.RFC3339), request.ID),
			},
			{
				PK: fmt.Sprintf("user:%s", request.ProviderDetails[0].ServiceProviderID),
				SK: fmt.Sprintf("approve:%s:%s", request.ScheduledTime.Format(time.RFC3339), request.ID),
			},
		}

		for _, item := range updateItems {
			// Create a new map for each approved item with all request data
			approvedItem := make(map[string]types.AttributeValue)
			for k, v := range requestItem {
				approvedItem[k] = v
			}
			approvedItem["PK"] = &types.AttributeValueMemberS{Value: item.PK}
			approvedItem["SK"] = &types.AttributeValueMemberS{Value: item.SK}
			approvedItem["status"] = &types.AttributeValueMemberS{Value: request.Status}
			approvedItem["requested_time"] = &types.AttributeValueMemberS{Value: request.RequestedTime.Format(time.RFC3339)}
			approvedItem["scheduled_time"] = &types.AttributeValueMemberS{Value: request.ScheduledTime.Format(time.RFC3339)}

			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: approvedItem,
				},
			})
		}
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			config.TABLENAME: writeRequests,
		},
	}

	for {
		result, err := s.client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to execute batch write: %v", err)
		}

		if len(result.UnprocessedItems) == 0 {
			break
		}

		// Retry with unprocessed items
		input.RequestItems = result.UnprocessedItems
	}

	return nil
}

func (s *ServiceRequestRepository) GetServiceRequestsByHouseholderID(ctx context.Context, householderID string, limit int, lastEvaluatedKey map[string]types.AttributeValue, status string) ([]model.ServiceRequest, map[string]types.AttributeValue, error) {
	var input *dynamodb.QueryInput
	if status == "" {
		input = &dynamodb.QueryInput{
			TableName:              aws.String(config.TABLENAME),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", householderID)},
				":skPrefix": &types.AttributeValueMemberS{Value: "request:"},
			},
			Limit:             aws.Int32(int32(limit)), // Apply limit
			ExclusiveStartKey: lastEvaluatedKey,        // Pagination cursor
		}
	} else {
		input = &dynamodb.QueryInput{
			TableName:              aws.String(config.TABLENAME),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", householderID)},
				":skPrefix": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:", status)},
			},
			Limit:             aws.Int32(int32(limit)),
			ExclusiveStartKey: lastEvaluatedKey,
		}
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query services: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.ServiceRequest{}, nil, nil
	}

	requests := make([]model.ServiceRequest, 0, len(result.Items))
	for _, item := range result.Items {
		var request model.ServiceRequest
		err = attributevalue.UnmarshalMap(item, &request)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal service: %v", err)
		}
		requests = append(requests, request)
	}

	return requests, result.LastEvaluatedKey, nil
}

func (s *ServiceRequestRepository) GetServiceRequestByID(ctx context.Context, requestID string, householderId string, status string) (*model.ServiceRequest, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", householderId)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", status, requestID)},
		},
	}

	result, err := s.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("request not found")
	}

	var request *model.ServiceRequest
	err = attributevalue.UnmarshalMap(result.Item, &request)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %v", err)
	}

	return request, nil

}

func (s ServiceRequestRepository) GetAllPendingRequestsByProvider(ctx context.Context, providerId string, serviceID string, limit, offset int) ([]model.ServiceRequest, error) {
	var input *dynamodb.QueryInput
	if serviceID == "" {
		input = &dynamodb.QueryInput{
			TableName:              aws.String(config.TABLENAME),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("service_request")},
				":skPrefix": &types.AttributeValueMemberS{Value: "request:"},
			},
		}
	} else {
		input = &dynamodb.QueryInput{
			TableName:              aws.String(config.TABLENAME),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("service_request")},
				":skPrefix": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:", serviceID)},
			},
		}
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.ServiceRequest{}, nil
	}

	requests := make([]model.ServiceRequest, 0, len(result.Items))
	for _, item := range result.Items {
		var request model.ServiceRequest
		var isAccepted bool
		err = attributevalue.UnmarshalMap(item, &request)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal service: %v", err)
		}
		for _, provider := range request.ProviderDetails {
			if provider.ServiceProviderID == providerId {
				isAccepted = true
			}
		}
		if !isAccepted {
			requests = append(requests, request)
		}
	}

	return requests, nil
}

func (s *ServiceRequestRepository) GetServiceRequestByProvider(ctx context.Context, requestID string, serviceId string) (*model.ServiceRequest, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service_request")},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", serviceId, requestID)},
		},
	}

	result, err := s.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("request not found")
	}

	var request *model.ServiceRequest
	err = attributevalue.UnmarshalMap(result.Item, &request)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %v", err)
	}

	return request, nil

}

func (s *ServiceRequestRepository) AcceptServiceRequestByProvider(ctx context.Context, updatedRequest *model.ServiceRequest, status string) error {
	providerDetailsList := make([]types.AttributeValue, len(updatedRequest.ProviderDetails))
	for i, provider := range updatedRequest.ProviderDetails {
		item, err := attributevalue.MarshalMap(provider)
		if err != nil {
			return fmt.Errorf("failed to marshal provider details: %v", err)
		}
		providerDetailsList[i] = &types.AttributeValueMemberM{Value: item}
	}

	updateMainInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("service_request")},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", updatedRequest.ServiceID, updatedRequest.ID)},
		},
		UpdateExpression: aws.String("SET #status = :status , provider_details = :provider_details"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":           &types.AttributeValueMemberS{Value: updatedRequest.Status},
			":provider_details": &types.AttributeValueMemberL{Value: providerDetailsList},
		},
	}

	_, err := s.client.UpdateItem(ctx, updateMainInput)
	if err != nil {
		return fmt.Errorf("fail to update the request: %v", err)
	}

	deleteRequestInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", *updatedRequest.HouseholderID)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", status, updatedRequest.ID)},
		},
	}
	_, err = s.client.DeleteItem(ctx, deleteRequestInput)
	if err != nil {
		return fmt.Errorf("fail to delete request: %v", err)
	}
	status = util.ConvertStatus(updatedRequest.Status)
	requestItem, err := attributevalue.MarshalMap(updatedRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	requestItem["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", *updatedRequest.HouseholderID)}
	requestItem["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("request:%s:%s", status, updatedRequest.ID)}
	requestItem["requested_time"] = &types.AttributeValueMemberS{Value: updatedRequest.RequestedTime.Format(time.RFC3339)}
	requestItem["scheduled_time"] = &types.AttributeValueMemberS{Value: updatedRequest.ScheduledTime.Format(time.RFC3339)}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(config.TABLENAME),
		Item:                requestItem,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}
	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return fmt.Errorf("request already exists with ID %s", updatedRequest.ID)
		}
		return fmt.Errorf("failed to add service: %v", err)
	}
	return nil
}

func (s *ServiceRequestRepository) ApproveServiceRequest(ctx context.Context, serviceRequest *model.ServiceRequest, status string, serviceId string) error {
	// Create batch write request items slice
	var writeRequests []types.WriteRequest

	// Add delete requests
	deleteRequests := []struct {
		PK string
		SK string
	}{
		{
			PK: fmt.Sprintf("service_request"),
			SK: fmt.Sprintf("request:%s:%s", serviceId, serviceRequest.ID),
		},
		{
			PK: fmt.Sprintf("user:%s", *serviceRequest.HouseholderID),
			SK: fmt.Sprintf("request:%s:%s", status, serviceRequest.ID),
		},
	}

	// Add delete requests to batch
	for _, del := range deleteRequests {
		writeRequests = append(writeRequests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: del.PK},
					"SK": &types.AttributeValueMemberS{Value: del.SK},
				},
			},
		})
	}

	// Marshal the service request once for all put operations
	status = util.ConvertStatus(serviceRequest.Status)
	requestItem, err := attributevalue.MarshalMap(serviceRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Add common fields that will be used in all put requests
	requestItem["requested_time"] = &types.AttributeValueMemberS{Value: serviceRequest.RequestedTime.Format(time.RFC3339)}
	requestItem["scheduled_time"] = &types.AttributeValueMemberS{Value: serviceRequest.ScheduledTime.Format(time.RFC3339)}

	// Define put requests
	putRequests := []struct {
		PK string
		SK string
	}{
		{
			PK: fmt.Sprintf("user:%s", *serviceRequest.HouseholderID),
			SK: fmt.Sprintf("request:%s:%s", status, serviceRequest.ID),
		},
		{
			PK: fmt.Sprintf("user:%s", *serviceRequest.HouseholderID),
			SK: fmt.Sprintf("approve:%s:%s", serviceRequest.ScheduledTime.Format(time.RFC3339), serviceRequest.ID),
		},
		{
			PK: fmt.Sprintf("user:%s", serviceRequest.ProviderDetails[0].ServiceProviderID),
			SK: fmt.Sprintf("approve:%s:%s", serviceRequest.ScheduledTime.Format(time.RFC3339), serviceRequest.ID),
		},
	}

	// Add put requests to batch
	for _, put := range putRequests {
		// Create a copy of requestItem for each put request
		putItem := make(map[string]types.AttributeValue)
		for k, v := range requestItem {
			putItem[k] = v
		}
		putItem["PK"] = &types.AttributeValueMemberS{Value: put.PK}
		putItem["SK"] = &types.AttributeValueMemberS{Value: put.SK}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: putItem,
			},
		})
	}

	// Process requests in batches (DynamoDB limit is 25 items per batch)
	const maxBatchSize = 25
	for i := 0; i < len(writeRequests); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(writeRequests) {
			end = len(writeRequests)
		}

		batch := writeRequests[i:end]
		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				config.TABLENAME: batch,
			},
		}

		// Retry logic for unprocessed items
		for {
			result, err := s.client.BatchWriteItem(ctx, input)
			if err != nil {
				return fmt.Errorf("failed to execute batch write: %v", err)
			}

			// Check for unprocessed items
			if len(result.UnprocessedItems) == 0 {
				break
			}

			// Retry with unprocessed items
			input.RequestItems = result.UnprocessedItems
		}
	}

	return nil
}
func (s ServiceRequestRepository) GetApproveServiceRequestsByHouseholderID(ctx context.Context, householderID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error) {
	var scanIndexForward bool

	// Determine sort direction
	switch strings.ToUpper(sortOrder) {
	case "DESC":
		scanIndexForward = false // Latest scheduled time first
	case "ASC":
		scanIndexForward = true // Earliest scheduled time first
	default:
		scanIndexForward = true // Default to ascending order
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(config.TABLENAME),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", householderID)},
			":skPrefix": &types.AttributeValueMemberS{Value: "approve:"},
		},
		ScanIndexForward: aws.Bool(scanIndexForward), // Controls sort order
	}

	// Add limit if specified
	if limit > 0 {
		input.Limit = aws.Int32(int32(limit))
	}

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
		return nil, fmt.Errorf("failed to query services: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.ServiceRequest{}, nil
	}

	requests := make([]model.ServiceRequest, 0, len(result.Items))
	for _, item := range result.Items {
		var request model.ServiceRequest
		err = attributevalue.UnmarshalMap(item, &request)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal service: %v", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func (s *ServiceRequestRepository) GetApproveServiceRequestsByProviderID(ctx context.Context, providerID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error) {
	var scanIndexForward bool

	// Determine sort direction
	switch strings.ToUpper(sortOrder) {
	case "DESC":
		scanIndexForward = false // Latest scheduled time first
	case "ASC":
		scanIndexForward = true // Earliest scheduled time first
	default:
		scanIndexForward = true // Default to ascending order
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(config.TABLENAME),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", providerID)},
			":skPrefix": &types.AttributeValueMemberS{Value: "approve:"},
		},
		ScanIndexForward: aws.Bool(scanIndexForward), // Controls sort order
	}

	// Add limit if specified
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
		return nil, fmt.Errorf("failed to query services: %v", err)
	}

	if len(result.Items) == 0 {
		return []model.ServiceRequest{}, nil
	}

	requests := make([]model.ServiceRequest, 0, len(result.Items))
	for _, item := range result.Items {
		var request model.ServiceRequest
		err = attributevalue.UnmarshalMap(item, &request)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal service: %v", err)
		}
		requests = append(requests, request)
	}

	return requests, nil
}
