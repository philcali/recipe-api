package shopping

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/routes/util"
)

type ShoppingListItem struct {
	Name        string  `json:"name"`
	Measurement string  `json:"measurement"`
	Amount      float32 `json:"amount"`
	Completed   bool    `json:"completed"`
}

type ShoppingListInput struct {
	Name      *string             `json:"name,omitempty"`
	Items     *[]ShoppingListItem `json:"items,omitempty"`
	ExpiresIn *time.Time          `json:"expiresIn,omitempty"`
}

func (l *ShoppingListInput) ToData(owner string) data.ShoppingListInputDTO {
	var expiresIn *int = nil
	if l.ExpiresIn != nil {
		expiresIn = aws.Int(int(l.ExpiresIn.Unix()))
	}
	return data.ShoppingListInputDTO{
		Name:        l.Name,
		ExpiresIn:   expiresIn,
		Owner:       &owner,
		UpdateToken: aws.String(uuid.NewString()),
		Items: util.MapOnList(l.Items, func(sli ShoppingListItem) data.ShoppingListItemDTO {
			return data.ShoppingListItemDTO{
				Name:        sli.Name,
				Measurement: sli.Measurement,
				Amount:      sli.Amount,
				Completed:   sli.Completed,
			}
		}),
	}
}

type ShoppingList struct {
	Id         string             `json:"listId"`
	Name       string             `json:"name"`
	Owner      *string            `json:"owner"`
	Items      []ShoppingListItem `json:"items"`
	ExpiresIn  *time.Time         `json:"expiresIn,omitempty"`
	CreateTime time.Time          `json:"createTime"`
	UpdateTime time.Time          `json:"updateTime"`
}

func NewShoppingList(list data.ShoppingListDTO) ShoppingList {
	var expiresIn *time.Time = nil
	if list.ExpiresIn != nil {
		expiresIn = aws.Time(time.Unix(int64(*list.ExpiresIn), 0))
	}
	return ShoppingList{
		Id:         list.SK,
		Name:       list.Name,
		CreateTime: list.CreateTime,
		UpdateTime: list.UpdateTime,
		ExpiresIn:  expiresIn,
		Owner:      list.Owner,
		Items: *util.MapOnList(&list.Items, func(slid data.ShoppingListItemDTO) ShoppingListItem {
			return ShoppingListItem{
				Name:        slid.Name,
				Measurement: slid.Measurement,
				Amount:      slid.Amount,
				Completed:   slid.Completed,
			}
		}),
	}
}
