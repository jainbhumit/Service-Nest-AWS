package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"service-nest/interfaces"
	"service-nest/logger"
	"service-nest/model"
	"service-nest/response"
	"service-nest/util"
)

type AdminController struct {
	adminService interfaces.AdminService
}

// NewAdminController initializes a new AdminController with the given service
func NewAdminController(adminService interfaces.AdminService) *AdminController {
	return &AdminController{
		adminService: adminService,
	}
}

// DeleteService allows the admin to delete a service
func (a *AdminController) DeleteService(w http.ResponseWriter, r *http.Request) {
	serviceID := mux.Vars(r)["serviceID"]
	ctx := r.Context()
	err := a.adminService.DeleteService(ctx, serviceID)
	if err != nil {
		logger.Error("error deleting service", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Error deleting service", 1006)
		return
	}

	response.SuccessResponse(w, nil, "Service deleted successfully", http.StatusOK)
}

// ViewReports allows the admin to view reports
func (a *AdminController) ViewReports(w http.ResponseWriter, r *http.Request) {
	// Fetch all reports (without limit and offset in service/repository)
	ctx := r.Context()
	categoryId := r.URL.Query().Get("category_id")
	if categoryId == "" {
		logger.Error("No query param for category", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "request status is required", 2001)
		return
	}
	limit, offset := util.GetPaginationParams(r)
	reports, err := a.adminService.ViewReports(ctx, limit, offset, categoryId)
	if err != nil {
		logger.Error("error fetching reports", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Error generating reports", 1006)
		return
	}

	paginatedReports := util.ApplyPagination(reports, limit, offset)

	response.SuccessResponse(w, paginatedReports, "Reports fetched successfully", http.StatusOK)
}

// DeactivateUserAccount allows the admin to deactivate a user account
func (a *AdminController) DeactivateUserAccount(w http.ResponseWriter, r *http.Request) {
	providerID := mux.Vars(r)["providerID"]
	ctx := r.Context()
	err := a.adminService.DeactivateAccount(ctx, providerID)
	if err != nil {
		logger.Error(fmt.Sprintf("error deactivating account %s", err.Error()), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Error deactivating account", 1006)
		return
	}

	response.SuccessResponse(w, nil, "Account deactivated successfully", http.StatusOK)
}

func (a *AdminController) AddService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request struct {
		Name        string `json:"category_name" validate:"required"`
		Description string `json:"description" validate:"required"`
		FileName    string `json:"file_name" validate:"required"`
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
	presignedUrl, objectUrl, _ := util.GeneratePresignedURL(context.TODO(), request.FileName)
	category := &model.Category{
		Name:        request.Name,
		Description: request.Description,
		ImageUrl:    objectUrl,
	}
	err = a.adminService.AddService(ctx, category)
	if err != nil {
		logger.Error("error Adding service", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1006)
		return
	}
	type responseBody struct {
		PreSignedUrl string `json:"pre_signed_url"`
	}
	url := responseBody{
		PreSignedUrl: presignedUrl,
	}
	response.SuccessResponse(w, url, "Service added successfully", http.StatusOK)
}
func (a *AdminController) ViewUserDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userEmail := vars["userEmail"]

	user, err := a.adminService.GetUserByEmail(ctx, userEmail)

	if err != nil {
		logger.Error("error fetching services", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1003)
		return
	}

	type responseBody struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Address string `json:"address"`
	}
	userDetail := responseBody{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Address: user.Address,
	}

	logger.Info("All services fetched successfully", nil)
	response.SuccessResponse(w, userDetail, "User fetch successfully", http.StatusOK)
}
