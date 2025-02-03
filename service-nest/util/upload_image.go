package util

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/url"
	config2 "service-nest/config"
	"strings"
	"time"
)

func GeneratePresignedURL(ctx context.Context, fileName string) (string, string, error) {
	if fileName == "" {
		return "", "", fmt.Errorf("filename cannot be empty")
	}

	// Load AWS Config with specific options
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(config2.REGION),
		config.WithDefaultsMode(aws.DefaultsModeInRegion),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Initialize S3 client
	client := s3.NewFromConfig(cfg)

	bucketName := config2.BUCKET
	key := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), fileName)

	// Create the presign client
	presignClient := s3.NewPresignClient(client, func(po *s3.PresignOptions) {
		po.Expires = 15 * time.Minute
	})

	// Generate presigned URL with specific options
	input := &s3.PutObjectInput{
		Bucket:      aws.String(config2.BUCKET),
		Key:         aws.String(key),
		ContentType: aws.String("image/png"), // Add content type
	}

	req, err := presignClient.PresignPutObject(ctx, input)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Construct the public object URL
	objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		bucketName,
		config2.REGION,
		key,
	)

	return req.URL, objectURL, nil
}

func DeleteFileFromS3(fileURL string) error {
	// Load AWS Config (default)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(config2.REGION))
	if err != nil {
		return err
	}

	// Initialize S3 Service
	svc := s3.NewFromConfig(cfg)

	// Parse the URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract the bucket and key
	bucketName := config2.BUCKET
	key := strings.TrimPrefix(parsedURL.Path, "/") // Remove the leading slash

	// Perform the deletion
	_, err = svc.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	return err
}
