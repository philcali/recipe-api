package data

import (
	"time"
)

type IngrediantDTO struct {
	Name        string `dynamodbav:"name"`
	Measurement string `dynamodbav:"measurement"`
}

type RecipeDTO struct {
	PK           string          `dynamodbav:"PK"`
	SK           string          `dynamodbav:"SK"`
	Name         string          `dynamodbav:"name"`
	Instructions string          `dynamodbav:"instructions"`
	Ingrediants  []IngrediantDTO `dynamodbav:"integrediants"`
	PrepareTime  time.Time       `dynamodbav:"prepareTime"`
	CreateTime   time.Time       `dynamodbav:"createTime"`
	UpdateTime   time.Time       `dynamodbav:"updateTime"`
}

type RecipeInputDTO struct {
	Name         *string          `dynamodbav:"name"`
	Instructions *string          `dynamodbav:"instructions"`
	Ingrediants  *[]IngrediantDTO `dynamodbav:"ingrediants"`
	PrepareTime  *time.Time       `dynamodbav:"prepareTime"`
}

type RecipeDataService interface {
	GetRecipe(accountId string, recipeId string) (RecipeDTO, error)
	CreateRecipe(accountId string, input RecipeInputDTO) (RecipeDTO, error)
	UpdateRecipe(accountId string, recipeId string, input RecipeInputDTO) (RecipeDTO, error)
	ListRecipes(accountId string, params QueryParams) (QueryResults[RecipeDTO], error)
	DeleteRecipe(accountId string, recipeId string) error
}
