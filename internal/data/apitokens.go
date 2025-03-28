package data

import "time"

type Scope string

const (
	RECIPE_READ         Scope = "recipes.readonly"
	RECIPE_WRITE        Scope = "recipes"
	LIST_READ           Scope = "lists.readonly"
	LIST_WRITE          Scope = "lists"
	SETTINGS_READ       Scope = "settings.readonly"
	SETTINGS_WRITE      Scope = "settings"
	SHARE_READ          Scope = "shares.readonly"
	SHARE_WRITE         Scope = "shares"
	AUDIT_READ          Scope = "audits.readonly"
	AUDIT_WRITE         Scope = "audits"
	SUBSCRIPTIONS_READ  Scope = "subscriptions.readonly"
	SUBSCRIPTIONS_WRITE Scope = "subscriptions"
	TOKENS_READ         Scope = "tokens.readonly"
	TOKENS_WRITE        Scope = "tokens"
	PROVIDER_READ       Scope = "providers.readonly"
	PROVIDER_WRITE      Scope = "providers"
)

type ApiTokenDTO struct {
	PK         string            `dynamodbav:"PK"`
	SK         string            `dynamodbav:"SK"`
	FirstIndex string            `dynamodbav:"GS1-PK"`
	AccountId  string            `dynamodbav:"accountId"`
	Name       string            `dynamodbav:"name"`
	Claims     map[string]string `dynamodbav:"claims"`
	Scopes     []Scope           `dynamodbav:"scopes"`
	ExpiresIn  *int              `dynamodbav:"expiresIn"`
	CreateTime time.Time         `dynamodbav:"createTime"`
	UpdateTime time.Time         `dynamodbav:"updateTime"`
}

type ApiTokenInputDTO struct {
	Name      *string            `dynamodbav:"name"`
	Scopes    *[]Scope           `dynamodbav:"scopes"`
	Claims    *map[string]string `dynamodbav:"map"`
	AccountId *string            `dynamodbav:"accountId"`
	ExpiresIn *int               `dynamodbav:"expiresIn"`
}

type ApiTokenDataService interface {
	Repository[ApiTokenDTO, ApiTokenInputDTO]
}
