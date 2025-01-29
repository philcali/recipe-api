package data

import "time"

type Scope string

const (
	RECIPE_READ  Scope = "recipes:read"
	RECIPE_WRITE Scope = "recipes:write"
	LIST_READ    Scope = "lists:read"
	LIST_WRITE   Scope = "lists:write"
)

type ApiTokenDTO struct {
	PK         string    `dynamodbav:"PK"`
	SK         string    `dynamodbav:"SK"`
	FirstIndex string    `dynamodbav:"GS1-PK"`
	AccountId  string    `dynamodbav:"accountId"`
	Name       string    `dynamodbav:"name"`
	Scopes     []Scope   `dynamodbav:"scopes"`
	ExpiresIn  *int      `dynamodbav:"expiresIn"`
	CreateTime time.Time `dynamodbav:"createTime"`
	UpdateTime time.Time `dynamodbav:"updateTime"`
}

type ApiTokenInputDTO struct {
	Name      *string  `dynamodbav:"name"`
	Scopes    *[]Scope `dynamodbav:"scopes"`
	AccountId *string  `dynamodbav:"accountId"`
	ExpiresIn *int     `dynamodbav:"expiresIn"`
}

type ApiTokenDataService interface {
	Repository[ApiTokenDTO, ApiTokenInputDTO]
}
