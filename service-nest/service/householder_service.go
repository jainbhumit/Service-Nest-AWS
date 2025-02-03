package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"service-nest/errs"
	"service-nest/interfaces"
	"service-nest/model"
	"service-nest/util"
	"time"
)

var GetUniqueID = util.GenerateUniqueID

type HouseholderService struct {
	householderRepo    interfaces.HouseholderRepository
	providerRepo       interfaces.ServiceProviderRepository
	serviceRepo        interfaces.ServiceRepository
	serviceRequestRepo interfaces.ServiceRequestRepository
	userRepo           interfaces.UserRepository
}

func NewHouseholderService(householderRepo interfaces.HouseholderRepository, providerRepo interfaces.ServiceProviderRepository, serviceRepo interfaces.ServiceRepository, serviceRequestRepo interfaces.ServiceRequestRepository, userRepo interfaces.UserRepository) interfaces.HouseholderService {
	return &HouseholderService{
		householderRepo:    householderRepo,
		providerRepo:       providerRepo,
		serviceRepo:        serviceRepo,
		serviceRequestRepo: serviceRequestRepo,
		userRepo:           userRepo,
	}
}
func (s *HouseholderService) ViewStatus(ctx context.Context, householderID string, limit int, lastEvaluatedKey string, status string) ([]model.ServiceRequest, map[string]types.AttributeValue, error) {
	// Fetch all service requests for the householder
	requests, lastKey, err := s.serviceRequestRepo.GetServiceRequestsByHouseholderID(ctx, householderID, limit, lastEvaluatedKey, status)
	if err != nil {
		return nil, nil, err
	}
	return requests, lastKey, nil
}

func (s *HouseholderService) GetServicesByCategoryId(ctx context.Context, categoryId string) ([]model.Service, error) {

	services, err := s.serviceRepo.GetServicesByCategoryId(ctx, categoryId)
	if err != nil {
		return nil, err
	}

	return services, nil
}

// RequestService allows the householder to request a service_test from a provider
func (s *HouseholderService) RequestService(ctx context.Context, householderRequest model.ServiceRequest) (string, error) {
	// Generate a unique ID for the service request
	requestID := util.GenerateUUID()

	// Fetch householder details from Db
	householder, err := s.userRepo.GetUserByID(ctx, *householderRequest.HouseholderID)
	if err != nil {
		return "", err
	}
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Printf("Failed to load location: %v", err)
	}
	// Create the service request
	serviceRequest := model.ServiceRequest{
		ID:                 requestID,
		ServiceName:        householderRequest.ServiceName,
		HouseholderName:    householder.Name,
		HouseholderID:      &householder.ID,
		HouseholderAddress: &householder.Address,
		HouseholderContact: householder.Contact,
		Description:        householderRequest.Description,
		ServiceID:          householderRequest.ServiceID,
		RequestedTime:      time.Now().In(location),
		ScheduledTime:      householderRequest.ScheduledTime,
		Status:             "Pending",
		ApproveStatus:      false,
	}
	// Save the service request to the repository
	err = s.serviceRequestRepo.SaveServiceRequest(ctx, serviceRequest)
	if err != nil {
		return "", err
	}

	return serviceRequest.ID, nil

}

// CancelServiceRequest allows the householder to cancel a service_test request
func (s *HouseholderService) CancelServiceRequest(ctx context.Context, requestID string, householderID string, status string) error {
	request, err := s.serviceRequestRepo.GetServiceRequestByID(ctx, requestID, householderID, status)
	if err != nil {
		return err
	}

	if *request.HouseholderID != householderID {
		return errors.New(errs.RequestNotBelongToHouseholder)
	}

	// Check if the scheduled time is less than 4 hours away
	currentTime := time.Now()
	fmt.Println(request.ScheduledTime, " ", currentTime, " ", request.ScheduledTime.Sub(currentTime))
	if request.ScheduledTime.Sub(currentTime) < 4*time.Hour {
		return fmt.Errorf(errs.RequestCancellationTooLate)
	}
	if request.Status == "Cancelled" {
		return fmt.Errorf(errs.RequestAlreadyCancelled)
	}
	request.Status = "Cancelled"
	return s.serviceRequestRepo.CancelServiceRequest(ctx, request, status)
}

