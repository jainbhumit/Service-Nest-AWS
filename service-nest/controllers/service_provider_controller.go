package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"service-nest/interfaces"
	"service-nest/logger"
	"service-nest/model"
	"service-nest/response"
	"service-nest/util"
	"time"
)

var GenerateUniqueID = util.GenerateUniqueID

type ServiceProviderController struct {
	serviceProviderService interfaces.ServiceProviderService
}

func NewServiceProviderController(serviceProviderService interfaces.ServiceProviderService) *ServiceProviderController {
	return &ServiceProviderController{serviceProviderService: serviceProviderService}
}

func (s *ServiceProviderController) AddService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request struct {
		Name        string  `json:"name" validate:"required"`
		Description string  `json:"description" validate:"required"`
		Price       float64 `json:"price" validate:"required"`
		Category    string  `json:"category" validate:"required"`
		ServiceId   string  `json:"service_id" validate:"required"`
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
	providerID := r.Context().Value("userID").(string)

	newService := &model.Service{
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Category:    request.Category,
		ProviderID:  providerID,
		ID:          request.ServiceId,
	}

	serviceId, err := s.serviceProviderService.AddService(ctx, providerID, *newService)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "error adding services", 1003)

		//http.Error(w, "Error adding service", http.StatusInternalServerError)
		return
	}
	var serviceID struct {
		ID string `json:"service_id"`
	}
	serviceID.ID = serviceId
	response.SuccessResponse(w, serviceID, "Service added successfully", http.StatusOK)

}

func (s *ServiceProviderController) ViewServices(w http.ResponseWriter, r *http.Request) {
	providerID := r.Context().Value("userID").(string)
	ctx := r.Context()
	services, err := s.serviceProviderService.ViewServices(ctx, providerID)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Error fetching services", 1006)
		//http.Error(w, "Error fetching services", http.StatusInternalServerError)
		return
	}
	if len(services) == 0 {
		response.SuccessResponse(w, nil, "Don't have a service offered", http.StatusOK)
		return
	}

	logger.Info("Service fetched successfully", nil)
	response.SuccessResponse(w, services, "service fetch successfully", http.StatusOK)

}
func (h *ServiceProviderController) UpdateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	serviceID := vars["service_id"]
	var request struct {
		Name        string  `json:"name" validate:"required"`
		Description string  `json:"description" validate:"required"`
		Price       float64 `json:"price" validate:"required"`
		Category    string  `json:"category" validate:"required"`
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
	providerID := r.Context().Value("userID").(string)

	updatedService := &model.Service{
		ID:          serviceID,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Category:    request.Category,
		ProviderID:  providerID,
	}

	err = h.serviceProviderService.UpdateService(ctx, *updatedService)
	if err != nil {
		logger.Error("Error updating service", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1003)
		//http.Error(w, "Error updating service", http.StatusInternalServerError)
		return
	}
	logger.Info("update successfully", nil)
	response.SuccessResponse(w, nil, "Service updated successfully", http.StatusOK)
}

func (s *ServiceProviderController) RemoveService(w http.ResponseWriter, r *http.Request) {
	providerID := r.Context().Value("userID").(string)
	ctx := r.Context()
	serviceID := mux.Vars(r)["service_id"]

	err := s.serviceProviderService.RemoveService(ctx, providerID, serviceID)
	if err != nil {
		logger.Error("Error removing service", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1008)
		//http.Error(w, "Error removing service", http.StatusInternalServerError)
		return
	}
	response.SuccessResponse(w, nil, "Service removed successfully", http.StatusOK)
	//json.NewEncoder(w).Encode("Service removed successfully")
}

func (s *ServiceProviderController) ViewServiceRequest(w http.ResponseWriter, r *http.Request) {
	providerID := r.Context().Value("userID").(string)
	ctx := r.Context()
	limit, offset := util.GetPaginationParams(r)
	serviceID := util.GetFilterParam(r, "serviceId")
	serviceRequests, err := s.serviceProviderService.GetAllServiceRequests(ctx, providerID, serviceID, limit, offset)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("error fetching request %v", err), 1006)
		return
	}

	type responseStruct struct {
		ID                 string    `json:"request_id"`
		ServiceName        string    `json:"service_name,omitempty"`
		ServiceID          string    `json:"service_id" `
		RequestedTime      time.Time `json:"requested_time"`
		ScheduledTime      time.Time `json:"scheduled_time"`
		HouseholderAddress string    `json:"address"`
		Description        string    `json:"description"`
		Status             string    `json:"status"`
	}
	responseBody := make([]responseStruct, 0)
	// Filter and display only pending requests

	for _, request := range serviceRequests {
		if request.ApproveStatus == false && request.Status != "Cancelled" {
			responseBody = append(responseBody, responseStruct{
				ID:                 request.ID,
				ServiceName:        request.ServiceName,
				ServiceID:          request.ServiceID,
				RequestedTime:      request.RequestedTime,
				ScheduledTime:      request.ScheduledTime,
				HouseholderAddress: *request.HouseholderAddress,
				Description:        request.Description,
				Status:             request.Status,
			})
		}
	}

	if len(responseBody) == 0 {
		logger.Info("No service requests found", nil)
		response.SuccessResponse(w, nil, "No pending service requests available", http.StatusOK)
		//color.Yellow("No pending service requests available.")
		return
	}
	logger.Info("Service request fetched successfully", nil)
	response.SuccessResponse(w, responseBody, "Service request fetched successfully", http.StatusOK)

}

