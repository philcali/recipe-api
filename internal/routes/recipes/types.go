package recipes

import (
	"time"

	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/routes/util"
)

type Ingredient struct {
	Name        string  `json:"name"`
	Measurement string  `json:"measurement"`
	Amount      float32 `json:"amount"`
}

type Nutrient struct {
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Amount int    `json:"amount"`
}

type RecipeInput struct {
	Name               *string       `json:"name"`
	Instructions       *string       `json:"instructions"`
	PrepareTimeMinutes *int          `json:"prepareTimeMinutes"`
	NumberOfServings   *int          `json:"numberOfServings"`
	Ingredients        *[]Ingredient `json:"ingredients"`
	Nutrients          *[]Nutrient   `json:"nutrients"`
}

func ConvertIngredientToData(in Ingredient) data.IngredientDTO {
	return data.IngredientDTO{
		Name:        in.Name,
		Measurement: in.Measurement,
		Amount:      in.Amount,
	}
}

func ConvertIngredientDataToTransfer(in data.IngredientDTO) Ingredient {
	return Ingredient{
		Name:        in.Name,
		Measurement: in.Measurement,
		Amount:      in.Amount,
	}
}

func (r *RecipeInput) ToData() data.RecipeInputDTO {
	return data.RecipeInputDTO{
		Name:               r.Name,
		Instructions:       r.Instructions,
		Ingredients:        util.MapOnList(r.Ingredients, ConvertIngredientToData),
		PrepareTimeMinutes: r.PrepareTimeMinutes,
		NumberOfServings:   r.NumberOfServings,
		Nutrients: util.MapOnList(r.Nutrients, func(n Nutrient) data.NutrientDTO {
			return data.NutrientDTO{
				Name:   n.Name,
				Amount: n.Amount,
				Unit:   n.Unit,
			}
		}),
	}
}

type Recipe struct {
	Id                 string       `json:"recipeId"`
	Name               string       `json:"name"`
	Instructions       string       `json:"instructions"`
	PrepareTimeMinutes *int         `json:"prepareTimeMinutes"`
	NumberOfServings   *int         `json:"numberOfServings"`
	Nutrients          []Nutrient   `json:"nutrients"`
	Ingredients        []Ingredient `json:"ingredients"`
	CreateTime         time.Time    `json:"createTime"`
	UpdateTime         time.Time    `json:"updateTime"`
}

func NewRecipe(recipe data.RecipeDTO) Recipe {
	return Recipe{
		Id:                 recipe.SK,
		Name:               recipe.Name,
		CreateTime:         recipe.CreateTime,
		UpdateTime:         recipe.UpdateTime,
		PrepareTimeMinutes: recipe.PrepareTimeMinutes,
		Instructions:       recipe.Instructions,
		NumberOfServings:   recipe.NumberOfServings,
		Ingredients:        *util.MapOnList(&recipe.Ingredients, ConvertIngredientDataToTransfer),
		Nutrients: *util.MapOnList(&recipe.Nutrients, func(nd data.NutrientDTO) Nutrient {
			return Nutrient{
				Name:   nd.Name,
				Unit:   nd.Unit,
				Amount: nd.Amount,
			}
		}),
	}
}
