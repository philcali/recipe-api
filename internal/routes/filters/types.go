package filters

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type FilterContext struct {
	Request  *events.APIGatewayV2HTTPRequest
	Response *events.APIGatewayV2HTTPResponse
	Context  *context.Context
}

type RequestFilter interface {
	Filter(ctx *FilterContext) (*FilterContext, bool)
}

type CorsFilter struct {
	Methods []string
	Origins []string
	Headers []string
}

func (cf *CorsFilter) Filter(ctx *FilterContext) (*FilterContext, bool) {
	if ctx.Request.RequestContext.HTTP.Method == "OPTIONS" {
		headers := ctx.Response.Headers
		if headers == nil {
			headers = make(map[string]string, 4)
		}
		headers["content-length"] = "0"
		headers["access-control-allow-headers"] = strings.Join(cf.Headers, ", ")
		headers["access-control-allow-methods"] = strings.Join(cf.Methods, ", ")
		headers["access-control-allow-origin"] = strings.Join(cf.Origins, ", ")
		return &FilterContext{
			Request: ctx.Request,
			Context: ctx.Context,
			Response: &events.APIGatewayV2HTTPResponse{
				Headers:    headers,
				StatusCode: ctx.Response.StatusCode,
			},
		}, true
	}
	return ctx, false
}

func DefaultFilterContext(event events.APIGatewayV2HTTPRequest, ctx context.Context) *FilterContext {
	return &FilterContext{
		Request: &event,
		Response: &events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
		},
		Context: &ctx,
	}
}

func DefaultCorsFilter() *CorsFilter {
	methods := [4]string{"GET", "PUT", "POST", "DELETE"}
	headers := [3]string{"Content-Type", "Content-Length", "Authorization"}
	origins := [1]string{"*"}
	return &CorsFilter{
		Methods: methods[:],
		Headers: headers[:],
		Origins: origins[:],
	}
}
