package config

import (
	"os"
	"testing"
)

func TestLoadLocalDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "local")
	t.Setenv("JWT_SECRET", "local-dev-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "local" {
		t.Errorf("AppEnv = %q, want local", cfg.AppEnv)
	}
	if cfg.AWSRegion != defaultAWSRegion {
		t.Errorf("AWSRegion = %q, want %q", cfg.AWSRegion, defaultAWSRegion)
	}
	if cfg.DynamoDBTable != defaultDynamoDBTable {
		t.Errorf("DynamoDBTable = %q, want %q", cfg.DynamoDBTable, defaultDynamoDBTable)
	}
	if cfg.SNSTopicARN != "" {
		t.Errorf("SNSTopicARN = %q, want empty for local", cfg.SNSTopicARN)
	}
	if TABLENAME != cfg.DynamoDBTable {
		t.Errorf("legacy TABLENAME = %q, want %q", TABLENAME, cfg.DynamoDBTable)
	}
}

func TestLoadProdRequiresJWTSecret(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "prod")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when JWT secret missing in prod")
	}
}

func TestLoadProdWithLegacySecretEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "prod")
	t.Setenv("SECRET", "prod-secret-from-legacy-env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.JWTSecret != "prod-secret-from-legacy-env" {
		t.Errorf("JWTSecret = %q, want legacy SECRET value", cfg.JWTSecret)
	}
	if cfg.SNSTopicARN != defaultSNSTopicARN {
		t.Errorf("SNSTopicARN = %q, want default prod ARN", cfg.SNSTopicARN)
	}
}

func TestLoadOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "local")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("DYNAMODB_TABLE", "custom-table")
	t.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")
	t.Setenv("S3_BUCKET", "custom-bucket")
	t.Setenv("SMTP_APP_PASSWORD", "smtp-pass")
	t.Setenv("SMTP_FROM", "dev@example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AWSRegion != "eu-west-1" {
		t.Errorf("AWSRegion = %q", cfg.AWSRegion)
	}
	if cfg.DynamoDBTable != "custom-table" {
		t.Errorf("DynamoDBTable = %q", cfg.DynamoDBTable)
	}
	if cfg.DynamoDBEndpoint != "http://localhost:8000" {
		t.Errorf("DynamoDBEndpoint = %q", cfg.DynamoDBEndpoint)
	}
	if cfg.S3Bucket != "custom-bucket" {
		t.Errorf("S3Bucket = %q", cfg.S3Bucket)
	}
	if !cfg.SMTPConfigured() {
		t.Error("SMTPConfigured() = false, want true")
	}
}

func TestLoadInfersProdFromLambdaEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "HelloWorldFunction")
	t.Setenv("JWT_SECRET", "lambda-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppEnv != "prod" {
		t.Errorf("AppEnv = %q, want prod", cfg.AppEnv)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"APP_ENV",
		"AWS_REGION",
		"AWS_LAMBDA_FUNCTION_NAME",
		"DYNAMODB_TABLE",
		"DYNAMODB_ENDPOINT",
		"S3_BUCKET",
		"S3_ENDPOINT",
		"SNS_TOPIC_ARN",
		"JWT_SECRET",
		"SECRET",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_FROM",
		"SMTP_APP_PASSWORD",
		"APP_PASSWORD",
	}
	for _, key := range keys {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
	current = nil
	TABLENAME = ""
	BUCKET = ""
	REGION = ""
	SNSARN = ""
}
