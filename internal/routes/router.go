package routes

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/exceptions"
)

type Route func(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error)

type Service interface {
	GetRoutes() map[string]Route
}

type CachedMatcher struct {
	Matcher    *regexp.Regexp
	ParamNames []string
	Mutex      *sync.Mutex
}

type CachedRoute struct {
	Method  string
	Path    string
	Route   Route
	Matcher *CachedMatcher
}

func (cr *CachedMatcher) Refresh(path string) *regexp.Regexp {
	cr.Mutex.Lock()
	defer cr.Mutex.Unlock()
	if cr.Matcher == nil {
		namex := regexp.MustCompile(":[^/]+")
		regexPath := namex.ReplaceAllStringFunc(path, func(found string) string {
			cr.ParamNames = append(cr.ParamNames, found[1:])
			return "([^/]+)"
		})
		cr.Matcher = regexp.MustCompile("^" + regexPath + "$")
	}
	return cr.Matcher
}

func (cr *CachedRoute) MatchEvent(event events.APIGatewayV2HTTPRequest) (map[string]string, bool) {
	if event.RequestContext.HTTP.Method != cr.Method {
		return nil, false
	}
	params := make(map[string]string, len(cr.Matcher.ParamNames))
	if event.RawPath == cr.Path {
		return params, true
	}
	values := cr.Matcher.Refresh(cr.Path).FindAllStringSubmatchIndex(event.RawPath, -1)
	if values != nil {
		for i, p := range cr.Matcher.ParamNames {
			params[p] = event.RawPath[values[0][i+2]:values[0][i+3]]
		}
	}
	return params, values != nil
}

type Router struct {
	Routes []CachedRoute
}

func NewRouter(services ...Service) *Router {
	var routes []CachedRoute
	for _, service := range services {
		for composite, route := range service.GetRoutes() {
			parts := strings.SplitN(composite, ":", 2)
			cachedRoute := CachedRoute{
				Method: parts[0],
				Path:   parts[1],
				Route:  route,
				Matcher: &CachedMatcher{
					Mutex: &sync.Mutex{},
				},
			}
			routes = append(routes, cachedRoute)
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
	for _, route := range r.Routes {
		if params, ok := route.MatchEvent(event); ok {
			resp, err := route.Route(event, context.WithValue(ctx, "Params", params))
			if err != nil {
				return translateError(err)
			}
			return resp
		}
	}
	return translateError(exceptions.NotFound("route", event.RawPath))
}
