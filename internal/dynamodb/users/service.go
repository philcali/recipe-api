package users

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

// Site Wide Users
const GLOBAL_ACCOUNT = "Global"

func NewUserService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.UserDTO, data.UserInputDTO] {
	return &services.RepositoryDynamoDBService[data.UserDTO, data.UserInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "User",
		Shim: func(pk, sk string) data.UserDTO {
			return data.UserDTO{PK: pk, SK: sk}
		},
		OnCreate: func(uid data.UserInputDTO, createTime time.Time, pk, sk string) data.UserDTO {
			return data.UserDTO{
				PK:         pk,
				SK:         sk,
				AccountId:  uid.AccountId,
				CreateTime: createTime,
				UpdateTime: createTime,
			}
		},
	}
}
