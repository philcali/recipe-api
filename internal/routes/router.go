package routes

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/exceptions"
)

type Route func(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error)

type Service interface {
	GetRoutes() map[string]Route
}

type Router struct {
	Routes map[string]Route
}

func NewRouter(services ...Service) *Router {
	var routes map[string]Route = make(map[string]Route)
	for _, service := range services {
		for composite, route := range service.GetRoutes() {
			routes[composite] = route
		}
	}
	return &Router{
		Routes: routes,
	}
}

func translateError(err error) events.APIGatewayV2HTTPResponse {
	statusCode := 500
	if re, ok := err.(exceptions.RequestError); ok {
		statusCode = re.ToServiceError().StatusCode
	}
	if se, ok := err.(*exceptions.ServiceError); ok {
		statusCode = se.StatusCode
	}
	body := "{\"message\": \"" + err.Error() + "\"}"
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": strconv.Itoa(len(body)),
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    headers,
	}
}

func (r *Router) Invoke(event events.APIGatewayV2HTTPRequest, ctx context.Context) events.APIGatewayV2HTTPResponse {
	for composite, route := range r.Routes {
		parts := strings.SplitN(composite, ":", 2)
		method := parts[0]
		path := parts[1]
		if event.RequestContext.HTTP.Method != method {
			continue
		}
		regexPath := strings.ReplaceAll(path, ":id", "[^/]+")
		reg, err := regexp.Compile(regexPath)
		if err != nil {
			return translateError(err)
		}
		if reg.MatchString(event.RawPath) {
			resp, err := route(event, ctx)
			if err != nil {
				return translateError(err)
			}
			return resp
		}
	}
	return translateError(exceptions.NotFound("route", event.RawPath))
}
