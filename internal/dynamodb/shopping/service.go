package shopping

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

type ShoppingListDynamoDBService struct {
	DynamoDB       dynamodb.Client
	TableName      string
	TokenMarshaler token.TokenMarshaler
}

func NewShoppingListDynamoDBService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.ShoppingListDTO, data.ShoppingListInputDTO] {
	return &services.RepositoryDynamoDBService[data.ShoppingListDTO, data.ShoppingListInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "ShoppingList",
		GetPK: func(sld data.ShoppingListDTO) string {
			return sld.PK
		},
		GetSK: func(sld data.ShoppingListDTO) string {
			return sld.SK
		},
		OnCreate: func(slid data.ShoppingListInputDTO, createTime time.Time, pk string, sk string) data.ShoppingListDTO {
			return data.ShoppingListDTO{
				PK:         pk,
				SK:         sk,
				CreateTime: createTime,
				UpdateTime: createTime,
				Name:       *slid.Name,
				Items:      *slid.Items,
				ExpiresIn:  slid.ExpiresIn,
			}
		},
		OnUpdate: func(slid data.ShoppingListInputDTO, ub expression.UpdateBuilder) {
			if slid.Name != nil {
				ub.Set(expression.Name("name"), expression.Value(slid.Name))
			}
			if slid.ExpiresIn != nil {
				ub.Set(expression.Name("expiresIn"), expression.Value(slid.ExpiresIn))
			}
			if slid.Items != nil {
				ub.Set(expression.Name("items"), expression.Value(slid.Items))
			}
			if slid.CompletedItems != nil {
				ub.Set(expression.Name("completedItems"), expression.Value(slid.CompletedItems))
			}
		},
		Shim: func(pk, sk string) data.ShoppingListDTO {
			return data.ShoppingListDTO{PK: pk, SK: sk}
		},
	}
}
