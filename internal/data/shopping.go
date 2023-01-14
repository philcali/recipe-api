package data

import "time"

type ShoppingListItemDTO struct {
	Name        string  `dynamodbav:"name"`
	Measurement string  `dynamodbav:"measurement"`
	Amount      float32 `dynamodbav:"amount"`
	Completed   bool    `dynamodbav:"completed"`
}

type ShoppingListDTO struct {
	PK         string                `dynamodbav:"PK"`
	SK         string                `dynamodbav:"SK"`
	Name       string                `dynamodbav:"name"`
	Items      []ShoppingListItemDTO `dynamodbav:"items"`
	ExpiresIn  *int                  `dynamodbav:"expiresIn"`
	CreateTime time.Time             `dynamodbav:"createTime"`
	UpdateTime time.Time             `dynamodbav:"updateTime"`
}

type ShoppingListInputDTO struct {
	Name      *string                `dynamodbav:"name"`
	Items     *[]ShoppingListItemDTO `dynamodbav:"items"`
	ExpiresIn *int                   `dynamodbav:"expiresIn"`
}

type ShoppingListDataService interface {
	Repository[ShoppingListDTO, ShoppingListInputDTO]
}
