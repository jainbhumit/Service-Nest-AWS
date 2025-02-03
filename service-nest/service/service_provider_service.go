package service

import (
	"context"
	"errors"
	"fmt"
	"service-nest/interfaces"
	"service-nest/model"
)

type ServiceProviderService struct {
	serviceProviderRepo interfaces.ServiceProviderRepository
	serviceRequestRepo  interfaces.ServiceRequestRepository
	serviceRepo         interfaces.ServiceRepository
	userRepo            interfaces.UserRepository
}

// NewServiceProviderService initializes a new ServiceProviderService
func NewServiceProviderService(serviceProviderRepo interfaces.ServiceProviderRepository, serviceRequestRepo interfaces.ServiceRequestRepository, serviceRepo interfaces.ServiceRepository, userRepo interfaces.UserRepository) interfaces.ServiceProviderService {
	return &ServiceProviderService{
		serviceProviderRepo: serviceProviderRepo,
		serviceRequestRepo:  serviceRequestRepo,
		serviceRepo:         serviceRepo,
		userRepo:            userRepo,
	}
}

// AddService associates a predefined service to a provider’s offerings using CategoryID.
func (s *ServiceProviderService) AddService(ctx context.Context, providerID string, newService model.Service) (string, error) {
	providerDetail, err := s.userRepo.GetUserByID(ctx, providerID)
	if err != nil {
		return "", err
	}
	newService.AvgRating = 0
	newService.RatingCount = 0
	newService.ProviderName = providerDetail.Name
	newService.ProviderContact = providerDetail.Contact
	newService.ProviderAddress = providerDetail.Address
	err = s.serviceRepo.SaveService(ctx, newService)
	if err != nil {
		return "", err
	}

	return newService.ID, nil
}

// UpdateService updates an existing service_test offered by the provider
func (s *ServiceProviderService) UpdateService(ctx context.Context, updatedService model.Service) error {
	// Save the updated service provider information
	err := s.serviceRepo.UpdateService(ctx, updatedService.ProviderID, updatedService)
	if err != nil {
		return err
	}

	// Update the service in the service repository
	return nil
}
func (s *ServiceProviderService) GetAllServiceRequests(ctx context.Context, providerId string, serviceID string, limit, offset int) ([]model.ServiceRequest, error) {
	return s.serviceRequestRepo.GetAllPendingRequestsByProvider(ctx, providerId, serviceID, limit, offset)
}

func (s *ServiceProviderService) RemoveService(ctx context.Context, providerID, serviceID string) error {
	err := s.serviceRepo.RemoveServiceByProviderID(ctx, providerID, serviceID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServiceProviderService) AcceptServiceRequest(ctx context.Context, providerID, requestID string, estimatedPrice string, serviceId string, status string) error {
	serviceRequest, err := s.serviceRequestRepo.GetServiceRequestByProvider(ctx, requestID, serviceId)
	if err != nil {
		return err
	}

	if serviceRequest.ApproveStatus {
		return fmt.Errorf("service request has already been approved")
	}

	// Update the service request status to "Accepted"
	serviceRequest.Status = "Accepted"

	// Get the ServiceProvider details
	provider, err := s.serviceRepo.GetProviderByServiceId(ctx, providerID, serviceId)
	if err != nil {
		return err
	}

	// Add ServiceProvider details to the ServiceRequest
	serviceRequest.ProviderDetails = append(serviceRequest.ProviderDetails, model.ServiceProviderDetails{
		ServiceProviderID: providerID,
		Name:              provider.ProviderName,
		Contact:           provider.ProviderContact,
		Address:           provider.ProviderAddress,
		Price:             estimatedPrice,
		Rating:            provider.AvgRating,
	})

	// Save the updated service request
	err = s.serviceRequestRepo.AcceptServiceRequestByProvider(ctx, serviceRequest, status)

	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceProviderService) ViewServices(ctx context.Context, providerID string) ([]model.Service, error) {
	providerService, err := s.serviceRepo.GetServiceByProviderID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	return providerService, nil
}

func (s *ServiceProviderService) ViewApprovedRequestsByProvider(ctx context.Context, providerID string, limit, offset int, sortOrder string) ([]model.ServiceRequest, error) {
	// Fetch all service requests related to the provider
	serviceRequests, err := s.serviceRequestRepo.GetApproveServiceRequestsByProviderID(ctx, providerID, limit, offset, sortOrder)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve service requests: %v", err)
	}
	fmt.Println(serviceRequests)

	if len(serviceRequests) == 0 {
		return nil, errors.New("no approved requests found for this provider")
	}

	return serviceRequests, nil
}

func (s *ServiceProviderService) GetReviews(ctx context.Context, providerID string, limit, offset int, serviceID string) ([]model.Review, error) {
	reviews, err := s.serviceProviderRepo.GetReviewsByProviderID(ctx, providerID, limit, offset, serviceID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}
