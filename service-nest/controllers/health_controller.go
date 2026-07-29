package controllers

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"service-nest/config"
	"service-nest/response"
)

// HealthController exposes liveness and readiness endpoints.
type HealthController struct {
	dynamo *dynamodb.Client
}

func NewHealthController(dynamo *dynamodb.Client) *HealthController {
	return &HealthController{dynamo: dynamo}
}

// Liveness reports whether the process is running.
func (h *HealthController) Liveness(w http.ResponseWriter, r *http.Request) {
	response.SuccessResponse(w, map[string]string{"status": "ok"}, "ok", http.StatusOK)
}

// Readiness reports whether required dependencies (DynamoDB) are reachable.
func (h *HealthController) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.dynamo == nil {
		response.ErrorResponse(w, http.StatusServiceUnavailable, "dynamodb client not configured", 1006)
		return
	}

	tableName := config.TABLENAME
	if tableName == "" {
		response.ErrorResponse(w, http.StatusServiceUnavailable, "dynamodb table not configured", 1006)
		return
	}

	_, err := h.dynamo.DescribeTable(r.Context(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		response.ErrorResponse(w, http.StatusServiceUnavailable, "dynamodb unavailable", 1006)
		return
	}

	response.SuccessResponse(w, map[string]string{"status": "ready"}, "ok", http.StatusOK)
}
