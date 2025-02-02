package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/apitokens"
	"philcali.me/recipes/internal/dynamodb/token"
)

type AuthThunk func(ctx context.Context, apiToken string) (*events.APIGatewayV2CustomAuthorizerSimpleResponse, error)

func JWTAuthThunk(ctx context.Context, apiToken string) (*events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/oauth2/userInfo", os.Getenv("AUTH_POOL_URL")), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Authorization", apiToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid %s with token: %s", req.URL.String(), apiToken)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	var claims map[string]json.RawMessage
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %v", err)
	}
	// We assume that a JWT auth is local user admin
	return &events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context: map[string]interface{}{
			"claims": claims,
			"scopes": []string{
				string(data.RECIPE_WRITE),
				string(data.LIST_WRITE),
				string(data.SUBSCRIPTIONS_WRITE),
				string(data.TOKENS_WRITE),
			},
		},
	}, nil
}

func ApiTokenAuth(ctx context.Context, apiToken string) (*events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	tableName := os.Getenv("TABLE_NAME")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := dynamodb.NewFromConfig(cfg)
	marshaler := token.NewGCM()
	apiTokens := apitokens.NewApiTokenService(tableName, *client, marshaler)
	bearerToken := strings.Split(apiToken, " ")
	if len(bearerToken) != 2 {
		return nil, fmt.Errorf("token provided is invalid: %s", apiToken)
	}
	tokenDTO, err := apiTokens.Get("Global", bearerToken[1])
	if err != nil {
		return nil, err
	}
	return &events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context: map[string]interface{}{
			"claims": map[string]string{
				"username": tokenDTO.AccountId,
			},
			"scopes": tokenDTO.Scopes,
		},
	}, nil
}

func HandleRequest(ctx context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	response := events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: false,
	}
	apiToken, ok := event.Headers["authorization"]
	thunks := []AuthThunk{
		JWTAuthThunk,
		ApiTokenAuth,
	}
	if ok {
		for _, authThunk := range thunks {
			newResp, err := authThunk(ctx, apiToken)
			if newResp != nil {
				return *newResp, err
			}
			if err != nil {
				fmt.Printf("Skipping auth due to %v\n", err)
			}
		}
	}
	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
