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

type CachedRoute struct {
	Method     string
	Matcher    *regexp.Regexp
	ParamNames []string
	Route      Route
}

type Router struct {
	Routes []CachedRoute
}

func NewRouter(services ...Service) (*Router, error) {
	var routes []CachedRoute
	for _, service := range services {
		for composite, route := range service.GetRoutes() {
			var names []string
			parts := strings.SplitN(composite, ":", 2)
			method := parts[0]
			path := parts[1]
			namex := regexp.MustCompile(":[^/]+")
			regexPath := namex.ReplaceAllStringFunc(path, func(found string) string {
				names = append(names, found[1:])
				return "([^/]+)"
			})
			reg, err := regexp.Compile("^" + regexPath + "$")
			if err != nil {
				return nil, err
			}
			cachedRoute := CachedRoute{
				Method:     method,
				Matcher:    reg,
				ParamNames: names,
				Route:      route,
			}
			routes = append(routes, cachedRoute)
		}
	}
	return &Router{
		Routes: routes,
	}, nil
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
	for _, route := range r.Routes {
		if route.Method != event.RequestContext.HTTP.Method {
			continue
		}
		if values := route.Matcher.FindAllStringSubmatchIndex(event.RawPath, -1); values != nil {
			params := make(map[string]string, len(route.ParamNames))
			for i, p := range route.ParamNames {
				params[p] = event.RawPath[values[0][i+2]:values[0][i+3]]
			}
			resp, err := route.Route(event, context.WithValue(ctx, "Params", params))
			if err != nil {
				return translateError(err)
			}
			return resp
		}
	}
	return translateError(exceptions.NotFound("route", event.RawPath))
}
