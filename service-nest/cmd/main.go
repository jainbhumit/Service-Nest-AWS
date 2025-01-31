package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"log"
	"net/http"
	"service-nest/repository"
	"service-nest/routers"
	"service-nest/service"
)

var gorillaLambda *gorillamux.GorillaMuxAdapter

func runApp(client *dynamodb.Client) {
	userRepo := repository.NewUserRepository(client)
	otpRepo := repository.NewOtpRepository()
	householderRepo := repository.NewHouseholderRepository(client)
	providerRepo := repository.NewServiceProviderRepository(client)
	serviceRepo := repository.NewServiceRepository(client)
	requestRepo := repository.NewServiceRequestRepository(client)
	userService := service.NewUserService(userRepo, otpRepo)

	householderService := service.NewHouseholderService(householderRepo, providerRepo, serviceRepo, requestRepo, userRepo)
	providerService := service.NewServiceProviderService(providerRepo, requestRepo, serviceRepo, userRepo)
	adminService := service.NewAdminService(serviceRepo, requestRepo, userRepo, providerRepo)

	r := routers.SetupRouter(userService, householderService, providerService, adminService)
	gorillaLambda = gorillamux.New(r)

}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, err := gorillaLambda.ProxyWithContext(ctx, *core.NewSwitchableAPIGatewayRequestV1(&req))
	if err != nil {
		log.Printf("Lambda proxy error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, err
	}
	return *r.Version1(), nil
}
func initializeDynamoDBClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}
	return dynamodb.NewFromConfig(cfg)
}
func main() {
	client := initializeDynamoDBClient()
	runApp(client)
	lambda.Start(Handler)
}
