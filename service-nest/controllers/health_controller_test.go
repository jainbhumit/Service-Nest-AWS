package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"service-nest/config"
)

func TestHealthController_Liveness(t *testing.T) {
	controller := NewHealthController(nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	controller.Liveness(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestHealthController_ReadinessWithoutClient(t *testing.T) {
	controller := NewHealthController(nil)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	controller.Readiness(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestHealthController_ReadinessWithoutTableName(t *testing.T) {
	originalTable := config.TABLENAME
	config.TABLENAME = ""
	t.Cleanup(func() {
		config.TABLENAME = originalTable
	})

	controller := NewHealthController(nil)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	controller.Readiness(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}
