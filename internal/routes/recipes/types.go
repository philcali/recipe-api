package recipes

import (
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
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
	Type               *string       `json:"type"`
	Thumbnail          *string       `json:"thumbnail"`
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

func (r *RecipeInput) ToData(owner string) data.RecipeInputDTO {
	return data.RecipeInputDTO{
		Name:               r.Name,
		Instructions:       r.Instructions,
		Ingredients:        util.MapOnList(r.Ingredients, ConvertIngredientToData),
		PrepareTimeMinutes: r.PrepareTimeMinutes,
		NumberOfServings:   r.NumberOfServings,
		Thumbnail:          r.Thumbnail,
		Type:               r.Type,
		Owner:              &owner,
		UpdateToken:        aws.String(uuid.NewString()),
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
	Thumbnail          *string      `json:"thumbnail"`
	Type               *string      `json:"type"`
	Owner              *string      `json:"email"`
	Nutrients          []Nutrient   `json:"nutrients"`
	Ingredients        []Ingredient `json:"ingredients"`
	CreateTime         time.Time    `json:"createTime"`
	UpdateTime         time.Time    `json:"updateTime"`
}

func StripFields(event events.APIGatewayV2HTTPRequest) func(data.RecipeDTO) Recipe {
	var stripThumbnail bool
	if stripFields, ok := event.QueryStringParameters["stripFields"]; ok {
		fields := strings.Split(stripFields, ",")
		for _, field := range fields {
			if strings.EqualFold(field, "thumbnail") {
				stripThumbnail = true
			}
		}
	}
	return func(rd data.RecipeDTO) Recipe {
		return NewRecipe(rd, stripThumbnail)
	}
}

func NewRecipe(recipe data.RecipeDTO, stripThumbail bool) Recipe {
	var thumbnail *string
	if !stripThumbail {
		thumbnail = recipe.Thumbnail
	}
	return Recipe{
		Id:                 recipe.SK,
		Name:               recipe.Name,
		CreateTime:         recipe.CreateTime,
		UpdateTime:         recipe.UpdateTime,
		PrepareTimeMinutes: recipe.PrepareTimeMinutes,
		Instructions:       recipe.Instructions,
		NumberOfServings:   recipe.NumberOfServings,
		Owner:              recipe.Owner,
		Thumbnail:          thumbnail,
		Type:               recipe.Type,
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
