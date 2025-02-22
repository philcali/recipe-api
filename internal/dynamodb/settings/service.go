package settings

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

func NewSettingService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.SettingsDTO, data.SettingsInputDTO] {
	return &services.RepositoryDynamoDBService[data.SettingsDTO, data.SettingsInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "Settings",
		Shim: func(pk, sk string) data.SettingsDTO {
			return data.SettingsDTO{PK: pk, SK: sk}
		},
		OnCreate: func(sid data.SettingsInputDTO, t time.Time, pk, sk string) data.SettingsDTO {
			return data.SettingsDTO{
				PK:               pk,
				SK:               sk,
				AutoShareLists:   *sid.AutoShareLists,
				AutoShareRecipes: *sid.AutoShareRecipes,
				CreateTime:       t,
				UpdateTime:       t,
			}
		},
		OnUpdate: func(sid data.SettingsInputDTO, ub expression.UpdateBuilder) {
			if sid.AutoShareLists != nil {
				ub.Set(expression.Name("autoShareLists"), expression.Value(sid.AutoShareLists))
			}
			if sid.AutoShareRecipes != nil {
				ub.Set(expression.Name("autoShareRecipes"), expression.Value(sid.AutoShareRecipes))
			}
		},
	}
}
