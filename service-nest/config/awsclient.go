package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// Clients holds AWS service clients configured for the current environment.
type Clients struct {
	DynamoDB *dynamodb.Client
	S3       *s3.Client
	SNS      *sns.Client
}

var activeClients *Clients

// NewClients builds AWS clients from EnvConfig (DynamoDB Local / MinIO when endpoints are set).
func NewClients(cfg *EnvConfig) (*Clients, error) {
	if cfg == nil {
		return nil, fmt.Errorf("env config must not be nil")
	}

	awsCfg, err := loadAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	clients := &Clients{
		DynamoDB: newDynamoDBClient(awsCfg, cfg),
		S3:       newS3Client(awsCfg, cfg),
		SNS:      sns.NewFromConfig(awsCfg),
	}

	activeClients = clients
	return clients, nil
}

// ActiveClients returns clients created by the most recent NewClients call.
func ActiveClients() *Clients {
	return activeClients
}

func loadAWSConfig(cfg *EnvConfig) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.AWSRegion),
	}

	if cfg.DynamoDBEndpoint != "" || cfg.S3Endpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("local", "local", ""),
		))
	}

	return awsconfig.LoadDefaultConfig(context.Background(), opts...)
}

func newDynamoDBClient(awsCfg aws.Config, cfg *EnvConfig) *dynamodb.Client {
	if cfg.DynamoDBEndpoint == "" {
		return dynamodb.NewFromConfig(awsCfg)
	}

	endpoint := cfg.DynamoDBEndpoint
	return dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
}

func newS3Client(awsCfg aws.Config, cfg *EnvConfig) *s3.Client {
	if cfg.S3Endpoint == "" {
		return s3.NewFromConfig(awsCfg)
	}

	endpoint := cfg.S3Endpoint
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

// S3Client returns the shared S3 client or builds one from the loaded env config.
func S3Client(ctx context.Context) (*s3.Client, error) {
	if activeClients != nil && activeClients.S3 != nil {
		return activeClients.S3, nil
	}

	cfg := Current()
	if cfg == nil {
		return nil, fmt.Errorf("application config is not loaded")
	}

	clients, err := NewClients(cfg)
	if err != nil {
		return nil, err
	}
	return clients.S3, nil
}
