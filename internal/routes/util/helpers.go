package util

import (
	"context"
	"encoding/json"
	"fmt"
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

func AuthorizationClaims(event events.APIGatewayV2HTTPRequest) map[string]string {
	jwt := event.RequestContext.Authorizer.JWT
	lambda := event.RequestContext.Authorizer.Lambda
	if jwt != nil {
		return jwt.Claims
	} else if lambda != nil {
		if c, ok := lambda["jwt"]; ok {
			if v, ok := c.(map[string]interface{}); ok {
				claims := make(map[string]string, len(v))
				for k, o := range v {
					claims[k] = fmt.Sprintf("%v", o)
				}
				return claims
			}
		}
	}
	return nil
}

func AuthorizedRoute(route routes.Route) routes.Route {
	return func(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
		claims := AuthorizationClaims(event)
		if username, ok := claims["username"]; ok {
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

func _serializeList[T interface{}, I interface{}, R interface{}](repo data.Repository[T, I], thunk func(T) R, indexName *string, event events.APIGatewayV2HTTPRequest, hash string) (events.APIGatewayV2HTTPResponse, error) {
	var limit int
	var nextToken *string
	var scanIndex *string
	var err error
	var items data.QueryResults[T]
	if sLimit, ok := event.QueryStringParameters["limit"]; ok {
		l, err := strconv.Atoi(sLimit)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput("Limit parameter was not a number type.")
		}
		limit = l
	}
	if token, ok := event.QueryStringParameters["nextToken"]; ok {
		nextToken = &token
	}
	if sortOrder, ok := event.QueryStringParameters["sortOrder"]; ok {
		scanIndex = &sortOrder
	}

	if indexName == nil {
		items, err = repo.List(hash, data.QueryParams{
			Limit:     limit,
			NextToken: nextToken,
			SortOrder: scanIndex,
		})
	} else {
		items, err = repo.ListByIndex(hash, *indexName, data.QueryParams{
			Limit:     limit,
			NextToken: nextToken,
			SortOrder: scanIndex,
		})
	}

	return SerializeResponseOK(ConvertQueryResultsPartial(thunk), items, err)
}

func SerializeList[T interface{}, I interface{}, R interface{}](repo data.Repository[T, I], thunk func(T) R, event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return _serializeList(repo, thunk, nil, event, Username(ctx))
}

func SerializeListByIndex[T interface{}, I interface{}, R interface{}](repo data.Repository[T, I], thunk func(T) R, indexName string, event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return _serializeList(repo, thunk, &indexName, event, Username(ctx))
}

func SerializeListByIndexAndHash[T interface{}, I interface{}, R interface{}](repo data.Repository[T, I], thunk func(T) R, indexName string, event events.APIGatewayV2HTTPRequest, hash string) (events.APIGatewayV2HTTPResponse, error) {
	return _serializeList(repo, thunk, &indexName, event, hash)
}

func SerializeResponseNoContent(err error) (events.APIGatewayV2HTTPResponse, error) {
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 204,
	}, nil
}

func IdentityThunk[I interface{}](input I) I {
	return input
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
