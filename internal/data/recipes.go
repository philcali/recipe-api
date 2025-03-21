package data

import (
	"time"
)

type IngredientDTO struct {
	Name        string  `dynamodbav:"name"`
	Measurement string  `dynamodbav:"measurement"`
	Amount      float32 `dynamodbav:"amount"`
}

type NutrientDTO struct {
	Name   string `dynamodbav:"name"`
	Unit   string `dynamodbav:"unit"`
	Amount int    `dynamodbav:"amount"`
}

type RecipeDTO struct {
	PK                 string          `dynamodbav:"PK"`
	SK                 string          `dynamodbav:"SK"`
	Name               string          `dynamodbav:"name"`
	Instructions       string          `dynamodbav:"instructions"`
	Owner              *string         `dynamodbav:"owner"`
	UpdateToken        *string         `dynamodbav:"updateToken"`
	Shared             *bool           `dynamodbav:"shared"`
	Thumbnail          *string         `dynamodbav:"thumbnail"`
	Type               *string         `dynamodbav:"type"`
	Ingredients        []IngredientDTO `dynamodbav:"ingredients"`
	Nutrients          []NutrientDTO   `dynamodbav:"nutrients"`
	PrepareTimeMinutes *int            `dynamodbav:"prepareTimeMinutes"`
	NumberOfServings   *int            `dynamodbav:"numberOfServings"`
	CreateTime         time.Time       `dynamodbav:"createTime"`
	UpdateTime         time.Time       `dynamodbav:"updateTime"`
}

type RecipeInputDTO struct {
	Name               *string          `dynamodbav:"name"`
	Owner              *string          `dynamodbav:"owner"`
	UpdateToken        *string          `dynamodbav:"updateToken"`
	Instructions       *string          `dynamodbav:"instructions"`
	Thumbnail          *string          `dynamodbav:"thumbnail"`
	Type               *string          `dynamodbav:"type"`
	Ingredients        *[]IngredientDTO `dynamodbav:"ingredients"`
	Nutrients          *[]NutrientDTO   `dynamodbav:"nutrients"`
	PrepareTimeMinutes *int             `dynamodbav:"prepareTimeMinutes"`
	NumberOfServings   *int             `dynamodbav:"numberOfServings"`
}

type RecipeDataService interface {
	Repository[RecipeDTO, RecipeInputDTO]
}
