package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"service-nest/config"
	"service-nest/errs"
	"service-nest/interfaces"
	"service-nest/model"
	"strings"
)

type UserRepository struct {
	client *dynamodb.Client
}

func NewUserRepository(db *dynamodb.Client) interfaces.UserRepository {
	return &UserRepository{client: db}
}

func (u *UserRepository) SaveUser(ctx context.Context, user *model.User) error {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(config.TABLENAME),
		Item: map[string]types.AttributeValue{
			"PK":        &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", user.Email)},
			"SK":        &types.AttributeValueMemberS{Value: user.Email},
			"id":        &types.AttributeValueMemberS{Value: user.ID},
			"name":      &types.AttributeValueMemberS{Value: user.Name},
			"email":     &types.AttributeValueMemberS{Value: user.Email},
			"password":  &types.AttributeValueMemberS{Value: user.Password},
			"role":      &types.AttributeValueMemberS{Value: user.Role},
			"address":   &types.AttributeValueMemberS{Value: user.Address},
			"contact":   &types.AttributeValueMemberS{Value: user.Contact},
			"is_active": &types.AttributeValueMemberBOOL{Value: user.IsActive},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	_, err := u.client.PutItem(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf(errs.EmailAlreadyUse)
		}
		return fmt.Errorf("%s: %v", errs.FailToCreateUser, err)
	}
	input = &dynamodb.PutItemInput{
		TableName: aws.String(config.TABLENAME),
		Item: map[string]types.AttributeValue{
			"PK":        &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", user.ID)},
			"SK":        &types.AttributeValueMemberS{Value: user.ID},
			"id":        &types.AttributeValueMemberS{Value: user.ID},
			"name":      &types.AttributeValueMemberS{Value: user.Name},
			"email":     &types.AttributeValueMemberS{Value: user.Email},
			"password":  &types.AttributeValueMemberS{Value: user.Password},
			"role":      &types.AttributeValueMemberS{Value: user.Role},
			"address":   &types.AttributeValueMemberS{Value: user.Address},
			"contact":   &types.AttributeValueMemberS{Value: user.Contact},
			"is_active": &types.AttributeValueMemberBOOL{Value: user.IsActive},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	_, err = u.client.PutItem(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf(errs.EmailAlreadyUse)
		}
		return fmt.Errorf("%s: %v", errs.FailToCreateUser, err)
	}

	return nil
}

func (u *UserRepository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", userID)},
			"SK": &types.AttributeValueMemberS{Value: userID},
		},
	}

	result, err := u.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf(errs.UserNotFound)
	}

	var user *model.User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %v", err)
	}

	return user, nil
}

func (u *UserRepository) UpdateUser(ctx context.Context, updatedUser *model.User, oldEmail string) error {
	updateMainInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", updatedUser.ID)},
			"SK": &types.AttributeValueMemberS{Value: updatedUser.ID},
		},
		UpdateExpression: aws.String("SET email = :email, password = :password, address = :address, contact = :contact"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email":    &types.AttributeValueMemberS{Value: updatedUser.Email},
			":address":  &types.AttributeValueMemberS{Value: updatedUser.Address},
			":contact":  &types.AttributeValueMemberS{Value: updatedUser.Contact},
			":password": &types.AttributeValueMemberS{Value: updatedUser.Password},
		},
		ConditionExpression: aws.String("attribute_exists(PK)"),
	}

	_, err := u.client.UpdateItem(ctx, updateMainInput)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf(errs.UserNotFound)
		}
		return fmt.Errorf("%s: %v", errs.FailToUpdateUser, err)
	}

	deleteOldEmailInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(config.TABLENAME),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", oldEmail)},
			"SK": &types.AttributeValueMemberS{Value: oldEmail},
		},
	}

	_, err = u.client.DeleteItem(ctx, deleteOldEmailInput)
	if err != nil {
		return fmt.Errorf("failed to delete old email entry: %v", err)
	}

	// Step 3: Create a new email entry (user:newEmail)
	putNewEmailInput := &dynamodb.PutItemInput{
		TableName: aws.String(config.TABLENAME),
		Item: map[string]types.AttributeValue{
			"PK":        &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", updatedUser.Email)},
			"SK":        &types.AttributeValueMemberS{Value: updatedUser.Email},
			"id":        &types.AttributeValueMemberS{Value: updatedUser.ID},
			"name":      &types.AttributeValueMemberS{Value: updatedUser.Name},
			"email":     &types.AttributeValueMemberS{Value: updatedUser.Email},
			"password":  &types.AttributeValueMemberS{Value: updatedUser.Password},
			"role":      &types.AttributeValueMemberS{Value: updatedUser.Role},
			"address":   &types.AttributeValueMemberS{Value: updatedUser.Address},
			"contact":   &types.AttributeValueMemberS{Value: updatedUser.Contact},
			"is_active": &types.AttributeValueMemberBOOL{Value: updatedUser.IsActive},
		},
	}

	_, err = u.client.PutItem(ctx, putNewEmailInput)
	if err != nil {
		return fmt.Errorf("failed to create new email entry: %v", err)
	}

	return nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(config.TABLENAME), // Replace with your DynamoDB table name
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("user:%s", email)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("%s", email)},
		},
	}

	// Fetch the user from DynamoDB
	result, err := u.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %v", err)
	}

	// Check if the user exists
	if result.Item == nil {
		return nil, fmt.Errorf(errs.UserNotFound)
	}

	// Define a struct to match DynamoDB item structure
	var user *model.User

	// Unmarshal the DynamoDB item
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %v", err)
	}

	return user, nil
}

func (u UserRepository) DeActivateUser(userID string) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepository) GetSecurityAnswerByEmail(userEmail string) (*string, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepository) UpdatePassword(userEmail, updatedPassword string) error {
	//TODO implement me
	panic("implement me")
}
