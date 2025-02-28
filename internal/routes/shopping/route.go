package shopping

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type ShoppingListService struct {
	data data.ShoppingListDataService
}

func NewRoute(data data.ShoppingListDataService) routes.Service {
	return &ShoppingListService{
		data: data,
	}
}

func (sl *ShoppingListService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/lists":                    util.AuthorizedRoute(sl.ListShoppingLists),
		"GET:/lists/:shoppingListId":    util.AuthorizedRoute(sl.GetShoppingList),
		"POST:/lists":                   util.AuthorizedRoute(sl.CreateShoppingList),
		"PUT:/lists/:shoppingListId":    util.AuthorizedRoute(sl.UpdateShoppingList),
		"DELETE:/lists/:shoppingListId": util.AuthorizedRoute(sl.DeleteShoppingList),
	}
}

func (sl *ShoppingListService) ListShoppingLists(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return util.SerializeList(sl.data, NewShoppingList, event, ctx)
}

func (sl *ShoppingListService) GetShoppingList(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	item, err := sl.data.Get(util.Username(ctx), util.RequestParam(ctx, "shoppingListId"))
	return util.SerializeResponseOK(NewShoppingList, item, err)
}

func (sl *ShoppingListService) CreateShoppingList(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := ShoppingListInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	claims := util.AuthorizationClaims(event)
	created, err := sl.data.Create(util.Username(ctx), input.ToData(claims["email"]))
	return util.SerializeResponseOK(NewShoppingList, created, err)
}

func (sl *ShoppingListService) UpdateShoppingList(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := ShoppingListInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	claims := util.AuthorizationClaims(event)
	item, err := sl.data.Update(util.Username(ctx), util.RequestParam(ctx, "shoppingListId"), input.ToData(claims["email"]))
	return util.SerializeResponseOK(NewShoppingList, item, err)
}

func (sl *ShoppingListService) DeleteShoppingList(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	err := sl.data.Delete(util.Username(ctx), util.RequestParam(ctx, "shoppingListId"))
	return util.SerializeResponseNoContent(err)
}
