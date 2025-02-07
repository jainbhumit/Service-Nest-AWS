package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"net/http"
	"service-nest/errs"
	"service-nest/interfaces"
	"service-nest/logger"
	"service-nest/model"
	"service-nest/response"
	"service-nest/util"
	"time"
)

type HouseholderController struct {
	householderService interfaces.HouseholderService
}

func NewHouseholderController(householderService interfaces.HouseholderService) *HouseholderController {
	return &HouseholderController{
		householderService: householderService,
	}
}

func (h *HouseholderController) GetAvailableServices(w http.ResponseWriter, r *http.Request) {
	categoryId := r.URL.Query().Get("category_id")
	if categoryId == "" {
		logger.Error("No query param for category", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "request status is required", 2001)
		return
	}
	ctx := r.Context()
	logger.Info(fmt.Sprintf("category id is %v", categoryId), nil)
	services, err := h.householderService.GetServicesByCategoryId(ctx, categoryId)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "error fetching services", 1006)
		return
	}
	response.SuccessResponse(w, services, "Available services", http.StatusOK)
}

func (h *HouseholderController) RequestService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request struct {
		ServiceName   string `json:"service_name" validate:"required"`
		Category      string `json:"category" validate:"required"`
		Description   string `json:"description" validate:"required"`
		ScheduledTime string `json:"scheduled_time" validate:"required"`
		CategoryId    string `json:"category_id" validate:"required"`
	}

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error("Invalid input", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input", 1001)
		return
	}

	// Validate the request body
	err := validate.Struct(request)
	if err != nil {
		logger.Error("Invalid request body", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}
	role := r.Context().Value("role").(string)
	var householderID string
	if role == "Admin" {
		householderID = r.URL.Query().Get("user_id")
		if householderID == "" {
			logger.Error("No query param", nil)
			response.ErrorResponse(w, http.StatusBadRequest, "user ID is required", 2001)
			return
		}
	} else if role == "Householder" {
		householderID = r.Context().Value("userID").(string)
	} else {
		logger.Error("Invalid role", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid role", 1007)
		return
	}
	// Example of retrieving the householder ID from the context

	scheduleTime, err := time.Parse("2006-01-02 15:04", request.ScheduledTime)
	if err != nil {
		logger.Error("Invalid request body", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid time format", 1001)
		return
	}
	houseHolderRequest := &model.ServiceRequest{
		ServiceName:   request.ServiceName,
		ServiceID:     request.CategoryId,
		Description:   request.Description,
		ScheduledTime: scheduleTime,
		HouseholderID: &householderID,
	}
	// Pass the request's ServiceName and ScheduledTime to the service layer
	requestID, err := h.householderService.RequestService(ctx, *houseHolderRequest)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "error requesting service", 1006)
		return
	}

	logger.Info("Service request successfully", nil)
	var respone struct {
		ID string `json:"request_id"`
	}
	respone.ID = requestID
	response.SuccessResponse(w, respone, "Service request successfully", http.StatusCreated)
}

func (h *HouseholderController) CancelServiceRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	requestID, ok := vars["request_id"]
	if !ok {
		logger.Error("Missing request Id in params", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Missing request Id in params", 2002)
		return
	}
	status := r.URL.Query().Get("status")
	status = util.ConvertStatus(status)
	if status == "" {
		logger.Error("No query param for status", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "request status is required", 2001)
		return
	}
	role := r.Context().Value("role").(string)
	var householderID string
	if role == "Admin" {
		householderID = r.URL.Query().Get("user_id")
		if householderID == "" {
			logger.Error("No query param", nil)
			response.ErrorResponse(w, http.StatusBadRequest, "user ID is required", 2001)
			return
		}
	} else if role == "Householder" {
		householderID = r.Context().Value("userID").(string)
	} else {
		logger.Error("Invalid role", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid role", 1007)
		return
	}
	err := h.householderService.CancelServiceRequest(ctx, requestID, householderID, status)
	if err != nil {
		if err.Error() == errs.RequestCancellationTooLate {
			response.SuccessResponse(w, nil, err.Error(), http.StatusOK)
		} else {
			logger.Error(fmt.Sprintf("Error cancelling request %v", err), nil)
			response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1006)
		}
		return
	}
	logger.Info("Request cancelled successfully", nil)
	response.SuccessResponse(w, nil, "Request cancelled successfully", http.StatusOK)
	//color.Green("Service request %s has been successfully canceled.", requestID)
}

