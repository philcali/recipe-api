package apitokens

import (
	"time"

	"philcali.me/recipes/internal/data"
)

type ApiToken struct {
	Name       string       `json:"name"`
	Value      string       `json:"value"`
	Scopes     []data.Scope `json:"scopes"`
	ExpiresIn  *time.Time   `json:"expiresIn"`
	CreateTime time.Time    `json:"createTime"`
	UpdateTime time.Time    `json:"updateTime"`
}

type ApiTokenInput struct {
	Name      *string      `json:"name,omitempty"`
	Scopes    []data.Scope `json:"scopes,omitempty"`
	ExpiresIn *time.Time   `json:"expiresIn,omitempty"`
}
