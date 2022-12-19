package util

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
)

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