func (h *HouseholderController) RescheduleServiceRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID            string `json:"id" validate:"required"`
		ScheduledTime string `json:"scheduled_time" validate:"required"`
	}
	ctx := r.Context()
	status := r.URL.Query().Get("status")
	status = util.ConvertStatus(status)
	if status == "" {
		logger.Error("No query param for status", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "request status is required", 2001)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error("Invalid input", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input", 1001)
		return
	}
	err := validate.Struct(request)
	if err != nil {
		logger.Error("Invalid request body", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	newTime, err := time.Parse("2006-01-02 15:04", request.ScheduledTime)
	if err != nil {
		logger.Error("Invalid request body", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid time format", 1001)
		return
	}
	role := r.Context().Value("role").(string)
	var householderID string
	if role == "Admin" {
		householderID = r.URL.Query().Get("user_id")
		if householderID == "" {
			logger.Error("No query param", nil)
			response.ErrorResponse(w, http.StatusBadRequest, "user ID is required", 2001)
			return
		}
	} else if role == "Householder" {
		householderID = r.Context().Value("userID").(string)
	} else {
		logger.Error("Invalid role", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid role", 1007)
		return
	}
	err = h.householderService.RescheduleServiceRequest(ctx, request.ID, newTime, householderID, status)
	if err != nil {
		logger.Error(fmt.Sprintf("Error rescheduling service %v", err), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1008)
		//color.Red("Error rescheduling service_test request: %v", err)
		return
	}
	logger.Info("Successfully rescheduled service request", nil)
	response.SuccessResponse(w, nil, "service request has been successfully rescheduled", http.StatusOK)
	//color.Green("Service request %s has been successfully rescheduled.", requestID)
}
func (h *HouseholderController) ViewBookingHistory(w http.ResponseWriter, r *http.Request) {

	var request struct {
		Key string `json:"start_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error("Invalid input", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input", 1001)
		return
	}
	ctx := r.Context()
	role := r.Context().Value("role").(string)
	var householderID string
	if role == "Admin" {
		householderID = r.URL.Query().Get("user_id")
		if householderID == "" {
			logger.Error("No query param", nil)
			response.ErrorResponse(w, http.StatusBadRequest, "user ID is required", 2001)
			return
		}
	} else if role == "Householder" {
		householderID = r.Context().Value("userID").(string)
	} else {
		logger.Error("Invalid role", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid role", 1007)
		return
	}

	limit, _ := util.GetPaginationParams(r)
	var lastKey map[string]types.AttributeValue
	if request.Key != "" {
		// Unmarshal the JSON string into a map
		err := json.Unmarshal([]byte(request.Key), &lastKey)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to unmarshal lastEvaluatedKey %v", err.Error()), nil)
			response.ErrorResponse(w, http.StatusBadRequest, "Invalid lastEvaluatedKey format", 2001)
			return
		}
	}
	logger.Info("Last Evaluated Key is ", map[string]interface{}{"key": lastKey})
	status := util.GetFilterParam(r, "status")
	status = util.ConvertStatus(status)

	serviceRequests, lastKey, err := h.householderService.ViewStatus(ctx, householderID, limit, lastKey, status)
	if err != nil {
		logger.Error("Failed to fetch service requests", map[string]interface{}{
			"householderID": householderID,
			"error":         err.Error(),
		})
		response.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch service requests", 1003)
		return
	}

	// Check if there are any service requests
	if len(serviceRequests) == 0 {
		logger.Info("No service requests found for householder", map[string]interface{}{
			"householderID": householderID,
		})
		response.SuccessResponse(w, nil, "No service request found", http.StatusOK)
		return
	}

	type responseStruct struct {
		ID              string                         `json:"request_id"`
		ServiceName     string                         `json:"service_name,omitempty"`
		ServiceID       string                         `json:"service_id" `
		RequestedTime   time.Time                      `json:"requested_time"`
		ScheduledTime   time.Time                      `json:"scheduled_time"`
		Status          string                         `json:"status"` // Pending, Accepted, Approved, Cancelled
		ApproveStatus   bool                           `json:"approve_status" bson:"approveStatus"`
		ProviderDetails []model.ServiceProviderDetails `json:"provider_details,omitempty" bson:"providerDetails,omitempty"`
	}
	responseBody := make([]responseStruct, 0)
	for _, request := range serviceRequests {
		currRequest := &responseStruct{
			ID:            request.ID,
			ServiceName:   request.ServiceName,
			ServiceID:     request.ServiceID,
			RequestedTime: request.RequestedTime,
			ScheduledTime: request.ScheduledTime,
			Status:        request.Status,
		}
		if request.Status == "Accepted" && request.ProviderDetails != nil && !request.ApproveStatus {
			for _, provider := range request.ProviderDetails {
				currRequest.ProviderDetails = append(currRequest.ProviderDetails, model.ServiceProviderDetails{
					ServiceProviderID: provider.ServiceProviderID,
					Name:              provider.Name,
					Contact:           provider.Contact,
					Address:           provider.Address,
					Price:             provider.Price,
					Rating:            provider.Rating,
				})
			}
		}
		responseBody = append(responseBody, *currRequest)
	}

	//// Handle the lastEvaluatedKey for the response
	//var nextKeyString string
	//if lastKey != nil {
	//	keyBytes, err := json.Marshal(lastKey)
	//	if err != nil {
	//		logger.Error(fmt.Sprintf("Failed to marshal lastEvaluatedKey %s", err.Error()), nil)
	//		response.ErrorResponse(w, http.StatusInternalServerError, "Internal server error", 1003)
	//		return
	//	}
	//	nextKeyString = url.QueryEscape(string(keyBytes)) // URL encode the key
	//}

	responseData := map[string]interface{}{
		"serviceRequests":  responseBody,
		"lastEvaluatedKey": lastKey, // Send key for next request
	}

	logger.Info("Service requests fetched successfully", map[string]interface{}{
		"householderID": householderID,
	})
	response.SuccessResponse(w, responseData, "Service Request fetched successfully", http.StatusOK)
}

func (h *HouseholderController) ViewApprovedRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role := r.Context().Value("role").(string)
	var householderID string
	if role == "Admin" {
		householderID = r.URL.Query().Get("user_id")
		if householderID == "" {
			logger.Error("No query param", nil)
			response.ErrorResponse(w, http.StatusBadRequest, "user ID is required", 2001)
			return
		}
	} else if role == "Householder" {
		householderID = r.Context().Value("userID").(string)
	} else {
		logger.Error("Invalid role", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid role", 1007)
		return
	}
	limit, offset := util.GetPaginationParams(r)
	order := util.GetFilterParam(r, "order")
	var sortOrder string
	if order == "New to Old" {
		sortOrder = "DESC"
	} else if order == "Old to New" {
		sortOrder = "ASC"
	} else {
		sortOrder = ""
	}
	approvedRequests, err := h.householderService.ViewApprovedRequests(ctx, householderID, limit, offset, sortOrder)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1008)
		return
	}

	if len(approvedRequests) == 0 {
		logger.Info("No approve service requests", nil)
		response.SuccessResponse(w, nil, "No approved service requests found", http.StatusOK)
		//fmt.Println("No approved service requests found.")
		return
	}
	type responseStruct struct {
		ID              string                         `json:"request_id"`
		ServiceName     string                         `json:"service_name,omitempty"`
		ServiceID       string                         `json:"service_id" `
		RequestedTime   time.Time                      `json:"requested_time"`
		ScheduledTime   time.Time                      `json:"scheduled_time"`
		Status          string                         `json:"status"`
		ProviderDetails []model.ServiceProviderDetails `json:"provider_details,omitempty" bson:"providerDetails,omitempty"`
	}
	responseBody := make([]responseStruct, 0)
	for _, request := range approvedRequests {
		if request.ApproveStatus {
			currRequest := &responseStruct{
				ID:            request.ID,
				ServiceName:   request.ServiceName,
				ServiceID:     request.ServiceID,
				RequestedTime: request.RequestedTime,
				ScheduledTime: request.ScheduledTime,
				Status:        request.Status,
			}
			for _, provider := range request.ProviderDetails {
				if provider.Approve == 1 {
					currRequest.ProviderDetails = append(currRequest.ProviderDetails, model.ServiceProviderDetails{
						ServiceProviderID: provider.ServiceProviderID,
						Name:              provider.Name,
						Contact:           provider.Contact,
						Address:           provider.Address,
						Price:             provider.Price,
						Rating:            provider.Rating,
					})
				}

			}
			responseBody = append(responseBody, *currRequest)

		}

	}
	logger.Info("approved request fetched", nil)
	response.SuccessResponse(w, responseBody, "Approved requests fetched", http.StatusOK)

}

func (h *HouseholderController) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RequestID  string `json:"request_id" validate:"required"`
		ProviderID string `json:"provider_id" validate:"required"`
		ServiceId  string `json:"service_id" validate:"required"`
		Status     string `json:"status" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error("Invalid input", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input", 1001)
		return
	}

	err := validate.Struct(request)
	if err != nil {
		logger.Error("Invalid request body", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}
	ctx := r.Context()
	role := r.Context().Value("role").(string)
	var householderID string
	if role == "Admin" {
		householderID = r.URL.Query().Get("user_id")
		if householderID == "" {
			logger.Error("No query param", nil)
			response.ErrorResponse(w, http.StatusBadRequest, "user ID is required", 2001)
			return
		}
	} else if role == "Householder" {
		householderID = r.Context().Value("userID").(string)
	} else {
		logger.Error("Invalid role", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid role", 1007)
		return
	}
	status := util.ConvertStatus(request.Status)
	// Call the approval function
	if err = h.householderService.ApproveServiceRequest(ctx, request.RequestID, request.ProviderID, householderID, request.ServiceId, status); err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Internal server error", 1006)
		return
	}

	logger.Info("Request approve successfully", nil)
	response.SuccessResponse(w, nil, "Request approve successfully", http.StatusOK)

}

func (h *HouseholderController) LeaveReview(w http.ResponseWriter, r *http.Request) {
	var reviewRequest struct {
		RequestID  string  `json:"request_id" validate:"required"`
		ServiceID  string  `json:"service_id" validate:"required"`
		ProviderID string  `json:"provider_id" validate:"required"`
		ReviewText string  `json:"review_text" validate:"required"`
		Rating     float64 `json:"rating" validate:"required"`
	}

	// Decode request body into reviewRequest struct
	if err := json.NewDecoder(r.Body).Decode(&reviewRequest); err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Error decoding review request", 1001)
		return
	}
	err := validate.Struct(reviewRequest)
	if err != nil {
		logger.Error("Invalid request body", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}
	// Get user from context (assuming AuthMiddleware has set user info in the context)
	userID := r.Context().Value("userID").(string)
	ctx := r.Context()

	// Validate rating input (between 1 and 5)
	if reviewRequest.Rating < 1 || reviewRequest.Rating > 5 {
		response.ErrorResponse(w, http.StatusBadRequest, "Rating should be between 1 and 5", 1001)
		return
	}

	// Call the householder service to add the review
	err = h.householderService.AddReview(ctx, reviewRequest.ProviderID, userID, reviewRequest.ServiceID, reviewRequest.ReviewText, reviewRequest.Rating, reviewRequest.RequestID)
	if err != nil {
		if err.Error() == "review already exists" {
			logger.Error(fmt.Sprintf("Error adding review %v", err), nil)
			response.ErrorResponse(w, http.StatusConflict, err.Error(), 1006)
		} else {
			logger.Error(fmt.Sprintf("Error adding review %v", err), nil)
			response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1006)
		}
		return
	}

	// Successfully added the review
	logger.Info("Review added successfully", nil)
	response.SuccessResponse(w, nil, "Review added successfully", http.StatusOK)
}

func (h *HouseholderController) GetAllServiceCategories(w http.ResponseWriter, r *http.Request) {
	// Call the householder service to get all service categories
	ctx := r.Context()
	categories, err := h.householderService.GetAllServiceCategory(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Error fetching categories: %v", err), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch service categories", 1006)
		return
	}

	// Successfully retrieved the categories
	logger.Info("Categories retrieved successfully", nil)
	response.SuccessResponse(w, categories, "Categories retrieved successfully", http.StatusOK)
}
