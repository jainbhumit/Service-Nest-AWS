package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"service-nest/logger"
	"time"
)

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

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		// Set CORS headers for preflight requests
		rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		rw.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			rw.WriteHeader(http.StatusOK)
			return
		}

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
		if rw.buffer.Len() > 0 {
			var jsonResponse interface{}
			if err := json.Unmarshal(rw.buffer.Bytes(), &jsonResponse); err == nil {
				responseFields["responseBody"] = jsonResponse
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
