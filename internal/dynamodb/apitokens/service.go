package apitokens

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

func NewApiTokenService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.ApiTokenDTO, data.ApiTokenInputDTO] {
	return &services.RepositoryDynamoDBService[data.ApiTokenDTO, data.ApiTokenInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "ApiToken",
		Shim: func(pk, sk string) data.ApiTokenDTO {
			return data.ApiTokenDTO{PK: pk, SK: sk}
		},
		OnCreate: func(atid data.ApiTokenInputDTO, t time.Time, pk, sk string) data.ApiTokenDTO {
			return data.ApiTokenDTO{
				PK:         pk,
				SK:         sk,
				FirstIndex: fmt.Sprintf("%s:ApiToken", *atid.AccountId),
				Name:       *atid.Name,
				ExpiresIn:  atid.ExpiresIn,
				Scopes:     *atid.Scopes,
				AccountId:  *atid.AccountId,
				CreateTime: t,
				UpdateTime: t,
			}
		},
		OnUpdate: func(atid data.ApiTokenInputDTO, ub expression.UpdateBuilder) {
			if atid.Name != nil {
				ub.Set(expression.Name("name"), expression.Value(atid.Name))
			}
			if atid.ExpiresIn != nil {
				ub.Set(expression.Name("expiresIn"), expression.Value(atid.ExpiresIn))
			}
		},
	}
}
