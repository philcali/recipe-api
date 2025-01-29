package subscriptions

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

func NewSubscriptionService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.SubscriptionDTO, data.SubscriptionInputDTO] {
	return &services.RepositoryDynamoDBService[data.SubscriptionDTO, data.SubscriptionInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "Subscription",
		Shim: func(pk, sk string) data.SubscriptionDTO {
			return data.SubscriptionDTO{PK: pk, SK: sk}
		},
		OnCreate: func(sid data.SubscriptionInputDTO, createTime time.Time, pk, sk string) data.SubscriptionDTO {
			return data.SubscriptionDTO{
				PK:            pk,
				SK:            sk,
				CreateTime:    createTime,
				UpdateTime:    createTime,
				Endpoint:      *sid.Endpoint,
				Protocol:      *sid.Protocol,
				SubscriberArn: *sid.SubscriberArn,
			}
		},
	}
}
