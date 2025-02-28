package events

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/settings"
	"philcali.me/recipes/internal/dynamodb/shares"
	"philcali.me/recipes/internal/dynamodb/shopping"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/test"
)

func TestCopyResources(t *testing.T) {
	localServer := test.StartLocalServer(test.LOCAL_DDB_PORT+2, t)
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

	t.Run("CopyHandler", func(t *testing.T) {
		settingData := settings.NewSettingService(tableName, *client, marshaler)
		sharingData := shares.NewShareService(tableName, *client, marshaler)

		copyHandler := &CopySharingResourceHandler{
			Setting:   settingData,
			Sharing:   sharingData,
			DynamoDB:  client,
			TableName: tableName,
		}

		accountId := uuid.NewString()
		_, err := settingData.CreateWithItemId(accountId, data.SettingsInputDTO{
			AutoShareLists:   aws.Bool(true),
			AutoShareRecipes: aws.Bool(true),
		}, "Global")

		if err != nil {
			t.Fatalf("Expected a set, but got %v", err)
		}

		remove := events.DynamoDBEventRecord{
			EventName: "REMOVE",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute(fmt.Sprintf("%s:Settings", accountId)),
				},
			},
		}

		if copyHandler.Filter(remove) {
			t.Fatalf("Expected the record to not be filtered %v", remove)
		}

		millis := time.Now()
		itemId := uuid.NewString()
		content, err := millis.MarshalText()
		insert := events.DynamoDBEventRecord{
			EventName: "INSERT",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute(fmt.Sprintf("%s:ShoppingList", accountId)),
					"SK": events.NewStringAttribute(itemId),
				},
				NewImage: map[string]events.DynamoDBAttributeValue{
					"PK":    events.NewStringAttribute(fmt.Sprintf("%s:ShoppingList", accountId)),
					"SK":    events.NewStringAttribute(itemId),
					"name":  events.NewStringAttribute("Giant"),
					"owner": events.NewStringAttribute("nobody@email.com"),
					"items": events.NewListAttribute([]events.DynamoDBAttributeValue{
						events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
							"name": events.NewStringAttribute("Milk"),
						}),
					}),
					"createTime": events.NewStringAttribute(string(content)),
					"updateTime": events.NewStringAttribute(string(content)),
				},
			},
		}

		if !copyHandler.Filter(insert) {
			t.Fatalf("Expected the record to filtered %v", insert)
		}

		if err := copyHandler.Apply(insert); err != nil {
			t.Fatalf("Failed to apply empty record %v", err)
		}

		status := data.APPROVED
		_, err = sharingData.Create(accountId, data.ShareRequestInputDTO{
			RequesterId:    aws.String(accountId),
			Approver:       aws.String(itemId),
			ApproverId:     aws.String("nobody2"),
			ApprovalStatus: &status,
			Requester:      aws.String("nobody@email.com"),
		})
		if err != nil {
			t.Fatalf("Failed to create share for %s", accountId)
		}

		if err := copyHandler.Apply(insert); err != nil {
			t.Fatalf("Failed to apply sharing record %v", err)
		}

		lists := shopping.NewShoppingListService(tableName, *client, marshaler)
		copied, err := lists.Get("nobody2", itemId)
		if err != nil {
			t.Fatalf("Failed to get copied list: %v", err)
		}

		if copied.Name != "Giant" {
			t.Fatal("Failed")
		}
	})
}
