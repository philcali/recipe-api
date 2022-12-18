package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/recipes"
)

type App struct {
	Router routes.Router
}

func NewApp() App {
	tableName := os.Getenv("TABLE_NAME")
	client := dynamodb.NewFromConfig(*aws.NewConfig())
	return App{
		Router: *routes.NewRouter(
			recipes.NewRoute(services.NewRecipeService(tableName, *client)),
		),
	}
}

func (app *App) HandleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return app.Router.Invoke(request, ctx), nil
}

func main() {
	app := NewApp()
	lambda.Start(app.HandleRequest)
}
