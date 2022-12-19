package recipes

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

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
