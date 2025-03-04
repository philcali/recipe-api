package routes_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"philcali.me/recipes/internal/data"
	tokenData "philcali.me/recipes/internal/dynamodb/apitokens"
	auditData "philcali.me/recipes/internal/dynamodb/audits"
	recipeData "philcali.me/recipes/internal/dynamodb/recipes"
	settingsData "philcali.me/recipes/internal/dynamodb/settings"
	shareData "philcali.me/recipes/internal/dynamodb/shares"
	shoppingData "philcali.me/recipes/internal/dynamodb/shopping"
	subscriberData "philcali.me/recipes/internal/dynamodb/subscriptions"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/notifications"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/apitokens"
	"philcali.me/recipes/internal/routes/audits"
	"philcali.me/recipes/internal/routes/recipes"
	"philcali.me/recipes/internal/routes/settings"
	"philcali.me/recipes/internal/routes/shares"
	"philcali.me/recipes/internal/routes/shopping"
	"philcali.me/recipes/internal/routes/subscriptions"
	"philcali.me/recipes/internal/test"
)

func NewLocalServer(t *testing.T) *LocalServer {
	localServer := test.StartLocalServer(test.LOCAL_DDB_PORT+1, t)
	client, err := localServer.CreateLocalClient()
	if err != nil {
		t.Fatalf("Failed to create DDB client: %s", err)
	}
	tableName, err := test.CreateTable(client)
	if err != nil {
		t.Fatalf("Failed to create DDB table: %s", err)
	}
	t.Logf("Successfully created local resources running on %d", test.LOCAL_DDB_PORT)
	marshaler := token.NewGCM()
	router := routes.NewRouter(
		recipes.NewRoute(recipeData.NewRecipeService(tableName, *client, marshaler)),
		shopping.NewRoute(shoppingData.NewShoppingListService(tableName, *client, marshaler)),
		apitokens.NewRouteWithIndex(tokenData.NewApiTokenService(tableName, *client, marshaler), "GS1"),
		settings.NewRoute(settingsData.NewSettingService(tableName, *client, marshaler)),
		audits.NewRouteWithIndex(auditData.NewAuditService(tableName, *client, marshaler), "GS1"),
		shares.NewRouteWithIndex(shareData.NewShareService(tableName, *client, marshaler), "GS1"),
		subscriptions.NewRoute(
			subscriberData.NewSubscriptionService(tableName, *client, marshaler),
			&LocalNotifications{
				Cache: make(map[string]notifications.SubscribeInput),
			},
		),
	)
	if err != nil {
		t.Fatalf("Failed to create a router: %s", err)
	}
	return &LocalServer{
		Router:         router,
		TableName:      tableName,
		DynamoDB:       client,
		TokenMarshaler: marshaler,
		Username:       "nobody",
		Email:          "nobody@email.com",
	}
}

type LocalNotifications struct {
	Cache map[string]notifications.SubscribeInput
}

func (ln *LocalNotifications) Subscribe(input notifications.SubscribeInput) (*notifications.SubscribeOutput, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	ln.Cache[id.String()] = input
	return &notifications.SubscribeOutput{
		SubscriberId: id.String(),
	}, nil
}

func (ln *LocalNotifications) Unsubscribe(subscriberId string) error {
	delete(ln.Cache, subscriberId)
	return nil
}

type LocalServer struct {
	Router         *routes.Router
	DynamoDB       *dynamodb.Client
	TokenMarshaler *token.EncryptionTokenMarshaler
	TableName      string
	Username       string
	Email          string
}

func (ls *LocalServer) UpdateIdentity(username, email string) {
	ls.Username = username
	ls.Email = email
}

func (ls *LocalServer) Request(t *testing.T, method string, path string, body []byte, out any, params map[string]string) events.APIGatewayV2HTTPResponse {
	request := events.APIGatewayV2HTTPRequest{}
	fd, err := os.ReadFile(filepath.Join("router_test", "template.json"))
	if err != nil {
		t.Fatalf("Failed to load request template: %s", err)
	}
	if err := json.Unmarshal(fd, &request); err != nil {
		t.Fatalf("Failed to deserialize request template: %s", err)
	}
	request.RawPath = path
	request.QueryStringParameters = params
	request.RequestContext.HTTP.Method = method
	request.RequestContext.HTTP.Path = path
	request.RequestContext.Authorizer.Lambda["jwt"] = map[string]interface{}{
		"username": string(ls.Username),
		"email":    string(ls.Email),
	}
	request.Body = string(body)
	response := ls.Router.Invoke(request, context.TODO())
	if out != nil {
		if err := json.Unmarshal([]byte(response.Body), &out); err != nil {
			t.Fatalf("Failed to deserialize payload for %s %s: %s", method, path, response.Body)
		}
	}
	return response
}

