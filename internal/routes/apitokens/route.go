package apitokens

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type ApiTokenService struct {
	data      data.ApiTokenDataService
	indexName string
}

func NewRouteWithIndex(data data.ApiTokenDataService, indexName string) routes.Service {
	return &ApiTokenService{
		data:      data,
		indexName: indexName,
	}
}

func NewRoute(data data.ApiTokenDataService) routes.Service {
	return NewRouteWithIndex(data, os.Getenv("INDEX_NAME_1"))
}

func _convertToken(tokenDTO data.ApiTokenDTO) ApiToken {
	var expiresIn *time.Time = nil
	if tokenDTO.ExpiresIn != nil {
		expiresIn = aws.Time(time.UnixMilli(int64(*tokenDTO.ExpiresIn)))
	}
	return ApiToken{
		Name:      tokenDTO.Name,
		Value:     tokenDTO.SK,
		Scopes:    tokenDTO.Scopes,
		ExpiresIn: expiresIn,
	}
}

func _generateTokenHash(length int) (string, error) {
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

func (as *ApiTokenService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/tokens":             util.AuthorizedRoute(as.ListTokens),
		"GET:/tokens/:tokenId":    util.AuthorizedRoute(as.GetToken),
		"POST:/tokens":            util.AuthorizedRoute(as.CreateToken),
		"PUT:/tokens/:tokenId":    util.AuthorizedRoute(as.UpdateToken),
		"DELETE:/tokens/:tokenId": util.AuthorizedRoute(as.DeleteToken),
	}
}

func (as *ApiTokenService) ListTokens(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return util.SerializeListByIndex(as.data, _convertToken, as.indexName, event, ctx)
}

func (as *ApiTokenService) GetToken(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	item, err := as.data.Get("Global", util.RequestParam(ctx, "tokenId"))
	return util.SerializeResponseOK(_convertToken, item, err)
}

func (as *ApiTokenService) CreateToken(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := ApiTokenInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	var expiresIn *int
	if input.ExpiresIn != nil {
		expiresIn = aws.Int(int(input.ExpiresIn.UnixMilli()))
	}
	tokenHash, err := _generateTokenHash(32)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InternalServer(err.Error())
	}
	created, err := as.data.CreateWithItemId("Global", data.ApiTokenInputDTO{
		Name:      input.Name,
		Scopes:    &input.Scopes,
		AccountId: aws.String(util.Username(ctx)),
		ExpiresIn: expiresIn,
	}, tokenHash)
	return util.SerializeResponseOK(_convertToken, created, err)
}

func (as *ApiTokenService) UpdateToken(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := ApiTokenInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	var expiresIn *int
	if input.ExpiresIn != nil {
		expiresIn = aws.Int(int(input.ExpiresIn.UnixMilli()))
	}
	item, err := as.data.Update("Global", util.RequestParam(ctx, "tokenId"), data.ApiTokenInputDTO{
		Name:      input.Name,
		ExpiresIn: expiresIn,
	})
	return util.SerializeResponseOK(_convertToken, item, err)
}

func (as *ApiTokenService) DeleteToken(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	err := as.data.Delete("Global", util.RequestParam(ctx, "tokenId"))
	return util.SerializeResponseNoContent(err)
}
