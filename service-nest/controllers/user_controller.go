package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"net/http"
	"service-nest/errs"
	"service-nest/interfaces"
	"service-nest/logger"
	"service-nest/model"
	"service-nest/response"
	"service-nest/util"
)

var CheckPassword = util.CheckPasswordHash
var HashPassword = util.HashPassword
var validate *validator.Validate
var GenerateJWT = util.GenerateJWT
var ValidatePassword = util.ValidatePassword

func init() {
	// initialize new validator
	validate = validator.New()
}

type UserController struct {
	userService interfaces.UserService
}

func NewUserController(userService interfaces.UserService) *UserController {
	return &UserController{userService: userService}
}

// LoginUser handles POST /login
func (u *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var userInput struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	var err error
	if err = json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input", 1001)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// validate userInput
	err = validate.Struct(userInput)
	if err != nil {
		logger.Error("Validation error", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	// Check if user exists and get the user details
	var user *model.User
	user, err = u.userService.CheckUserExists(ctx, userInput.Email)
	if err != nil {
		logger.Error("Invalid email or password", map[string]interface{}{"email": userInput.Email})
		response.ErrorResponse(w, http.StatusUnauthorized, "Invalid email or password", 1005)
		//http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Verify the password
	if !CheckPassword(userInput.Password, user.Password) {
		logger.Error("Invalid password", map[string]interface{}{"email": userInput.Email})
		response.ErrorResponse(w, http.StatusUnauthorized, "Invalid password", 1005)
		//http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}
	if user.IsActive == false {
		logger.Error("User is deactivated by admin", map[string]interface{}{"email": userInput.Email})
		response.ErrorResponse(w, http.StatusUnauthorized, "user Deactivated by admin", 1007)
		return
	}

	// Generate JWT token
	var tokenString string
	tokenString, err = GenerateJWT(user.ID, user.Role)
	if err != nil {
		logger.Error("Error generating token", map[string]interface{}{"email": userInput.Email})
		response.ErrorResponse(w, http.StatusInternalServerError, "Error generating token", 1006)
		//http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Return the token as JSON
	logger.Info("token generated", map[string]interface{}{"email": userInput.Email})
	response.SuccessResponse(w, map[string]interface{}{"token": tokenString}, "Token generate successfully", http.StatusCreated)
}
func (u *UserController) SignupUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var newUser struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
		Role     string `json:"role" validate:"required"`
		Address  string `json:"address" validate:"required"`
		Contact  string `json:"contact" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		logger.Error("Invalid input", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid input", 1001)
		return
	}
	err := validate.Struct(newUser)
	if err != nil {
		logger.Error("Validation error", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}
	// Check if user already exists
	exists, err := u.userService.CheckUserExists(ctx, newUser.Email)

	if exists != nil && err == nil {
		logger.Error("User already exists", map[string]interface{}{"email": newUser.Email})
		response.ErrorResponse(w, http.StatusConflict, "User already exists", 1009)
		//http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the password before saving
	hashedPassword, err := HashPassword(newUser.Password)
	if err != nil {
		logger.Error("Error hashing password", map[string]interface{}{"email": newUser.Email})
		response.ErrorResponse(w, http.StatusInternalServerError, "Error hashing password", 1006)
		//http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	newUser.Password = hashedPassword

	// Save the user
	user := &model.User{
		Name:     newUser.Name,
		Email:    newUser.Email,
		Password: newUser.Password,
		Role:     newUser.Role,
		Address:  newUser.Address,
		Contact:  newUser.Contact,
		IsActive: true,
	}
	err = u.userService.CreateUser(ctx, user)
	if err != nil {
		logger.Error("Error creating user", map[string]interface{}{"email": newUser.Email})
		response.ErrorResponse(w, http.StatusInternalServerError, "Error creating user", 1006)
		//http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	logger.Info("User created sucessfully", map[string]interface{}{"email": newUser.Email})
	response.SuccessResponse(w, nil, "User created successfully", http.StatusCreated)
	//w.WriteHeader(http.StatusCreated)
	//
	//json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

// ViewProfileByIDHandler handles GET /users/{id}
func (u *UserController) ViewProfileByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Context().Value("userID").(string)

	// Call the UserService to get the user profile
	user, err := u.userService.ViewProfileByID(ctx, userID)
	if err != nil {
		logger.Error(err.Error(), nil)
		response.ErrorResponse(w, http.StatusNotFound, "error viewing user", 1008)
		//http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	logger.Info("user profile fetched", map[string]interface{}{"email": user.Email})
	response.SuccessResponse(w, user, "User profile fetched successfully", http.StatusOK)

}

// UpdateUserHandler handles PUT /users/{id}
func (u *UserController) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Context().Value("userID").(string)

	// Parse incoming JSON data
	var updateData struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
		Address  *string `json:"address"`
		Contact  *string `json:"contact"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		logger.Error("Invalid input", map[string]interface{}{"body": updateData})
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		//http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	err := ValidatePassword(*updateData.Password)
	if err != nil {
		logger.Error("Error validating password", map[string]interface{}{"email": *updateData.Email})
		response.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v", err), 1001)
		//http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}
	hashedPassword, err := HashPassword(*updateData.Password)
	if err != nil {
		logger.Error("Error hashing password", map[string]interface{}{"email": *updateData.Email})
		response.ErrorResponse(w, http.StatusInternalServerError, "Error hashing password", 1006)
		//http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Call the UserService to update the user profile
	err = u.userService.UpdateUser(ctx, userID, updateData.Email, &hashedPassword, updateData.Address, updateData.Contact)
	if err != nil {
		response.ErrorResponse(w, http.StatusInternalServerError, err.Error(), 1006)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("user updated sucessfully", map[string]interface{}{"email": *updateData.Email})
	response.SuccessResponse(w, nil, "User updated successfully", http.StatusOK)

}

func (u *UserController) ForgetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	// Parse incoming JSON data
	var updateData struct {
		Email          *string `json:"email" validate:"required"`
		SecurityAnswer *string `json:"security_answer" validate:"required"`
		Password       *string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	err := validate.Struct(updateData)

	if err != nil {
		logger.Error("Validation error", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	err = ValidatePassword(*updateData.Password)
	if err != nil {
		logger.Error("Error validating password", map[string]interface{}{"email": *updateData.Email})
		response.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v", err), 1001)

		return
	}
	hashedPassword, err := HashPassword(*updateData.Password)
	if err != nil {
		logger.Error("Error hashing password", map[string]interface{}{"email": *updateData.Email})
		response.ErrorResponse(w, http.StatusInternalServerError, "Error hashing password", 1006)
		return
	}

	// Call the UserService to update the user profile

	err1 := u.userService.ForgetPasword(*updateData.Email, *updateData.SecurityAnswer, hashedPassword)
	if err1 != nil {

		if err1.Error() == errs.UserNotFound {
			logger.Error("User not found", nil)
			response.ErrorResponse(w, http.StatusNotFound, "email doesn't exist", 1008)
			return
		} else if err1.Error() == errs.IncorrectSecurityAnswer {
			logger.Error(err1.Error(), nil)
			response.ErrorResponse(w, http.StatusUnauthorized, "incorrect security answer", 1007)
			return
		}
		logger.Error(err1.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Error updating user", 1006)
		return
	}
	logger.Info("password updated sucessfully", map[string]interface{}{"email": *updateData.Email})
	response.SuccessResponse(w, nil, "User password updated successfully", http.StatusOK)

}

func (u *UserController) VerifyOtpAndUpdatePassword(w http.ResponseWriter, r *http.Request) {

	// Parse incoming JSON data
	var updateData struct {
		Email    *string `json:"email" validate:"required"`
		Otp      *string `json:"otp" validate:"required"`
		Password *string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	err := validate.Struct(updateData)

	if err != nil {
		logger.Error("Validation error", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	err = ValidatePassword(*updateData.Password)
	if err != nil {
		logger.Error("Error validating password", map[string]interface{}{"email": *updateData.Email})
		response.ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%v", err), 1001)

		return
	}
	hashedPassword, err := HashPassword(*updateData.Password)
	if err != nil {
		logger.Error("Error hashing password", map[string]interface{}{"email": *updateData.Email})
		response.ErrorResponse(w, http.StatusInternalServerError, "Error hashing password", 1006)
		return
	}

	// Call the UserService to update the user profile

	err1 := u.userService.VerifyAndUpdatePassword(*updateData.Email, *updateData.Otp, hashedPassword)
	if err1 != nil {

		if err1.Error() == errs.UserNotFound {
			logger.Error("User not found", nil)
			response.ErrorResponse(w, http.StatusNotFound, "email doesn't exist", 1008)
			return
		} else if err1.Error() == errs.IncorrectSecurityAnswer {
			logger.Error(err1.Error(), nil)
			response.ErrorResponse(w, http.StatusUnauthorized, "incorrect security answer", 1007)
			return
		} else if err1.Error() == "Invalid Otp" {
			logger.Error(err1.Error(), nil)
			response.ErrorResponse(w, http.StatusUnauthorized, "incorrect otp", 1008)
			return
		}
		logger.Error(err1.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, "Error updating user", 1006)
		return
	}
	logger.Info("password updated sucessfully", map[string]interface{}{"email": *updateData.Email})
	response.SuccessResponse(w, nil, "User password updated successfully", http.StatusOK)

}

func (u *UserController) GenerateOtp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Parse incoming JSON data
	var updateData struct {
		Email *string `json:"email" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	err := validate.Struct(updateData)

	if err != nil {
		logger.Error("Validation error", nil)
		response.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", 1001)
		return
	}

	err1 := u.userService.GenerateOtp(ctx, *updateData.Email)
	if err1 != nil {
		logger.Error(err1.Error(), nil)
		response.ErrorResponse(w, http.StatusInternalServerError, err1.Error(), 1006)
		return
	}
	logger.Info("Otp send successfully", map[string]interface{}{"email": *updateData.Email})
	response.SuccessResponse(w, nil, "Otp Sent successfully", http.StatusOK)

}
