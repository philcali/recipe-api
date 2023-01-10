package recipes

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

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
		"GET:/recipes":              util.AuthorizedRoute(rs.ListRecipes),
		"GET:/recipes/:recipeId":    util.AuthorizedRoute(rs.GetRecipe),
		"POST:/recipes":             util.AuthorizedRoute(rs.CreateRecipe),
		"PUT:/recipes/:recipeId":    util.AuthorizedRoute(rs.UpdateRecipe),
		"DELETE:/recipes/:recipeId": util.AuthorizedRoute(rs.DeleteRecipe),
	}
}

func (rs *RecipeService) ListRecipes(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return util.SerializeList[data.RecipeDTO, data.RecipeInputDTO](rs.data, NewRecipe, event, ctx)
}

func (rs *RecipeService) GetRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	item, err := rs.data.Get(util.Username(ctx), util.RequestParam(ctx, "recipeId"))
	return util.SerializeResponseOK(NewRecipe, item, err)
}

func (rs *RecipeService) CreateRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := RecipeInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	created, err := rs.data.Create(util.Username(ctx), input.ToData())
	return util.SerializeResponseOK(NewRecipe, created, err)
}

func (rs *RecipeService) UpdateRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := RecipeInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	item, err := rs.data.Update(util.Username(ctx), util.RequestParam(ctx, "recipeId"), input.ToData())
	return util.SerializeResponseOK(NewRecipe, item, err)
}

func (rs *RecipeService) DeleteRecipe(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	err := rs.data.Delete(util.Username(ctx), util.RequestParam(ctx, "recipeId"))
	return util.SerializeResponseNoContent(err)
}
