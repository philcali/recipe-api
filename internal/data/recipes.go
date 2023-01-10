package data

import (
	"time"
)

type IngredientDTO struct {
	Name        string  `dynamodbav:"name"`
	Measurement string  `dynamodbav:"measurement"`
	Amount      float32 `dynamodbav:"amount"`
}

type RecipeDTO struct {
	PK                 string          `dynamodbav:"PK"`
	SK                 string          `dynamodbav:"SK"`
	Name               string          `dynamodbav:"name"`
	Instructions       string          `dynamodbav:"instructions"`
	Ingredients        []IngredientDTO `dynamodbav:"ingredients"`
	PrepareTimeMinutes int             `dynamodbav:"prepareTimeMinutes"`
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
	Repository[RecipeDTO, RecipeInputDTO]
}
