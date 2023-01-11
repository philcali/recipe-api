package data

import "time"

type ShoppingListDTO struct {
	PK         string          `dynamodbav:"PK"`
	SK         string          `dynamodbav:"SK"`
	Name       string          `dynamodbav:"name"`
	Items      []IngredientDTO `dynamodbav:"ingredients"`
	ExpiresIn  *int            `dynamodbav:"expiresIn"`
	CreateTime time.Time       `dynamodbav:"createTime"`
	UpdateTime time.Time       `dynamodbav:"updateTime"`
}

type ShoppingListInputDTO struct {
	Name      *string          `dynamodbav:"name"`
	Items     *[]IngredientDTO `dynamodbav:"ingredients"`
	ExpiresIn *int             `dynamodbav:"expiresIn"`
}

type ShoppingListDataService interface {
	Repository[ShoppingListDTO, ShoppingListInputDTO]
}
