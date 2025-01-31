package service

import (
	"context"
	"errors"
	"service-nest/interfaces"
	"service-nest/model"
	"service-nest/util"
)

type AdminService struct {
	serviceRepo        interfaces.ServiceRepository
	userRepo           interfaces.UserRepository
	householderRepo    interfaces.HouseholderRepository
	providerRepo       interfaces.ServiceProviderRepository
	serviceRequestRepo interfaces.ServiceRequestRepository
}

func NewAdminService(serviceRepo interfaces.ServiceRepository, serviceRequestRepo interfaces.ServiceRequestRepository, userRepo interfaces.UserRepository, providerRepo interfaces.ServiceProviderRepository) interfaces.AdminService {
	return &AdminService{
		serviceRepo:        serviceRepo,
		userRepo:           userRepo,
		providerRepo:       providerRepo,
		serviceRequestRepo: serviceRequestRepo,
	}
}

func (s *AdminService) ViewReports(limit, offset int) ([]model.ServiceRequest, error) {

	return s.serviceRequestRepo.GetAllServiceRequests(limit, offset)

}
func (s *AdminService) DeleteService(ctx context.Context, serviceID string) error {
	return s.serviceRepo.RemoveCategory(ctx, serviceID)
}

func (s *AdminService) DeactivateAccount(userID string) error {
	provider, err := s.providerRepo.GetProviderByID(userID)
	if err != nil {
		return err
	}

	provider.IsActive = false
	err = s.providerRepo.UpdateServiceProvider(provider)
	if err != nil {
		return err
	}
	err = s.userRepo.DeActivateUser(userID)
	if err != nil {
		return err
	}

	// Delete all services associated with the user
	err = s.providerRepo.DeleteServicesByProviderID(userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) GetAllService(limit, offset int) ([]model.Service, error) {
	return s.serviceRepo.GetAllServices(limit, offset)
}

func (s *AdminService) AddService(ctx context.Context, category *model.Category) error {
	category.ID = util.GenerateUUID()
	err := s.serviceRepo.AddCategory(ctx, category)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user.Role != "Householder" {
		return nil, errors.New("You are not a Householder")
	}
	return user, nil
}
