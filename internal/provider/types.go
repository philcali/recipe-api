package provider

import (
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/routes/recipes"
)

type FilterInput struct {
	Category       *string
	Area           *string
	MainIngredient *string
}

type RecipeProvider interface {
	Random() (data.QueryResults[recipes.Recipe], error)
	Lookup(id string) (data.QueryResults[recipes.Recipe], error)
	Search(text string) (data.QueryResults[recipes.Recipe], error)
	Filter(input FilterInput) (data.QueryResults[recipes.Recipe], error)
}
