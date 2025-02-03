package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"log"
	"net/http"
	config2 "service-nest/config"
	"service-nest/model"
	"service-nest/repository"
	"service-nest/routers"
	"service-nest/service"
)

var gorillaLambda *gorillamux.GorillaMuxAdapter

func NewErrorNotifier(snsClient *sns.Client, topicArn string) *model.ErrorNotifier {
	return &model.ErrorNotifier{
		SnsClient: snsClient,
		TopicArn:  topicArn,
	}
}

func runApp(client *dynamodb.Client, notifier *model.ErrorNotifier) {
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

	r := routers.SetupRouter(userService, householderService, providerService, adminService, notifier)
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
func initializeDynamoDBClient() (*dynamodb.Client, *model.ErrorNotifier) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}
	snsClient := sns.NewFromConfig(cfg)

	// Create error notifier with your SNS topic ARN
	notifier := NewErrorNotifier(snsClient, config2.SNSARN)
	return dynamodb.NewFromConfig(cfg), notifier
}
func main() {
	client, notfier := initializeDynamoDBClient()

	runApp(client, notfier)
	lambda.Start(Handler)
}
