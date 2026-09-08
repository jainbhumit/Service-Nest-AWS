package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	defaultAWSRegion        = "us-east-1"
	defaultDynamoDBTable    = "servicenest"
	defaultS3Bucket         = "service-nest-category1"
	defaultSNSTopicARN      = "arn:aws:sns:us-east-1:116777895904:Service-Nest"
	defaultSMTPHost         = "smtp.gmail.com"
	defaultSMTPPort         = "587"
	maxGenericEnvLen        = 512
	maxTableNameLen         = 128
	maxRegionLen            = 32
	maxBucketLen            = 128
	maxSMTPPortLen          = 8
	maxAppEnvLen            = 16
)

// EnvConfig holds runtime configuration loaded from environment variables.
type EnvConfig struct {
	AppEnv           string
	AWSRegion        string
	DynamoDBTable    string
	DynamoDBEndpoint string
	S3Bucket         string
	S3Endpoint       string
	SNSTopicARN      string
	JWTSecret        string
	SMTPHost         string
	SMTPPort         string
	SMTPFrom         string
	SMTPAppPassword  string
}

var current *EnvConfig

// Legacy exports used by repositories and utilities. Set by Load().
var (
	TABLENAME string
	BUCKET    string
	REGION    string
	SNSARN    string
)

// Load reads configuration from the environment, validates it, and updates legacy exports.
func Load() (*EnvConfig, error) {
	cfg := &EnvConfig{
		AppEnv:           inferAppEnv(),
		AWSRegion:        envString("AWS_REGION", defaultAWSRegion, maxRegionLen),
		DynamoDBTable:    envString("DYNAMODB_TABLE", defaultDynamoDBTable, maxTableNameLen),
		DynamoDBEndpoint: envString("DYNAMODB_ENDPOINT", "", maxGenericEnvLen),
		S3Bucket:         envString("S3_BUCKET", defaultS3Bucket, maxBucketLen),
		S3Endpoint:       envString("S3_ENDPOINT", "", maxGenericEnvLen),
		JWTSecret:        firstNonEmptyEnv([]string{"JWT_SECRET", "SECRET"}, maxGenericEnvLen),
		SMTPHost:         envString("SMTP_HOST", defaultSMTPHost, maxGenericEnvLen),
		SMTPPort:         envString("SMTP_PORT", defaultSMTPPort, maxSMTPPortLen),
		SMTPFrom:         envString("SMTP_FROM", "", maxGenericEnvLen),
		SMTPAppPassword:  firstNonEmptyEnv([]string{"SMTP_APP_PASSWORD", "APP_PASSWORD"}, maxGenericEnvLen),
	}

	cfg.SNSTopicARN = resolveSNSTopicARN(cfg.AppEnv)

	if err := validate(cfg); err != nil {
		return nil, err
	}

	applyLegacyExports(cfg)
	current = cfg
	return cfg, nil
}

// Current returns the last successfully loaded configuration.
func Current() *EnvConfig {
	return current
}

func inferAppEnv() string {
	if explicit := strings.TrimSpace(os.Getenv("APP_ENV")); explicit != "" {
		return envString("APP_ENV", explicit, maxAppEnvLen)
	}
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		return "prod"
	}
	return "local"
}

func resolveSNSTopicARN(appEnv string) string {
	if explicit := strings.TrimSpace(os.Getenv("SNS_TOPIC_ARN")); explicit != "" {
		return truncateEnv(explicit, maxGenericEnvLen)
	}
	if appEnv == "local" {
		return ""
	}
	return defaultSNSTopicARN
}

func applyLegacyExports(cfg *EnvConfig) {
	TABLENAME = cfg.DynamoDBTable
	BUCKET = cfg.S3Bucket
	REGION = cfg.AWSRegion
	SNSARN = cfg.SNSTopicARN
}

func validate(cfg *EnvConfig) error {
	if cfg == nil {
		return errors.New("config must not be nil")
	}
	if cfg.AWSRegion == "" {
		return errors.New("AWS_REGION must not be empty")
	}
	if cfg.DynamoDBTable == "" {
		return errors.New("DYNAMODB_TABLE must not be empty")
	}
	if cfg.S3Bucket == "" {
		return errors.New("S3_BUCKET must not be empty")
	}
	if cfg.AppEnv == "prod" && cfg.JWTSecret == "" {
		return errors.New("JWT_SECRET or SECRET must be set in prod")
	}
	if cfg.AppEnv == "local" && cfg.JWTSecret == "" {
		return errors.New("JWT_SECRET or SECRET must be set for local development")
	}
	return nil
}

func envString(key, defaultVal string, maxLen int) string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultVal
	}
	return truncateEnv(raw, maxLen)
}

func firstNonEmptyEnv(keys []string, maxLen int) string {
	for _, key := range keys {
		if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
			return truncateEnv(raw, maxLen)
		}
	}
	return ""
}

func truncateEnv(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(value) > maxLen {
		return value[:maxLen]
	}
	return value
}

// IsLocal reports whether the application runs in local mode.
func (c *EnvConfig) IsLocal() bool {
	if c == nil {
		return false
	}
	return c.AppEnv == "local"
}

// SMTPConfigured reports whether outbound email can be attempted.
func (c *EnvConfig) SMTPConfigured() bool {
	if c == nil {
		return false
	}
	return c.SMTPAppPassword != "" && c.SMTPFrom != "" && c.SMTPHost != "" && c.SMTPPort != ""
}

// String returns a redacted summary for debugging (no secrets).
func (c *EnvConfig) String() string {
	if c == nil {
		return "config(nil)"
	}
	return fmt.Sprintf(
		"app_env=%s region=%s table=%s sns_configured=%t smtp_configured=%t",
		c.AppEnv,
		c.AWSRegion,
		c.DynamoDBTable,
		c.SNSTopicARN != "",
		c.SMTPConfigured(),
	)
}
