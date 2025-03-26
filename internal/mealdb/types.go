package mealdb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"philcali.me/recipes/internal/routes/recipes"
)

type Meal struct {
	Id           string `json:"idMeal"`
	Name         string `json:"strMeal"`
	Category     string `json:"strCategory"`
	Area         string `json:"strArea"`
	Instructions string `json:"strInstructions"`
	Thumbnail    string `json:"strMealThumb"`
	Source       string `json:"strSource"`
	Ingredient1  string `json:"strIngredient1"`
	Ingredient2  string `json:"strIngredient2"`
	Ingredient3  string `json:"strIngredient3"`
	Ingredient4  string `json:"strIngredient4"`
	Ingredient5  string `json:"strIngredient5"`
	Ingredient6  string `json:"strIngredient6"`
	Ingredient7  string `json:"strIngredient7"`
	Ingredient8  string `json:"strIngredient8"`
	Ingredient9  string `json:"strIngredient9"`
	Ingredient10 string `json:"strIngredient10"`
	Ingredient11 string `json:"strIngredient11"`
	Ingredient12 string `json:"strIngredient12"`
	Ingredient13 string `json:"strIngredient13"`
	Ingredient14 string `json:"strIngredient14"`
	Ingredient15 string `json:"strIngredient15"`
	Ingredient16 string `json:"strIngredient16"`
	Ingredient17 string `json:"strIngredient17"`
	Ingredient18 string `json:"strIngredient18"`
	Ingredient19 string `json:"strIngredient19"`
	Ingredient20 string `json:"strIngredient20"`
	Measure1     string `json:"strMeasure1"`
	Measure2     string `json:"strMeasure2"`
	Measure3     string `json:"strMeasure3"`
	Measure4     string `json:"strMeasure4"`
	Measure5     string `json:"strMeasure5"`
	Measure6     string `json:"strMeasure6"`
	Measure7     string `json:"strMeasure7"`
	Measure8     string `json:"strMeasure8"`
	Measure9     string `json:"strMeasure9"`
	Measure10    string `json:"strMeasure10"`
	Measure11    string `json:"strMeasure11"`
	Measure12    string `json:"strMeasure12"`
	Measure13    string `json:"strMeasure13"`
	Measure14    string `json:"strMeasure14"`
	Measure15    string `json:"strMeasure15"`
	Measure16    string `json:"strMeasure16"`
	Measure17    string `json:"strMeasure17"`
	Measure18    string `json:"strMeasure18"`
	Measure19    string `json:"strMeasure19"`
	Measure20    string `json:"strMeasure20"`
}

func ToRecipe(m Meal) recipes.Recipe {
	ingredients := make([]recipes.Ingredient, 0)
	body, err := json.Marshal(m)
	if err == nil {
		var bagOfStrings map[string]string
		if err := json.Unmarshal(body, &bagOfStrings); err == nil {
			for i := 1; i <= 20; i++ {
				name := strings.TrimSpace(bagOfStrings[fmt.Sprintf("strIngredient%d", i)])
				measurement := strings.TrimSpace(bagOfStrings[fmt.Sprintf("strMeasure%d", i)])
				if name == "" {
					continue
				}
				amount := float32(1.0)
				if measurement == "To taste" || measurement == "To serve" {
					measurement = "whole"
				} else {
					pv, pm, found := strings.Cut(measurement, " ")
					if found {
						value, err := strconv.Atoi(pv)
						if err == nil {
							amount = float32(value)
						} else if pv == "½" {
							amount = 0.5
						} else if strings.Contains(pv, "/") {
							n := strings.Split(pv, "/")
							numerator, nerr := strconv.Atoi(n[0])
							denominator, derr := strconv.Atoi(n[1])
							if nerr == nil && derr == nil {
								amount = float32(numerator) / float32(denominator)
							}
						}
						measurement = pm
					} else {
						measurement = "whole"
					}
				}

				ingredients = append(ingredients, recipes.Ingredient{
					Name:        name,
					Amount:      amount,
					Measurement: measurement,
				})
			}
		}
	}
	return recipes.Recipe{
		Id:           m.Id,
		Name:         m.Name,
		Instructions: m.Instructions,
		Thumbnail:    &m.Thumbnail,
		Ingredients:  ingredients,
	}
}

func ConvertCategoryToRecipe(m FilteredMeal) recipes.Recipe {
	return recipes.Recipe{
		Id:        m.Id,
		Name:      m.Name,
		Thumbnail: &m.Thumbnail,
	}
}

type FilteredMeal struct {
	Id        string `json:"idMeal"`
	Name      string `json:"strMeal"`
	Thumbnail string `json:"strMealThumb"`
}

type Category struct {
	Id          string `json:"idCategory"`
	Name        string `json:"strCategory"`
	Thumbnail   string `json:"strCategoryThumb"`
	Description string `json:"strCategoryDescription"`
}

type CategoryResponse struct {
	Categories []Category `json:"categories"`
}

type QueryResponse struct {
	Meals []Meal `json:"meals"`
}

type FilterResponse struct {
	Meals []FilteredMeal `json:"meals"`
}
