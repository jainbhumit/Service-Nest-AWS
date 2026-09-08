package config

import (
	"strings"
	"testing"
)

func TestNewClientsWithLocalEndpoints(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("APP_ENV", "local")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")
	t.Setenv("S3_ENDPOINT", "http://localhost:9000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	clients, err := NewClients(cfg)
	if err != nil {
		t.Fatalf("NewClients() error = %v", err)
	}
	if clients.DynamoDB == nil || clients.S3 == nil || clients.SNS == nil {
		t.Fatal("expected all AWS clients to be initialized")
	}
}

func TestS3ClientRequiresLoadedConfig(t *testing.T) {
	activeClients = nil
	current = nil

	_, err := S3Client(t.Context())
	if err == nil || !strings.Contains(err.Error(), "not loaded") {
		t.Fatalf("S3Client() error = %v, want config not loaded", err)
	}
}
