package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const LOCAL_DDB_PORT = 8000

func CreateTable(client *dynamodb.Client) (string, error) {
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

func (l *LocalDynamoServer) CreateLocalClient() (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRetryMaxAttempts(10),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: fmt.Sprintf("http://localhost:%d", l.Port)}, nil
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

type LocalDynamoServer struct {
	Process *os.Process
	Port    int
}

func StartLocalServer(port int, t *testing.T) *LocalDynamoServer {
	workingDir := os.Getenv("PWD")
	cmd := exec.Command(
		"java", fmt.Sprintf("-Djava.library.path=%s/../../dynamodb/DynamoDBLocal_list", workingDir),
		"-jar", fmt.Sprintf("%s/../../dynamodb/DynamoDBLocal.jar", workingDir),
		"-port", strconv.Itoa(port),
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
	return &LocalDynamoServer{Port: port, Process: cmd.Process}
}
