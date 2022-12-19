package recipes

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type Ingredient struct {
	Name        string `json:"name"`
	Measurement string `json:"measurement"`
}

type RecipeInput struct {
	Name            *string       `json:"name"`
	Instructions    *string       `json:"instructions"`
	PreparationTime *int          `json:"preparationTime"`
	Ingredients     *[]Ingredient `json:"ingredients"`
}

func (r *RecipeInput) ToData() data.RecipeInputDTO {
	var ingredients []data.IngredientDTO
	if r.Ingredients != nil {
		ingredients = make([]data.IngredientDTO, len(*r.Ingredients))
		for i, id := range *r.Ingredients {
			ingredients[i] = data.IngredientDTO{
				Name:        id.Name,
				Measurement: id.Measurement,
			}
		}
	}
	return data.RecipeInputDTO{
		Name:               r.Name,
		Instructions:       r.Instructions,
		Ingredients:        &ingredients,
		PrepareTimeMinutes: r.PreparationTime,
	}
}

type Recipe struct {
	Id           string       `json:"recipeId"`
	Name         string       `json:"name"`
	Instructions string       `json:"instructions"`
	PrepareTime  *int         `json:"prepareTimeMinutes"`
	Ingredients  []Ingredient `json:"ingredients"`
	CreateTime   time.Time    `json:"createTime"`
	UpdateTime   time.Time    `json:"updateTime"`
}

func NewRecipe(recipe data.RecipeDTO) Recipe {
	var ingredients []Ingredient
	if recipe.Ingredients != nil {
		ingredients = make([]Ingredient, len(recipe.Ingredients))
		for i, id := range recipe.Ingredients {
			ingredients[i] = Ingredient{
				Name:        id.Name,
				Measurement: id.Measurement,
			}
		}
	}
	return Recipe{
		Id:           recipe.SK,
		Name:         recipe.Name,
		CreateTime:   recipe.CreateTime,
		UpdateTime:   recipe.UpdateTime,
		PrepareTime:  recipe.PrepareTimeMinutes,
		Instructions: recipe.Instructions,
		Ingredients:  ingredients,
	}
}

type RecipeService struct {
	data data.RecipeDataService
}

func NewRoute(data data.RecipeDataService) routes.Service {
	return &RecipeService{
		data: data,
	}
}

func _getRecipeId(event events.APIGatewayV2HTTPRequest) string {
	parts := strings.Split(event.RawPath, "/")
	return parts[len(parts)-1]
}

func (rs *RecipeService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/recipes":        rs.ListRecipes,
		"GET:/recipes/:id":    rs.GetRecipe,
		"POST:/recipes":       rs.CreateRecipe,
		"PUT:/recipes/:id":    rs.UpdateRecipe,
		"DELETE:/recipes/:id": rs.DeleteRecipe,
	}
}

func (rs *RecipeService) ListRecipes(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	var limit int
	var nextToken []byte
	var err error
	if sLimit, ok := event.QueryStringParameters["limit"]; ok {
		if limit, err = strconv.Atoi(sLimit); err != nil {
			return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput("Limit parameter was not a number type.")
		}
	}
	if token, ok := event.QueryStringParameters["nextToken"]; ok {
		nextToken = []byte(token)
	}
	items, err := rs.data.ListRecipes(event.RequestContext.AccountID, data.QueryParams{
		Limit:     limit,
		NextToken: nextToken,
	})
	return util.SerializeResponseOK(util.ConvertQueryResultsPartial(NewRecipe), items, err)
}

func (rs *RecipeService) GetRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	item, err := rs.data.GetRecipe(event.RequestContext.AccountID, _getRecipeId(event))
	return util.SerializeResponseOK(NewRecipe, item, err)
}

func (rs *RecipeService) CreateRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := RecipeInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	accountId := event.RequestContext.AccountID
	created, err := rs.data.CreateRecipe(accountId, input.ToData())
	return util.SerializeResponseOK(NewRecipe, created, err)
}

func (rs *RecipeService) UpdateRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := RecipeInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	item, err := rs.data.UpdateRecipe(event.RequestContext.AccountID, _getRecipeId(event), input.ToData())
	return util.SerializeResponseOK(NewRecipe, item, err)
}

func (rs *RecipeService) DeleteRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	err := rs.data.DeleteRecipe(event.RequestContext.AccountID, _getRecipeId(event))
	return util.SerializeResponseNoContent(err)
}
