package shopping

import (
	"time"

	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/routes/recipes"
	"philcali.me/recipes/internal/routes/util"
)

type ShoppingListInput struct {
	Name  *string               `json:"name"`
	Items *[]recipes.Ingredient `json:"items"`
}

func (l *ShoppingListInput) ToData() data.ShoppingListInputDTO {
	return data.ShoppingListInputDTO{
		Name:  l.Name,
		Items: util.MapOnList(l.Items, recipes.ConvertIngredientToData),
	}
}

type ShoppingList struct {
	Id         string               `json:"shoppingListId"`
	Name       string               `json:"name"`
	Items      []recipes.Ingredient `json:"ingredients"`
	CreateTime time.Time            `json:"createTime"`
	UpdateTime time.Time            `json:"updateTime"`
}

func NewShoppingList(list data.ShoppingListDTO) ShoppingList {
	return ShoppingList{
		Id:         list.SK,
		Name:       list.Name,
		CreateTime: list.CreateTime,
		UpdateTime: list.UpdateTime,
		Items:      *util.MapOnList(&list.Items, recipes.ConvertIngredientDataToTransfer),
	}
}