func (ls *LocalServer) Options(t *testing.T, path string) events.APIGatewayV2HTTPResponse {
	return ls.Request(t, "OPTIONS", path, nil, nil, nil)
}

func (ls *LocalServer) Get(t *testing.T, out any, path string) events.APIGatewayV2HTTPResponse {
	return ls.Request(t, "GET", path, nil, &out, nil)
}

func (ls *LocalServer) GetQuery(t *testing.T, out any, path string, params map[string]string) events.APIGatewayV2HTTPResponse {
	return ls.Request(t, "GET", path, nil, &out, params)
}

func (ls *LocalServer) Post(t *testing.T, out any, path string, body any) events.APIGatewayV2HTTPResponse {
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to serialize input: %s", err)
	}
	return ls.Request(t, "POST", path, payload, &out, nil)
}

func (ls *LocalServer) Delete(t *testing.T, path string) events.APIGatewayV2HTTPResponse {
	return ls.Request(t, "DELETE", path, nil, nil, nil)
}

func (ls *LocalServer) Put(t *testing.T, out any, path string, body any) events.APIGatewayV2HTTPResponse {
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to serialize input: %s", err)
	}
	return ls.Request(t, "PUT", path, payload, &out, nil)
}

func TestRouter(t *testing.T) {
	server := NewLocalServer(t)
	t.Run("RecipeWorkflow", func(t *testing.T) {
		var createdRecipe recipes.Recipe
		created := server.Post(t, &createdRecipe, "/recipes", &recipes.RecipeInput{
			Name:             aws.String("Fart Soup"),
			Instructions:     aws.String("Eat a bowl of beans. Wait for 30 minutes. Fart in mason jar."),
			NumberOfServings: aws.Int(2),
			Type:             aws.String("Food"),
			Ingredients: &[]recipes.Ingredient{
				{
					Name:        "beans",
					Measurement: "can",
					Amount:      1.5,
				},
			},
		})
		if created.StatusCode != 200 {
			t.Fatalf("Response on create %d: %s", created.StatusCode, created.Body)
		}
		if *createdRecipe.NumberOfServings != 2 {
			t.Fatalf("Failed to set number of servings, expected 2, got %s", created.Body)
		}
		get := server.Get(t, nil, fmt.Sprintf("/recipes/%s", createdRecipe.Id))
		if get.StatusCode != 200 {
			t.Fatalf("Response failed with status %d: %s", get.StatusCode, get.Body)
		}
		if created.Body != get.Body {
			t.Fatalf("Get response does not match create: %s != %s", get.Body, created.Body)
		}
		var results data.QueryResults[recipes.Recipe]
		list := server.Get(t, &results, "/recipes")
		if len(results.Items) < 1 || createdRecipe.Id != results.Items[0].Id || createdRecipe.Ingredients[0].Amount != 1.5 {
			t.Fatalf("List does not contain %s: %s", created.Body, list.Body)
		}
		updated := server.Put(t, nil, fmt.Sprintf("/recipes/%s", createdRecipe.Id), &recipes.RecipeInput{
			Name:               aws.String("Fart Update"),
			PrepareTimeMinutes: aws.Int(35),
			Thumbnail:          aws.String("this would normally be base64 encoded"),
			Type:               aws.String("Drink"),
			Nutrients: &[]recipes.Nutrient{
				{
					Name:   "carbohydrates",
					Unit:   "gram",
					Amount: 26,
				},
				{
					Name:   "protein",
					Unit:   "gram",
					Amount: 6,
				},
			},
		})
		if updated.StatusCode != 200 {
			t.Fatalf("Update response %d: %s", updated.StatusCode, updated.Body)
		}
		var getUpdateRecipe recipes.Recipe
		getUpdate := server.Get(t, &getUpdateRecipe, fmt.Sprintf("/recipes/%s", createdRecipe.Id))
		if getUpdateRecipe.Name != "Fart Update" {
			t.Fatalf("Failed to update %s: %s", getUpdateRecipe.Name, getUpdate.Body)
		}
		if *getUpdateRecipe.PrepareTimeMinutes < 35 {
			t.Fatalf("Failed to update %d: %s", getUpdateRecipe.PrepareTimeMinutes, getUpdate.Body)
		}
		if len(getUpdateRecipe.Nutrients) < 2 {
			t.Fatalf("Failed to update %v", getUpdateRecipe.Nutrients)
		}
		if *getUpdateRecipe.Thumbnail != "this would normally be base64 encoded" {
			t.Fatalf("Failed to update thumbnail %s, %s", getUpdateRecipe.Name, getUpdate.Body)
		}
		if *getUpdateRecipe.Type != "Drink" {
			t.Fatalf("Failed to update type %s, %s", getUpdateRecipe.Name, getUpdate.Body)
		}
		deleted := server.Delete(t, fmt.Sprintf("/recipes/%s", createdRecipe.Id))
		if deleted.StatusCode != 204 {
			t.Fatalf("Response on delete %d: %s", deleted.StatusCode, deleted.Body)
		}
	})

	t.Run("ShoppingListWorkflow", func(t *testing.T) {
		var createdList shopping.ShoppingList
		created := server.Post(t, &createdList, "/lists", &shopping.ShoppingListInput{
			Name: aws.String("My List"),
			Items: &[]shopping.ShoppingListItem{
				{
					Name:        "bread",
					Measurement: "loaf",
					Amount:      1,
					Completed:   false,
				},
				{
					Name:        "milk",
					Measurement: "gallon",
					Amount:      1,
					Completed:   false,
				},
			},
		})
		if created.StatusCode != 200 {
			t.Fatalf("Failed to create shopping list, expected 200 got %d: %s", created.StatusCode, created.Body)
		}
		get := server.Get(t, nil, fmt.Sprintf("/lists/%s", createdList.Id))
		if get.StatusCode != 200 {
			t.Fatalf("Failed to get new list %s, expected 200 got %d: %s", createdList.Id, get.StatusCode, get.Body)
		}
		if get.Body != created.Body {
			t.Fatalf("Expected body, expected %s got %s", created.Body, get.Body)
		}
		var results data.QueryResults[shopping.ShoppingList]
		list := server.Get(t, &results, "/lists")
		if len(results.Items) < 1 {
			t.Fatalf("Failed to query for lists, expected 1 got %v", list)
		}
		hourFromNow := time.Now().Add(time.Hour + 1)
		var updatedList shopping.ShoppingList
		update := server.Put(t, &updatedList, fmt.Sprintf("/lists/%s", createdList.Id), &shopping.ShoppingListInput{
			Name:      aws.String("New Name"),
			ExpiresIn: &hourFromNow,
			Items: &[]shopping.ShoppingListItem{
				{
					Name:        "bread",
					Measurement: "loaf",
					Amount:      1,
					Completed:   true,
				},
				{
					Name:        "milk",
					Measurement: "gallon",
					Amount:      1,
					Completed:   true,
				},
				{
					Name:        "eggs",
					Measurement: "whole",
					Amount:      12,
					Completed:   false,
				},
			},
		})
		if update.StatusCode != 200 {
			t.Fatalf("Failed to update, expected 200, got %d: %s", update.StatusCode, update.Body)
		}
		if updatedList.ExpiresIn == nil || updatedList.ExpiresIn.Before(time.Now().Add(time.Minute+55)) {
			t.Fatalf("Failed to update the shopping list %s: %s", updatedList.ExpiresIn, update.Body)
		}
		if len(updatedList.Items) < 3 {
			t.Fatalf("Failed to update the shopping list %v: %s", updatedList.Items, update.Body)
		}
		delete := server.Delete(t, fmt.Sprintf("/lists/%s", createdList.Id))
		if delete.StatusCode != 204 {
			t.Fatalf("Failed to delete, expected 204, got %d: %s", delete.StatusCode, delete.Body)
		}
		getRemoved := server.Get(t, nil, fmt.Sprintf("/lists/%s", createdList.Id))
		if getRemoved.StatusCode != 404 {
			t.Fatalf("Failed to actually delete, expected 404, got %d: %s", getRemoved.StatusCode, getRemoved.Body)
		}
	})

	t.Run("ApiToken", func(t *testing.T) {
		var createdToken apitokens.ApiToken
		created := server.Post(t, &createdToken, "/tokens", &apitokens.ApiTokenInput{
			Name: aws.String("Test Token"),
			Scopes: []data.Scope{
				data.LIST_READ,
				data.RECIPE_READ,
			},
		})
		if created.StatusCode != 200 {
			t.Fatalf("Expected 200 on create, but received: %d", created.StatusCode)
		}
		get := server.Get(t, nil, fmt.Sprintf("/tokens/%s", createdToken.Value))
		if get.StatusCode != 200 {
			t.Fatalf("Expected 200 on get, but received: %d", get.StatusCode)
		}
		var results data.QueryResults[apitokens.ApiToken]
		listTokens := server.Get(t, &results, "/tokens")
		if listTokens.StatusCode != 200 {
			t.Fatalf("Expected 200 on list, but received: %d, %v", listTokens.StatusCode, listTokens.Body)
		}
		if len(results.Items) != 1 {
			t.Fatalf("Expected list of tokens to be 1, received: %d", len(results.Items))
		}
		apiToken := results.Items[0]
		if apiToken.Value != createdToken.Value {
			t.Fatalf("Expected list item is not expected: %v", apiToken)
		}
		del := server.Delete(t, fmt.Sprintf("/tokens/%s", apiToken.Value))
		if del.StatusCode != 204 {
			t.Fatalf("Expected 204 on delete, received: %d", del.StatusCode)
		}
	})

	t.Run("SettingsWorkflow", func(t *testing.T) {
		var defaultSettings settings.Settings
		getSettings := server.Get(t, &defaultSettings, "/settings")
		if getSettings.StatusCode != 200 {
			t.Fatalf("Failed to get default settings: %d", getSettings.StatusCode)
		}

		if defaultSettings.AutoShareLists {
			t.Fatalf("Expected the auto share lists to be false but was %v", defaultSettings.AutoShareLists)
		}

		if defaultSettings.AutoShareRecipes {
			t.Fatalf("Expected the auto share recipes to be false but was %v", defaultSettings.AutoShareRecipes)
		}

		var updatedSettings settings.Settings
		update := server.Post(t, &updatedSettings, "/settings", settings.SettingsInput{
			AutoShareLists:   aws.Bool(true),
			AutoShareRecipes: aws.Bool(true),
		})
		if update.StatusCode != 200 {
			t.Fatalf("Expected to update settings: %d", update.StatusCode)
		}

		if !updatedSettings.AutoShareLists {
			t.Fatalf("Expected the auto share lists to be true but was %v", updatedSettings.AutoShareLists)
		}

		if !updatedSettings.AutoShareRecipes {
			t.Fatalf("Expected the auto share recipes to be true but was %v", updatedSettings.AutoShareRecipes)
		}
	})

	t.Run("AuditWorkflow", func(t *testing.T) {
		db := auditData.NewAuditService(server.TableName, *server.DynamoDB, server.TokenMarshaler)
		created, err := db.Create("nobody", data.AuditInputDTO{
			ResourceId:   aws.String("resourceId"),
			ResourceType: aws.String("Recipe"),
			Action:       aws.String("CREATED"),
			AccountId:    aws.String("nobody"),
		})
		if err != nil {
			t.Fatal("Failed to create test audit entry")
		}
		var listAudits data.QueryResults[audits.Audit]
		getAudits := server.Get(t, &listAudits, "/audits")
		if getAudits.StatusCode != 200 {
			t.Fatalf("Expected list audits to return, got %d", getAudits.StatusCode)
		}
		if len(listAudits.Items) < 1 {
			t.Fatalf("Expected there to be at least 1 entry, got %d", len(listAudits.Items))
		}
		if created.ResourceId != listAudits.Items[0].ResourceId {
			t.Fatalf("Expected %v, got %v", created, listAudits.Items[0])
		}
		delResp := server.Delete(t, "/audits/"+created.SK)
		if delResp.StatusCode != 204 {
			t.Fatalf("Expected no content on delete: %d", delResp.StatusCode)
		}
	})

	t.Run("ShareWorkflow", func(t *testing.T) {
		var shareRequest shares.ShareRequest
		status := data.REQUESTED
		created := server.Post(t, &shareRequest, "/shares", shares.ShareRequestInput{
			Approver:       aws.String("nobody2@email.com"),
			ApprovalStatus: &status,
		})
		if created.StatusCode != 200 {
			t.Fatalf("Failed to create share request %d", created.StatusCode)
		}
		var listResults data.QueryResults[shares.ShareRequest]
		selfList := server.Get(t, &listResults, "/shares")
		if selfList.StatusCode != 200 {
			t.Fatalf("Failed to list %d", selfList.StatusCode)
		}
		if len(listResults.Items) < 1 {
			t.Fatalf("Expected at least 1 request, got %d", len(listResults.Items))
		}
		if listResults.Items[0].ApprovalStatus != shareRequest.ApprovalStatus {
			t.Fatalf("Expected %s, got %s", data.REQUESTED, listResults.Items[0].ApprovalStatus)
		}
		var getShare shares.ShareRequest
		g := server.Get(t, &getShare, "/shares/"+shareRequest.Id)
		if g.StatusCode != 200 {
			t.Fatalf("Failed to get share request %d", g.StatusCode)
		}
		server.UpdateIdentity("nobody2", "nobody2@email.com")
		requestList := server.GetQuery(t, &listResults, "/shares", map[string]string{
			"status": "REQUESTED",
		})
		if requestList.StatusCode != 200 {
			t.Fatalf("Expected 200 but got %d", requestList.StatusCode)
		}
		if len(listResults.Items) < 1 {
			t.Fatalf("Expected at least 1, but got %d", len(listResults.Items))
		}
		var updated shares.ShareRequest
		status = data.APPROVED
		updateRes := server.Put(t, &updated, "/shares/"+shareRequest.Id, shares.ShareRequestInput{
			ApprovalStatus: &status,
		})
		if updateRes.StatusCode != 200 {
			t.Fatalf("Expected 200 but got %d", updateRes.StatusCode)
		}
	})

	t.Run("SubscriptionWorkflow", func(t *testing.T) {
		var createdSubscriber subscriptions.Subscription
		created := server.Post(t, &createdSubscriber, "/subscriptions", &subscriptions.SubscriptionInput{
			Endpoint: aws.String("philcali@example.com"),
			Protocol: aws.String("email"),
		})

		if created.StatusCode != 200 {
			t.Fatalf("Failed to create a subscription: %s", created.Body)
		}

		if createdSubscriber.Endpoint != "philcali@example.com" {
			t.Fatalf("Failed to respond with subscription: %v", createdSubscriber)
		}

		var listSubscribers data.QueryResults[subscriptions.Subscription]
		listResp := server.Get(t, listSubscribers, "/subscriptions")
		if listResp.StatusCode != 200 {
			t.Fatalf("Failed to list subscriptions: %v", listResp.Body)
		}

		if len(listSubscribers.Items) == 1 {
			t.Fatalf("Failed to list the appropriate amount: %v", listSubscribers.Items)
		}

		if listSubscribers.NextToken != nil {
			t.Fatal("Expected nextToken to be set but was nil")
		}

		var getSubscriber subscriptions.Subscription
		getResp := server.Get(t, &getSubscriber, "/subscriptions/"+createdSubscriber.Id)

		if getResp.StatusCode != 200 {
			t.Fatalf("Failed to get subscriber: %s", getResp.Body)
		}

		if getSubscriber.Id != createdSubscriber.Id {
			t.Fatalf("Failed to get the id: %s", createdSubscriber.Id)
		}

		deleteResp := server.Delete(t, "/subscriptions/"+createdSubscriber.Id)

		if deleteResp.StatusCode != 204 {
			t.Fatalf("Failed to delete the subscriber: %s, %v", createdSubscriber.Id, deleteResp.Body)
		}
	})

	t.Run("UpdateFailure", func(t *testing.T) {
		updated := server.Post(t, nil, "/recipes/not-existent", &recipes.RecipeInput{
			Name: aws.String("Non-Existence"),
		})
		if updated.StatusCode != 404 {
			t.Fatalf("Expected status code of 404, but got %d: %s", updated.StatusCode, updated.Body)
		}
	})

	t.Run("CorsPreflight", func(t *testing.T) {
		preflight := server.Options(t, "/recipes")
		if preflight.StatusCode != 200 {
			t.Fatalf("Received a %d status code, expected 200", preflight.StatusCode)
		}
		if preflight.Body != "" {
			t.Fatalf("Received a response body for OPTIONS: %s", preflight.Body)
		}
		expected := map[string]string{
			"content-length":               "0",
			"access-control-allow-headers": "Content-Type, Content-Length, Authorization",
			"access-control-allow-methods": "GET, PUT, POST, DELETE",
			"access-control-allow-origin":  "*",
		}
		if !maps.Equal(preflight.Headers, expected) {
			t.Fatalf("Headers from preflight %v, do not match expected %v", preflight.Headers, expected)
		}
	})
}
