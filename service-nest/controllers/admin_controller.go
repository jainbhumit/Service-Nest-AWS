package controllers

import (
	"github.com/gorilla/mux"
	"net/http"
	"service-nest/interfaces"
	"service-nest/logger"
	"service-nest/model"
	"service-nest/response"
	"service-nest/util"
	"strings"
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

// ManageServices handles the services management functionality
func (a *AdminController) ViewAllService(w http.ResponseWriter, r *http.Request) {
	// Fetch all services (without limit and offset in service/repository)
	limit, offset := util.GetPaginationParams(r)
	services, err := a.adminService.GetAllService(limit, offset)
	if err != nil {
		logger.Error("error fetching services", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "error fetching services", 1003)
		return
	}

	logger.Info("All services fetched successfully", nil)
	response.SuccessResponse(w, services, "All available services", http.StatusOK)
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
	limit, offset := util.GetPaginationParams(r)
	reports, err := a.adminService.ViewReports(limit, offset)
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

	err := a.adminService.DeactivateAccount(providerID)
	if err != nil {
		logger.Error("error deactivating account", nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Error deactivating account", 1006)
		return
	}

	response.SuccessResponse(w, nil, "Account deactivated successfully", http.StatusOK)
}

func (a *AdminController) AddService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, "error parsing form", 1006)
		logger.Error("error parsing multipart form:", map[string]interface{}{"error": err})
		return
	}

	name := r.FormValue("name")
	if name == "" {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input name field missing", 1001)
		logger.Error("Invalid request body name field missing", nil)
		return
	}

	description := r.FormValue("description")
	if description == "" {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input description field missing", 1001)
		logger.Error("Invalid request body description field missing", nil)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input image field missing", 1001)
		logger.Error("Invalid request body image field missing", nil)
		return
	}
	defer file.Close()

	// Check file size (10MB limit)
	if handler.Size > 10*1024*1024 {
		response.ErrorResponse(w, http.StatusBadRequest, "Image file size too large (max 10MB)", 1001)
		logger.Error("Image file size too large", nil)
		return
	}

	// Upload the file to S3 bucket
	imageUrl, err := util.UploadFileToS3(file, handler.Filename)
	if err != nil {
		if strings.Contains(err.Error(), "invalid file type") {
			response.ErrorResponse(w, http.StatusBadRequest, "Invalid file type. Only images are allowed", 1001)
		} else {
			response.ErrorResponse(w, http.StatusInternalServerError, "Error uploading image", 1002)
		}
		logger.Error("Error uploading image", map[string]interface{}{"error": err})
		return
	}

	// Create a new category entry
	category := &model.Category{
		Name:        name,
		Description: description,
		ImageUrl:    imageUrl,
	}

	err = a.adminService.AddService(ctx, category)
	if err != nil {
		logger.Error("error Adding service", map[string]interface{}{"error": err})
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1006)
		return
	}

	response.SuccessResponse(w, nil, "Service added successfully", http.StatusOK)
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
