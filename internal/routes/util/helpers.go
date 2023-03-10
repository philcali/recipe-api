package util

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
)

func MapOnList[I interface{}, O interface{}](ls *[]I, thunk func(I) O) *[]O {
	var result []O
	if ls != nil {
		result = make([]O, len(*ls))
		for i, elem := range *ls {
			result[i] = thunk(elem)
		}
	}
	return &result
}

func AuthorizedRoute(route routes.Route) routes.Route {
	return func(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
		if username, ok := event.RequestContext.Authorizer.JWT.Claims["username"]; ok {
			return route(event, context.WithValue(ctx, "Username", username))
		}
		return events.APIGatewayV2HTTPResponse{}, exceptions.InternalServer("Unexpected internal error")
	}
}

func RequestParam(ctx context.Context, param string) string {
	return ctx.Value("Params").(map[string]string)[param]
}

func Username(ctx context.Context) string {
	return ctx.Value("Username").(string)
}

func SerializeResponse[T interface{}, R interface{}](delayed func(T) R, thing T, err error, statusCode int) (events.APIGatewayV2HTTPResponse, error) {
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	body, err := json.Marshal(delayed(thing))
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": strconv.Itoa(len(body)),
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func SerializeResponseOK[T interface{}, R interface{}](delayed func(T) R, thing T, err error) (events.APIGatewayV2HTTPResponse, error) {
	return SerializeResponse(delayed, thing, err, 200)
}

func SerializeList[T interface{}, I interface{}, R interface{}](repo data.Repository[T, I], thunk func(T) R, event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
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
	items, err := repo.List(Username(ctx), data.QueryParams{
		Limit:     limit,
		NextToken: nextToken,
	})
	return SerializeResponseOK(ConvertQueryResultsPartial(thunk), items, err)
}

func SerializeResponseNoContent(err error) (events.APIGatewayV2HTTPResponse, error) {
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 204,
	}, nil
}

func ConvertQueryResults[D interface{}, R interface{}](items data.QueryResults[D], thunk func(D) R) data.QueryResults[R] {
	if items.Items != nil {
		newItems := make([]R, len(items.Items))
		for i, rd := range items.Items {
			newItems[i] = thunk(rd)
		}
		return data.QueryResults[R]{
			Items:     newItems,
			NextToken: items.NextToken,
		}
	}
	return data.QueryResults[R]{
		Items: make([]R, 0),
	}
}

func ConvertQueryResultsPartial[D interface{}, R interface{}](thunk func(D) R) func(data.QueryResults[D]) data.QueryResults[R] {
	return func(d data.QueryResults[D]) data.QueryResults[R] {
		return ConvertQueryResults(d, thunk)
	}
}
