package routers

import (
	"github.com/gorilla/mux"
	"net/http"
	"service-nest/controllers"
	"service-nest/interfaces"
	"service-nest/middlewares"
	"service-nest/response"
)

func SetupRouter(userService interfaces.UserService, householderService interfaces.HouseholderService, providerService interfaces.ServiceProviderService, adminService interfaces.AdminService) *mux.Router {
	r := mux.NewRouter()
	r.Use(middlewares.CORSMiddleware)
	r.Use(middlewares.LoggingMiddleware)
	// Public Routes
	userController := controllers.NewUserController(userService)

	r.HandleFunc("/signup", userController.SignupUser).Methods("POST")

	r.HandleFunc("/login", userController.LoginUser).Methods("POST")

	r.HandleFunc("/forget", userController.ForgetPasswordHandler).Methods("Put")

	r.HandleFunc("/forgot", userController.VerifyOtpAndUpdatePassword).Methods("PUT")

	r.HandleFunc("/otp", userController.GenerateOtp).Methods("POST")
	// Protected Routes (JWT authentication required)
	api := r.PathPrefix("/api").Subrouter()

	// User routes

	userRoutes := api.PathPrefix("/user").Subrouter()
	userRoutes.Use(middlewares.AuthMiddleware)
	userRoutes.HandleFunc("/profile", userController.ViewProfileByIDHandler).Methods("GET")
	userRoutes.HandleFunc("/profile", userController.UpdateUserHandler).Methods("PUT")

	// Householder routes connected to admin
	householderController := controllers.NewHouseholderController(householderService)
	userRoutes.HandleFunc("/categories", householderController.GetAllServiceCategories).Methods("GET")

	userRoutes.HandleFunc("/services/request", householderController.RequestService).Methods("POST")

	userRoutes.HandleFunc("/services/request", householderController.RescheduleServiceRequest).Methods("PUT")

	userRoutes.HandleFunc("/services/request/{request_id}", householderController.CancelServiceRequest).Methods("PATCH")

	userRoutes.HandleFunc("/bookings", householderController.ViewBookingHistory).Methods("GET")

	userRoutes.HandleFunc("/services/request/approve", householderController.ApproveRequest).Methods("PUT")

	householderRoutes := api.PathPrefix("/householder").Subrouter()
	householderRoutes.Use(middlewares.AuthMiddleware)
	householderRoutes.Use(middlewares.HouseHolderAuthMiddleware)

	householderRoutes.HandleFunc("/review", householderController.LeaveReview).Methods("POST")

	// Service provider routes
	serviceProviderController := controllers.NewServiceProviderController(providerService)

	providerRoutes := api.PathPrefix("/provider").Subrouter()
	providerRoutes.Use(middlewares.AuthMiddleware)
	providerRoutes.Use(middlewares.ServiceProviderAuthMiddleware)

	providerRoutes.HandleFunc("/service", serviceProviderController.AddService).Methods("POST")

	providerRoutes.HandleFunc("/service", serviceProviderController.ViewServices).Methods("GET")

	providerRoutes.HandleFunc("/service/{service_id}", serviceProviderController.UpdateService).Methods("PUT")

	providerRoutes.HandleFunc("/service/{service_id}", serviceProviderController.RemoveService).Methods("DELETE")

	providerRoutes.HandleFunc("/service/requests", serviceProviderController.ViewServiceRequest).Methods("GET")

	providerRoutes.HandleFunc("/service/requests", serviceProviderController.AcceptServiceRequest).Methods("POST")

	providerRoutes.HandleFunc("/reviews", serviceProviderController.ViewReviews).Methods("GET")
	userRoutes.HandleFunc("/service/request/approved", func(w http.ResponseWriter, r *http.Request) {
		// Get role from context
		role, ok := r.Context().Value("role").(string)
		if !ok {
			// If the role is not present or not a string, return an error
			response.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized access - missing or invalid role", 1002)
			return
		}
		// Role-based request handling
		switch role {
		case "Householder":
			householderController.ViewApprovedRequest(w, r)
		case "Admin":
			householderController.ViewApprovedRequest(w, r)
		case "ServiceProvider":
			serviceProviderController.ViewApprovedRequests(w, r)
		}
	}).Methods("GET")

	//admin routes
	adminRoutes := api.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(middlewares.AuthMiddleware)
	adminRoutes.Use(middlewares.AdminAuthMiddleware)

	// Initialize AdminController
	adminController := controllers.NewAdminController(adminService)

	// Routes for admin actions
	//adminRoutes.HandleFunc("/services", adminController.ViewAllService).Methods("GET")
	adminRoutes.HandleFunc("/service/{serviceID}", adminController.DeleteService).Methods("DELETE")
	adminRoutes.HandleFunc("/reports", adminController.ViewReports).Methods("GET")
	adminRoutes.HandleFunc("/deactivate/{providerID}", adminController.DeactivateUserAccount).Methods("PATCH")
	adminRoutes.HandleFunc("/service", adminController.AddService).Methods("POST")
	adminRoutes.HandleFunc("/users/{userEmail}", adminController.ViewUserDetail).Methods("GET")

	userRoutes.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("role").(string)
		if !ok {
			response.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized access - missing or invalid role", 1002)
			return
		}
		switch role {
		case "Householder":
			householderController.GetAvailableServices(w, r)
		case "Admin":
			adminController.ViewAllService(w, r)
		default:
			// If role is not recognized, return a forbidden status
			response.ErrorResponse(w, http.StatusForbidden, "Forbidden - role not allowed", 1002)
		}
	}).Methods("GET")
	return r
}
