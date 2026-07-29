package app

import (
	"fmt"

	"github.com/gorilla/mux"
	"service-nest/config"
	"service-nest/model"
	"service-nest/repository"
	"service-nest/routers"
	"service-nest/service"
)

// Application wires dependencies and exposes the HTTP router.
type Application struct {
	Router   *mux.Router
	Notifier *model.ErrorNotifier
}

// Bootstrap loads AWS clients and builds the router with all services.
func Bootstrap(cfg *config.EnvConfig) (*Application, error) {
	if cfg == nil {
		return nil, fmt.Errorf("env config must not be nil")
	}

	clients, err := config.NewClients(cfg)
	if err != nil {
		return nil, fmt.Errorf("aws clients: %w", err)
	}

	notifier := &model.ErrorNotifier{
		SnsClient: clients.SNS,
		TopicArn:  cfg.SNSTopicARN,
	}

	userRepo := repository.NewUserRepository(clients.DynamoDB)
	otpRepo := repository.NewOtpRepository()
	householderRepo := repository.NewHouseholderRepository(clients.DynamoDB)
	providerRepo := repository.NewServiceProviderRepository(clients.DynamoDB)
	serviceRepo := repository.NewServiceRepository(clients.DynamoDB)
	requestRepo := repository.NewServiceRequestRepository(clients.DynamoDB)

	userService := service.NewUserService(userRepo, otpRepo)
	householderService := service.NewHouseholderService(householderRepo, providerRepo, serviceRepo, requestRepo, userRepo)
	providerService := service.NewServiceProviderService(providerRepo, requestRepo, serviceRepo, userRepo)
	adminService := service.NewAdminService(serviceRepo, requestRepo, userRepo, providerRepo)

	router := routers.SetupRouter(userService, householderService, providerService, adminService, notifier, clients.DynamoDB)

	return &Application{
		Router:   router,
		Notifier: notifier,
	}, nil
}
