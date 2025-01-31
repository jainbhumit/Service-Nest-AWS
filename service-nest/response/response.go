package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	ErrorCode int         `json:"error_code,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// SuccessResponse generates a standard success response
func SuccessResponse(w http.ResponseWriter, data interface{}, message string, code int) {
	response := Response{
		Status:  "Success",
		Message: message,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// ErrorResponse generates a standard error response
func ErrorResponse(w http.ResponseWriter, statusCode int, errMessage string, code int) {
	response := Response{
		Status:    "Fail",
		ErrorCode: code,
		Message:   errMessage,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
