package recipes

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
)

type Ingrediant struct {
	Name        string `json:"name"`
	Measurement string `json:"measurement"`
}

type RecipeInput struct {
	Name            *string       `json:"name"`
	Instructions    *string       `json:"instructions"`
	PreparationTime *time.Time    `json:"preparationTime"`
	Ingrediants     *[]Ingrediant `json:"ingrediants"`
}

func (r *RecipeInput) ToData() data.RecipeInputDTO {
	var ingredients []data.IngrediantDTO
	if r.Ingrediants != nil {
		for _, id := range *r.Ingrediants {
			ingredients = append(ingredients, data.IngrediantDTO{
				Name:        id.Name,
				Measurement: id.Measurement,
			})
		}
	}
	return data.RecipeInputDTO{
		Name:         r.Name,
		Instructions: r.Instructions,
		Ingrediants:  &ingredients,
		PrepareTime:  r.PreparationTime,
	}
}

type Recipe struct {
	Id           string       `json:"recipeId"`
	Name         string       `json:"name"`
	Instructions string       `json:"instructions"`
	PrepareTime  *time.Time   `json:"prepareTime"`
	Ingrediants  []Ingrediant `json:"ingrediants"`
	CreateTime   time.Time    `json:"createTime"`
	UpdateTime   time.Time    `json:"updateTime"`
}

func NewRecipe(recipe data.RecipeDTO) Recipe {
	var ingrediants []Ingrediant
	for _, id := range recipe.Ingrediants {
		ingrediants = append(ingrediants, Ingrediant{
			Name:        id.Name,
			Measurement: id.Measurement,
		})
	}
	return Recipe{
		Id:           recipe.SK,
		Name:         recipe.Name,
		CreateTime:   recipe.CreateTime,
		UpdateTime:   recipe.UpdateTime,
		PrepareTime:  &recipe.PrepareTime,
		Instructions: recipe.Instructions,
		Ingrediants:  ingrediants,
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
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	body, err := json.Marshal(items)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": strconv.Itoa(len(body)),
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func (rs *RecipeService) GetRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{}, nil
}

func (rs *RecipeService) CreateRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := RecipeInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	accountId := event.RequestContext.AccountID
	created, err := rs.data.CreateRecipe(accountId, input.ToData())
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	body, err := json.Marshal(NewRecipe(created))
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": strconv.Itoa(len(body)),
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func (rs *RecipeService) UpdateRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{}, nil
}

func (rs *RecipeService) DeleteRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{}, nil
}
