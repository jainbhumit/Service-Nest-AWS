package util

import (
	"context"
	"fmt"
	"net/url"
	"service-nest/config"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GeneratePresignedURL(ctx context.Context, fileName string) (string, string, error) {
	if fileName == "" {
		return "", "", fmt.Errorf("filename cannot be empty")
	}

	client, err := config.S3Client(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to create S3 client: %w", err)
	}

	bucketName := config.BUCKET
	key := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), fileName)

	presignClient := s3.NewPresignClient(client, func(po *s3.PresignOptions) {
		po.Expires = 15 * time.Minute
	})

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		ContentType: aws.String("image/png"),
	}

	req, err := presignClient.PresignPutObject(ctx, input)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		bucketName,
		config.REGION,
		key,
	)

	envCfg := config.Current()
	if envCfg != nil && envCfg.S3Endpoint != "" {
		objectURL = fmt.Sprintf("%s/%s/%s", strings.TrimRight(envCfg.S3Endpoint, "/"), bucketName, key)
	}

	return req.URL, objectURL, nil
}

func DeleteFileFromS3(fileURL string) error {
	client, err := config.S3Client(context.TODO())
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	bucketName := config.BUCKET
	key := strings.TrimPrefix(parsedURL.Path, "/")

	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	return err
}
