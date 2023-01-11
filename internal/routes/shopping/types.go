package shopping

import (
	"time"

	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/routes/recipes"
	"philcali.me/recipes/internal/routes/util"
)

type ShoppingListInput struct {
	Name      *string               `json:"name"`
	Items     *[]recipes.Ingredient `json:"items"`
	ExpiresIn *int                  `json:"expiresIn"`
}

func (l *ShoppingListInput) ToData() data.ShoppingListInputDTO {
	return data.ShoppingListInputDTO{
		Name:      l.Name,
		ExpiresIn: l.ExpiresIn,
		Items:     util.MapOnList(l.Items, recipes.ConvertIngredientToData),
	}
}

type ShoppingList struct {
	Id         string               `json:"listId"`
	Name       string               `json:"name"`
	Items      []recipes.Ingredient `json:"ingredients"`
	ExpiresIn  *int                 `json:"expiresIn"`
	CreateTime time.Time            `json:"createTime"`
	UpdateTime time.Time            `json:"updateTime"`
}

func NewShoppingList(list data.ShoppingListDTO) ShoppingList {
	return ShoppingList{
		Id:         list.SK,
		Name:       list.Name,
		CreateTime: list.CreateTime,
		UpdateTime: list.UpdateTime,
		ExpiresIn:  list.ExpiresIn,
		Items:      *util.MapOnList(&list.Items, recipes.ConvertIngredientDataToTransfer),
	}
}
