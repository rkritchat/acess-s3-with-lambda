package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"upload-file-s3/internal/config"
	"upload-file-s3/internal/file"
)

var fileService file.Service

func main() {
	cfg := config.InitConfig()
	fileService = file.NewService(cfg.AwsSession, cfg.Env)
	lambda.Start(handler)
}

func handler(_ context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case http.MethodPost:
		return fileService.UploadFile(req)
	case http.MethodGet:
		return fileService.Download(req)
	default:
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
}
