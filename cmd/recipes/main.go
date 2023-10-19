package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	recipeData "philcali.me/recipes/internal/dynamodb/recipes"
	shoppingData "philcali.me/recipes/internal/dynamodb/shopping"
	subscriberData "philcali.me/recipes/internal/dynamodb/subscriptions"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/recipes"
	"philcali.me/recipes/internal/routes/shopping"
	"philcali.me/recipes/internal/routes/subscriptions"
	"philcali.me/recipes/internal/sns/services"
)

type App struct {
	Router routes.Router
}

func NewApp() App {
	tableName := os.Getenv("TABLE_NAME")
	topicArn := os.Getenv("TOPIC_ARN")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("Failed to load AWS config.")
	}
	client := dynamodb.NewFromConfig(cfg)
	snsClient := sns.NewFromConfig(cfg)
	marshaler := token.NewGCM()
	router := routes.NewRouter(
		recipes.NewRoute(recipeData.NewRecipeService(tableName, *client, marshaler)),
		shopping.NewRoute(shoppingData.NewShoppingListDynamoDBService(tableName, *client, marshaler)),
		subscriptions.NewRoute(
			subscriberData.NewSubscriptionDynamoDBService(tableName, *client, marshaler),
			&services.NotificationSNSService{
				Sns:      *snsClient,
				TopicArn: topicArn,
			},
		),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to cache all routes: %s", err))
	}
	return App{
		Router: *router,
	}
}

func (app *App) HandleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return app.Router.Invoke(request, ctx), nil
}

func main() {
	app := NewApp()
	lambda.Start(app.HandleRequest)
}
