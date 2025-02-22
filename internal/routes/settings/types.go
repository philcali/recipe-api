package settings

import "time"

type Settings struct {
	AutoShareLists   bool      `json:"autoShareLists"`
	AutoShareRecipes bool      `json:"autoShareRecipes"`
	CreateTime       time.Time `json:"createTime"`
	UpdateTime       time.Time `json:"updateTime"`
}

type SettingsInput struct {
	AutoShareLists   *bool `json:"autoShareLists"`
	AutoShareRecipes *bool `json:"autoShareRecipes"`
}
