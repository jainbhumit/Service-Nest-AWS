package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"service-nest/logger"
	"service-nest/model"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type ErrorNotifier struct {
	SnsClient *sns.Client
	TopicArn  string
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	buffer     *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		buffer:         bytes.NewBuffer([]byte{}),
	}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.buffer.Write(b)
	return rw.ResponseWriter.Write(b)
}

func LoggingMiddleware(notifier *model.ErrorNotifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			// Set CORS headers for preflight requests
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			rw.Header().Set("Access-Control-Allow-Headers", "*")

			if r.Method == http.MethodOptions {
				rw.WriteHeader(http.StatusOK)
				return
			}

			// Recover from panics
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()

					// Set 500 status code
					rw.WriteHeader(http.StatusInternalServerError)

					// Create error response
					errorResponse := map[string]interface{}{
						"status":     "Fail",
						"message":    "Internal server error",
						"error_code": 1500,
					}

					json.NewEncoder(rw).Encode(errorResponse)

					// Prepare error details for notification
					errorDetails := map[string]interface{}{
						"method":       r.Method,
						"path":         r.URL.Path,
						"statusCode":   http.StatusInternalServerError,
						"requestID":    r.Header.Get("X-Amzn-Request-Id"),
						"duration":     time.Since(start).String(),
						"sourceIP":     r.Header.Get("X-Forwarded-For"),
						"responseBody": errorResponse,
						"error":        fmt.Sprintf("%v", err),
					}

					// Send notification
					if notifier != nil {
						if err := notifier.NotifyError(errorDetails, stack); err != nil {
							logger.Error("Failed to send SNS notification", map[string]interface{}{
								"error": err.Error(),
							})
						}
					}

					// Log the error
					logger.Error("Panic recovered", errorDetails)
				}
			}()

			// Read and restore the request body
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// Extract request headers
			headers := make(map[string]string)
			for key, values := range r.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}

			// Log request
			requestFields := map[string]interface{}{
				"event":     "request",
				"method":    r.Method,
				"path":      r.URL.Path,
				"query":     r.URL.RawQuery,
				"headers":   headers,
				"sourceIP":  r.Header.Get("X-Forwarded-For"),
				"userAgent": r.Header.Get("User-Agent"),
				"requestID": r.Header.Get("X-Amzn-Request-Id"),
				"stage":     r.Header.Get("X-Amzn-Api-Stage"),
			}

			// Add request body if it exists and is JSON
			if len(bodyBytes) > 0 {
				var jsonBody interface{}
				if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
					requestFields["body"] = jsonBody
				}
			}

			logger.Info("API Request", requestFields)

			// Process the request
			next.ServeHTTP(rw, r)

			// Prepare response logging
			responseFields := map[string]interface{}{
				"event":      "response",
				"method":     r.Method,
				"path":       r.URL.Path,
				"statusCode": rw.statusCode,
				"duration":   time.Since(start).String(),
				"durationMs": time.Since(start).Milliseconds(),
				"requestID":  r.Header.Get("X-Amzn-Request-Id"),
			}

			// Try to parse response body if it's JSON
			var jsonResponse interface{}
			if rw.buffer.Len() > 0 {
				if err := json.Unmarshal(rw.buffer.Bytes(), &jsonResponse); err == nil {
					responseFields["responseBody"] = jsonResponse
				}
			}

			// Check for 500 errors and notify if necessary
			if rw.statusCode == http.StatusInternalServerError && notifier != nil {
				stack := debug.Stack()
				if err := notifier.NotifyError(responseFields, stack); err != nil {
					logger.Error("Failed to send SNS notification", map[string]interface{}{
						"error": err.Error(),
					})
				}
			}

			// Log response with appropriate level based on status code
			if rw.statusCode >= 400 {
				logger.Error("API Response", responseFields)
			} else {
				logger.Info("API Response", responseFields)
			}
		})
	}
}
