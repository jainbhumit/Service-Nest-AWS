package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	appconfig "service-nest/config"
	"service-nest/internal/app"
	"service-nest/util"
)

var gorillaLambda *gorillamux.GorillaMuxAdapter

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

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if err := util.InitJWT(cfg.JWTSecret); err != nil {
		log.Fatalf("jwt init failed: %v", err)
	}

	application, err := app.Bootstrap(cfg)
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	log.Printf("service-nest lambda starting: %s", cfg.String())
	gorillaLambda = gorillamux.New(application.Router)
	lambda.Start(Handler)
}
