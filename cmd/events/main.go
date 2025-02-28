package main

import (
	"context"
	"fmt"
	"os"

	lambdaEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/dynamodb/audits"
	"philcali.me/recipes/internal/dynamodb/settings"
	"philcali.me/recipes/internal/dynamodb/shares"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/dynamodb/users"
	"philcali.me/recipes/internal/events"
)

func HandleRequest(ctx context.Context, event lambdaEvents.DynamoDBEvent) error {
	tableName := os.Getenv("TABLE_NAME")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	client := dynamodb.NewFromConfig(cfg)
	marshaler := token.NewGCM()
	userData := users.NewUserService(tableName, *client, marshaler)
	auditData := audits.NewAuditService(tableName, *client, marshaler)
	shareData := shares.NewShareService(tableName, *client, marshaler)
	settingData := settings.NewSettingService(tableName, *client, marshaler)

	handlers := []events.EventFilter{
		events.DefaultUserHandler(userData),
		events.DefaultAuditHandler(auditData),
		events.DefaultDeleteAssociatedHandler(shareData),
		events.DefaultCopyApprovedRequestHandler(shareData),
		&events.CopySharingResourceHandler{
			Sharing:   shareData,
			Setting:   settingData,
			DynamoDB:  client,
			TableName: tableName,
		},
	}

	// TODO: make a router for this
	for _, record := range event.Records {
		for _, handler := range handlers {
			if handler.Filter(record) {
				err := handler.Apply(record)
				if err != nil {
					fmt.Printf("ERROR: failed to handle %s: %v", err.Error(), record)
					break
				}
			}
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