// RescheduleServiceRequest allows the householder to reschedule a service_test request
func (s *HouseholderService) RescheduleServiceRequest(ctx context.Context, requestID string, newTime time.Time, householderID string, status string) error {
	request, err := s.serviceRequestRepo.GetServiceRequestByID(ctx, requestID, householderID, status)
	if err != nil {
		return err
	}

	if *request.HouseholderID != householderID {
		return errors.New(errs.RequestNotBelongToHouseholder)
	}
	if request.Status != "Pending" && request.Status != "Accepted" {
		return fmt.Errorf(errs.OnlyPendingRequestRescheduled)
	}

	request.ScheduledTime = newTime
	return s.serviceRequestRepo.UpdateServiceRequest(ctx, request, status)
}

func (s *HouseholderService) AddReview(ctx context.Context, providerID, householderID, serviceID, comments string, rating float64, requestId string) error {
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Printf("Failed to load location: %v", err)
	}

	// Create the review object
	review := model.Review{
		ID:            util.GenerateUUID(),
		ProviderID:    providerID,
		ServiceID:     serviceID,
		HouseholderID: householderID,
		Rating:        rating,
		Comments:      comments,
		ReviewDate:    time.Now().In(location),
		RequestId:     requestId,
	}

	// Save the review in the repository
	err = s.providerRepo.AddReview(ctx, review)
	if err != nil {
		return err
	}

	// Recalculate and update the provider's rating
	provider, err := s.serviceRepo.GetProviderByServiceId(ctx, providerID, serviceID)
	if err != nil {
		return errors.New(errs.FailUpdateRating)
	}
	updatedRating := util.CalculateRating(provider.AvgRating, provider.RatingCount, rating)
	provider.AvgRating = updatedRating
	provider.RatingCount = provider.RatingCount + 1
	s.serviceRepo.UpdateProviderRating(ctx, provider)
	return nil
}

func (s *HouseholderService) ApproveServiceRequest(ctx context.Context, requestID string, providerID string, householderID string, serviceId string, status string) error {
	// Retrieve the service request by ID

	_, err := s.serviceRepo.GetProviderByServiceId(ctx, providerID, serviceId)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	serviceRequest, err := s.serviceRequestRepo.GetServiceRequestByID(ctx, requestID, householderID, status)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if *serviceRequest.HouseholderID != householderID {
		return errors.New(errs.RequestNotBelongToHouseholder)
	}
	// Check if the request has already been approved
	if serviceRequest.ApproveStatus {
		return errors.New(errs.RequestAlreadyApproved)
	}

	// Set the approval status to true
	serviceRequest.ApproveStatus = true
	serviceRequest.Status = "Approved"
	var providerDetail model.ServiceProviderDetails
	for _, provider := range serviceRequest.ProviderDetails {
		if provider.ServiceProviderID == providerID {
			provider.Approve = 1
			providerDetail = provider
			break
		}
	}
	serviceRequest.ProviderDetails = append(serviceRequest.ProviderDetails, providerDetail)
	// Update the service request in the repository
	if err := s.serviceRequestRepo.ApproveServiceRequest(ctx, serviceRequest, status, serviceId); err != nil {
		return fmt.Errorf("%v: %v", errs.NotUpdateRequest, err)
	}

	return nil
}
func (s *HouseholderService) ViewApprovedRequests(ctx context.Context, householderID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error) {
	// Retrieve all service requests for the householder
	serviceRequests, err := s.serviceRequestRepo.GetApproveServiceRequestsByHouseholderID(ctx, householderID, limit, offset, sortOrder)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", errs.NotRetrieveRequest, err)
	}

	// Filter to only include approved requests
	var approvedRequests []model.ServiceRequest
	for _, req := range serviceRequests {
		if req.ApproveStatus && req.Status != "Cancelled" {
			approvedRequests = append(approvedRequests, req)
		}
	}

	if len(approvedRequests) == 0 {
		return nil, errors.New(errs.NoApproveRequestFound)
	}

	return approvedRequests, nil
}

func (s *HouseholderService) GetAllServiceCategory(ctx context.Context) ([]model.Category, error) {
	categories, err := s.serviceRepo.GetAllCategory(ctx)
	if err != nil {
		return nil, err
	}
	return categories, nil
}
