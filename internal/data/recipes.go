package data

import (
	"time"
)

type IngredientDTO struct {
	Name        string `dynamodbav:"name"`
	Measurement string `dynamodbav:"measurement"`
}

type RecipeDTO struct {
	PK                 string          `dynamodbav:"PK"`
	SK                 string          `dynamodbav:"SK"`
	Name               string          `dynamodbav:"name"`
	Instructions       string          `dynamodbav:"instructions"`
	Ingredients        []IngredientDTO `dynamodbav:"ingredients"`
	PrepareTimeMinutes *int            `dynamodbav:"prepareTimeMinutes"`
	CreateTime         time.Time       `dynamodbav:"createTime"`
	UpdateTime         time.Time       `dynamodbav:"updateTime"`
}

type RecipeInputDTO struct {
	Name               *string          `dynamodbav:"name"`
	Instructions       *string          `dynamodbav:"instructions"`
	Ingredients        *[]IngredientDTO `dynamodbav:"ingredients"`
	PrepareTimeMinutes *int             `dynamodbav:"prepareTimeMinutes"`
}

type RecipeDataService interface {
	GetRecipe(accountId string, recipeId string) (RecipeDTO, error)
	CreateRecipe(accountId string, input RecipeInputDTO) (RecipeDTO, error)
	UpdateRecipe(accountId string, recipeId string, input RecipeInputDTO) (RecipeDTO, error)
	ListRecipes(accountId string, params QueryParams) (QueryResults[RecipeDTO], error)
	DeleteRecipe(accountId string, recipeId string) error
}
