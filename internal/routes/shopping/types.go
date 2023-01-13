package shopping

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/routes/recipes"
	"philcali.me/recipes/internal/routes/util"
)

type ShoppingListInput struct {
	Name      *string               `json:"name"`
	Items     *[]recipes.Ingredient `json:"items"`
	ExpiresIn *time.Time            `json:"expiresIn"`
}

func (l *ShoppingListInput) ToData() data.ShoppingListInputDTO {
	var expiresIn int
	if l.ExpiresIn != nil {
		expiresIn = int(l.ExpiresIn.Unix())
	}
	return data.ShoppingListInputDTO{
		Name:      l.Name,
		ExpiresIn: aws.Int(expiresIn),
		Items:     util.MapOnList(l.Items, recipes.ConvertIngredientToData),
	}
}

type ShoppingList struct {
	Id         string               `json:"listId"`
	Name       string               `json:"name"`
	Items      []recipes.Ingredient `json:"ingredients"`
	ExpiresIn  *time.Time           `json:"expiresIn"`
	CreateTime time.Time            `json:"createTime"`
	UpdateTime time.Time            `json:"updateTime"`
}

func NewShoppingList(list data.ShoppingListDTO) ShoppingList {
	var expiresIn time.Time
	if list.ExpiresIn != nil {
		expiresIn = time.Unix(int64(*list.ExpiresIn), 0)
	}
	return ShoppingList{
		Id:         list.SK,
		Name:       list.Name,
		CreateTime: list.CreateTime,
		UpdateTime: list.UpdateTime,
		ExpiresIn:  &expiresIn,
		Items:      *util.MapOnList(&list.Items, recipes.ConvertIngredientDataToTransfer),
	}
}
