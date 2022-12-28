package routes_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/recipes"
)

const LOCAL_DDB_PORT = 8000

func _createTable(client *dynamodb.Client) (string, error) {
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: aws.String("PK"),
			KeyType:       types.KeyTypeHash,
		},
		{
			AttributeName: aws.String("SK"),
			KeyType:       types.KeyTypeRange,
		},
	}
	atrributes := []types.AttributeDefinition{
		{
			AttributeName: aws.String("PK"),
			AttributeType: types.ScalarAttributeTypeS,
		},
		{
			AttributeName: aws.String("SK"),
			AttributeType: types.ScalarAttributeTypeS,
		},
	}
	output, err := client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName:            aws.String("RecipeData"),
		KeySchema:            keySchema,
		BillingMode:          types.BillingModePayPerRequest,
		AttributeDefinitions: atrributes,
	})
	if err != nil {
		return "", err
	}
	waiter := dynamodb.NewTableExistsWaiter(client, func(tewo *dynamodb.TableExistsWaiterOptions) {
		tewo.LogWaitAttempts = true
	})
	_, err = waiter.WaitForOutput(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: output.TableDescription.TableName,
	}, time.Second*5)
	return *output.TableDescription.TableName, err
}

func _createLocalClient() (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRetryMaxAttempts(10),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: fmt.Sprintf("http://localhost:%d", LOCAL_DDB_PORT)}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     "fake",
				SecretAccessKey: "fake",
				SessionToken:    "fake",
			}}),
	)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

func _startLocalServer(t *testing.T) {
	workingDir := os.Getenv("PWD")
	cmd := exec.Command(
		"java", fmt.Sprintf("-Djava.library.path=%s/../../dynamodb/DynamoDBLocal_list", workingDir),
		"-jar", fmt.Sprintf("%s/../../dynamodb/DynamoDBLocal.jar", workingDir),
		"-port", strconv.Itoa(LOCAL_DDB_PORT),
		"-inMemory",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start local DDB server: %s", err)
	}
	t.Cleanup(func() {
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate local DDB server: %s", err)
		}
	})
}

func NewLocalServer(t *testing.T) *LocalServer {
	client, err := _createLocalClient()
	if err != nil {
		t.Fatalf("Failed to create DDB client: %s", err)
	}
	tableName, err := _createTable(client)
	if err != nil {
		t.Fatalf("Failed to create DDB table: %s", err)
	}
	t.Logf("Successfully created local resources running on %d", LOCAL_DDB_PORT)
	router, err := routes.NewRouter(
		recipes.NewRoute(
			services.NewRecipeService(tableName, *client, token.NewGCM()),
		),
	)
	if err != nil {
		t.Fatalf("Failed to create a router: %s", err)
	}
	return &LocalServer{
		Router: router,
	}
}

type LocalServer struct {
	Router *routes.Router
}

func (ls *LocalServer) Request(t *testing.T, method string, path string, body []byte, out any) events.APIGatewayV2HTTPResponse {
	request := events.APIGatewayV2HTTPRequest{}
	fd, err := os.ReadFile(filepath.Join("router_test", "template.json"))
	if err != nil {
		t.Fatalf("Failed to load request template: %s", err)
	}
	if err := json.Unmarshal(fd, &request); err != nil {
		t.Fatalf("Failed to deserialize request template: %s", err)
	}
	request.RawPath = path
	request.RequestContext.HTTP.Method = method
	request.RequestContext.HTTP.Path = path
	request.Body = string(body)
	response := ls.Router.Invoke(request, context.TODO())
	if out != nil {
		if err := json.Unmarshal([]byte(response.Body), &out); err != nil {
			t.Fatalf("Failed to deserialize payload for %s %s: %s", method, path, response.Body)
		}
	}
	return response
}

func (ls *LocalServer) Get(t *testing.T, out any, path string) events.APIGatewayV2HTTPResponse {
	return ls.Request(t, "GET", path, nil, &out)
}

func (ls *LocalServer) Post(t *testing.T, out any, path string, body any) events.APIGatewayV2HTTPResponse {
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to serialize input: %s", err)
	}
	return ls.Request(t, "POST", path, payload, &out)
}

func (ls *LocalServer) Delete(t *testing.T, path string) events.APIGatewayV2HTTPResponse {
	return ls.Request(t, "DELETE", path, nil, nil)
}

func (ls *LocalServer) Put(t *testing.T, out any, path string, body any) events.APIGatewayV2HTTPResponse {
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to serialize input: %s", err)
	}
	return ls.Request(t, "PUT", path, payload, &out)
}

func TestRouter(t *testing.T) {
	_startLocalServer(t)
	server := NewLocalServer(t)
	t.Run("RecipeWorkflow", func(t *testing.T) {
		var createdRecipe recipes.Recipe
		created := server.Post(t, &createdRecipe, "/recipes", &recipes.RecipeInput{
			Name:               aws.String("Fart Soup"),
			Instructions:       aws.String("Eat a bowl of beans. Wait for 30 minutes. Fart in mason jar."),
			PrepareTimeMinutes: aws.Int(30),
			Ingredients: &[]recipes.Ingredient{
				{
					Name:        "beans",
					Measurement: "can",
					Amount:      1.5,
				},
			},
		})
		if 200 != created.StatusCode {
			t.Fatalf("Response on create %d: %s", created.StatusCode, created.Body)
		}
		get := server.Get(t, nil, fmt.Sprintf("/recipes/%s", createdRecipe.Id))
		if 200 != get.StatusCode {
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
		})
		if 200 != updated.StatusCode {
			t.Fatalf("Update response %d: %s", updated.StatusCode, updated.Body)
		}
		var getUpdateRecipe recipes.Recipe
		getUpdate := server.Get(t, &getUpdateRecipe, fmt.Sprintf("/recipes/%s", createdRecipe.Id))
		if getUpdateRecipe.Name != "Fart Update" {
			t.Fatalf("Failed to update %s: %s", getUpdateRecipe.Name, getUpdate.Body)
		}
		if getUpdateRecipe.PrepareTimeMinutes < 35 {
			t.Fatalf("Failed to update %d: %s", getUpdateRecipe.PrepareTimeMinutes, getUpdate.Body)
		}
		deleted := server.Delete(t, fmt.Sprintf("/recipes/%s", createdRecipe.Id))
		if 204 != deleted.StatusCode {
			t.Fatalf("Response on delete %d: %s", deleted.StatusCode, deleted.Body)
		}
	})

	t.Run("UpdateFailure", func(t *testing.T) {
		updated := server.Post(t, nil, "/recipe/not-existent", &recipes.RecipeInput{
			Name: aws.String("Non-Existence"),
		})
		if 404 != updated.StatusCode {
			t.Fatalf("Expected status code of 404, but got %d: %s", updated.StatusCode, updated.Body)
		}
	})
}
