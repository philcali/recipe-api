package external

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/mealdb"
	"philcali.me/recipes/internal/provider"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type ExternalService struct {
	Service provider.RecipeProvider
}

func NewExternalService() routes.Service {
	return &ExternalService{
		Service: mealdb.NewDefaultMealClient(),
	}
}

func (es *ExternalService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/providers/mealdb":                 util.AuthorizedRoute(es.Search),
		"GET:/providers/mealdb/:mealId/recipes": util.AuthorizedRoute(es.Lookup),
		"GET:/providers/mealdb/random":          util.AuthorizedRoute(es.Random),
	}
}

func (es *ExternalService) Lookup(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	lookupId := util.RequestParam(ctx, "mealId")
	query, err := es.Service.Lookup(lookupId)
	return util.SerializeResponseOK(util.IdentityThunk, query, err)
}

func (es *ExternalService) Search(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	text, ok := event.QueryStringParameters["search"]
	if !ok {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput("Need a search parameter set")
	}
	query, err := es.Service.Search(text)
	return util.SerializeResponseOK(util.IdentityThunk, query, err)
}

func (es *ExternalService) Random(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	query, err := es.Service.Random()
	return util.SerializeResponseOK(util.IdentityThunk, query, err)
}
