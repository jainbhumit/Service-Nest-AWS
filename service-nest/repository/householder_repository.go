package repository

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"service-nest/interfaces"
	"service-nest/model"
)

type MySQLHouseholderRepository struct {
	db *dynamodb.Client
}

func (m MySQLHouseholderRepository) SaveHouseholder(householder *model.Householder) error {
	//TODO implement me
	panic("implement me")
}

// NewHouseholderRepository creates a new instance of MySQLHouseholderRepository
func NewHouseholderRepository(client *dynamodb.Client) interfaces.HouseholderRepository {
	return &MySQLHouseholderRepository{
		db: client,
	}
}
