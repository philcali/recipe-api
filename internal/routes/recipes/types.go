package recipes

import (
	"time"

	"philcali.me/recipes/internal/data"
)

type Ingredient struct {
	Name        string  `json:"name"`
	Measurement string  `json:"measurement"`
	Amount      float32 `json:"amount"`
}

type RecipeInput struct {
	Name               *string       `json:"name"`
	Instructions       *string       `json:"instructions"`
	PrepareTimeMinutes *int          `json:"prepareTimeMinutes"`
	Ingredients        *[]Ingredient `json:"ingredients"`
}

func (r *RecipeInput) ToData() data.RecipeInputDTO {
	var ingredients []data.IngredientDTO
	if r.Ingredients != nil {
		ingredients = make([]data.IngredientDTO, len(*r.Ingredients))
		for i, id := range *r.Ingredients {
			ingredients[i] = data.IngredientDTO{
				Name:        id.Name,
				Measurement: id.Measurement,
				Amount:      id.Amount,
			}
		}
	}
	return data.RecipeInputDTO{
		Name:               r.Name,
		Instructions:       r.Instructions,
		Ingredients:        &ingredients,
		PrepareTimeMinutes: r.PrepareTimeMinutes,
	}
}

type Recipe struct {
	Id                 string       `json:"recipeId"`
	Name               string       `json:"name"`
	Instructions       string       `json:"instructions"`
	PrepareTimeMinutes int          `json:"prepareTimeMinutes"`
	Ingredients        []Ingredient `json:"ingredients"`
	CreateTime         time.Time    `json:"createTime"`
	UpdateTime         time.Time    `json:"updateTime"`
}

func NewRecipe(recipe data.RecipeDTO) Recipe {
	var ingredients []Ingredient
	if recipe.Ingredients != nil {
		ingredients = make([]Ingredient, len(recipe.Ingredients))
		for i, id := range recipe.Ingredients {
			ingredients[i] = Ingredient{
				Name:        id.Name,
				Measurement: id.Measurement,
				Amount:      id.Amount,
			}
		}
	}
	return Recipe{
		Id:                 recipe.SK,
		Name:               recipe.Name,
		CreateTime:         recipe.CreateTime,
		UpdateTime:         recipe.UpdateTime,
		PrepareTimeMinutes: recipe.PrepareTimeMinutes,
		Instructions:       recipe.Instructions,
		Ingredients:        ingredients,
	}
}
