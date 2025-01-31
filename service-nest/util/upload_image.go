package util

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"mime/multipart"
	"net/url"
	config2 "service-nest/config"
	"strings"
	"time"
)

func UploadFileToS3(file multipart.File, fileName string) (string, error) {
	// Load AWS Config (default)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(config2.REGION))
	if err != nil {
		return "", err
	}

	//Initialise S3 Service
	svc := s3.NewFromConfig(cfg)
	bucketName := config2.BUCKET
	key := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), fileName)

	// Upload the file to S3
	_, err = svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Return the S3 URL of the uploaded file
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, key), nil
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
