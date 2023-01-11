package recipes

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

type RecipeDynamoDBService struct {
	DynamoDB       dynamodb.Client
	TableName      string
	TokenMarshaler token.TokenMarshaler
}

func NewRecipeService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.RecipeDTO, data.RecipeInputDTO] {
	return &services.RepositoryDynamoDBService[data.RecipeDTO, data.RecipeInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "Recipe",
		GetPK: func(rd data.RecipeDTO) string {
			return rd.PK
		},
		GetSK: func(rd data.RecipeDTO) string {
			return rd.SK
		},
		Shim: func(pk, sk string) data.RecipeDTO {
			return data.RecipeDTO{PK: pk, SK: sk}
		},
		OnCreate: func(input data.RecipeInputDTO, now time.Time, pk, sk string) data.RecipeDTO {
			return data.RecipeDTO{
				PK:                 pk,
				SK:                 sk,
				Name:               *input.Name,
				Instructions:       *input.Instructions,
				Ingredients:        *input.Ingredients,
				PrepareTimeMinutes: input.PrepareTimeMinutes,
				CreateTime:         now,
				UpdateTime:         now,
			}
		},
		OnUpdate: func(input data.RecipeInputDTO, update expression.UpdateBuilder) {
			if input.Name != nil {
				update.Set(expression.Name("name"), expression.Value(input.Name))
			}
			if input.Instructions != nil {
				update.Set(expression.Name("instructions"), expression.Value(input.Instructions))
			}
			if input.Ingredients != nil {
				update.Set(expression.Name("ingredients"), expression.Value(input.Ingredients))
			}
			if input.PrepareTimeMinutes != nil {
				update.Set(expression.Name("prepareTimeMinutes"), expression.Value(input.PrepareTimeMinutes))
			}
		},
	}
}