func (s *ServiceProviderController) AcceptServiceRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID             string `json:"request_id" validate:"required"`
		EstimatedPrice string `json:"price" validate:"required"`
		ServiceID      string `json:"service_id" validate:"required"`
		Status         string `json:"status" validate:"required"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid body", 1001)
		//http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	err = validate.Struct(request)
	if err != nil {
		logger.Error("Invalid request body", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}
	ctx := r.Context()
	providerID := r.Context().Value("userID").(string)
	status := util.ConvertStatus(request.Status)
	err = s.serviceProviderService.AcceptServiceRequest(ctx, providerID, request.ID, request.EstimatedPrice, request.ServiceID, status)

	if err != nil {
		logger.Error(err.Error(), nil)
		if err.Error() == "provider not found" {
			response.SuccessResponse(w, nil, "provider not found", 200)
			return
		}
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1008)
		//http.Error(w, "Error accepting service request", http.StatusInternalServerError)
		return
	}

	logger.Info("Request accept successfully", nil)
	response.SuccessResponse(w, nil, "Request accept successfully", http.StatusOK)

}

func (s *ServiceProviderController) ViewApprovedRequests(w http.ResponseWriter, r *http.Request) {
	providerID := r.Context().Value("userID").(string)
	ctx := r.Context()
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
	approvedRequests, err := s.serviceProviderService.ViewApprovedRequestsByProvider(ctx, providerID, limit, offset, sortOrder)

	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1008)

		//http.Error(w, "Error fetching approved requests", http.StatusInternalServerError)
		return
	}
	fmt.Println(approvedRequests)
	type responseStruct struct {
		ID                 string                         `json:"request_id"`
		ServiceName        string                         `json:"service_name,omitempty"`
		HouseholderId      string                         `json:"householder_id"`
		HouseholderName    string                         `json:"householder_name"`
		HouseholderAddress string                         `json:"householder_address"`
		ApproveStatus      bool                           `json:"approve_status"`
		ServiceID          string                         `json:"service_id" `
		RequestedTime      time.Time                      `json:"requested_time"`
		ScheduledTime      time.Time                      `json:"scheduled_time"`
		Contact            string                         `json:"householder_contact"`
		Status             string                         `json:"status"`
		ProviderDetails    []model.ServiceProviderDetails `json:"provider_details,omitempty" bson:"providerDetails,omitempty"`
	}
	responseBody := make([]responseStruct, 0)
	for _, request := range approvedRequests {
		if request.ApproveStatus {
			currRequest := &responseStruct{
				ID:                 request.ID,
				ServiceName:        request.ServiceName,
				ServiceID:          request.ServiceID,
				RequestedTime:      request.RequestedTime,
				ScheduledTime:      request.ScheduledTime,
				Status:             request.Status,
				HouseholderId:      *request.HouseholderID,
				HouseholderAddress: *request.HouseholderAddress,
				HouseholderName:    request.HouseholderName,
				Contact:            request.HouseholderContact,
			}
			responseBody = append(responseBody, *currRequest)

		}

	}
	response.SuccessResponse(w, responseBody, "Approve requests fetched successfully", http.StatusOK)

	//json.NewEncoder(w).Encode(approvedRequests)
}

func (s *ServiceProviderController) ViewReviews(w http.ResponseWriter, r *http.Request) {
	providerID := r.Context().Value("userID").(string)
	limit, offset := util.GetPaginationParams(r)
	// Call the service to get the reviews
	serviceID := util.GetFilterParam(r, "serviceId")
	reviews, err := s.serviceProviderService.GetReviews(providerID, limit, offset, serviceID)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1003)

		//http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}

	// Send reviews as JSON response
	response.SuccessResponse(w, reviews, "Reviews fetched successfully", http.StatusOK)
}
