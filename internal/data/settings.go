package data

import "time"

type SettingsDTO struct {
	AutoShareLists   bool      `dynamodbav:"autoShareLists"`
	AutoShareRecipes bool      `dynamodbav:"autoShareRecipes"`
	PK               string    `dynamodbav:"PK"`
	SK               string    `dynamodbav:"SK"`
	CreateTime       time.Time `dynamodbav:"createTime"`
	UpdateTime       time.Time `dynamodbav:"updateTime"`
}

type SettingsInputDTO struct {
	AutoShareLists   *bool `dynamodbav:"autoShareLists"`
	AutoShareRecipes *bool `dynamodbav:"autoShareRecipes"`
}

type SettingsRepository interface {
	Repository[SettingsDTO, SettingsInputDTO]
}
