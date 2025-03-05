package events

import (
	"fmt"
	"strings"
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

	t.Run("UpdateHandler", func(t *testing.T) {
		sharingData := shares.NewShareService(tableName, *client, marshaler)

		updateHandler := &UpdateSharedResourceHandler{
			Sharing:   sharingData,
			DynamoDB:  client,
			TableName: tableName,
		}

		accountId := uuid.NewString()
		millis := time.Now()
		itemId := uuid.NewString()
		content, _ := millis.MarshalText()
		modify := events.DynamoDBEventRecord{
			EventName: "MODIFY",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute(fmt.Sprintf("%s:ShoppingList", accountId)),
					"SK": events.NewStringAttribute(itemId),
				},
				OldImage: map[string]events.DynamoDBAttributeValue{
					"PK":    events.NewStringAttribute(fmt.Sprintf("%s:ShoppingList", accountId)),
					"SK":    events.NewStringAttribute(itemId),
					"name":  events.NewStringAttribute("Giant"),
					"owner": events.NewStringAttribute("nobody@email.com"),
					"items": events.NewListAttribute([]events.DynamoDBAttributeValue{
						events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
							"name": events.NewStringAttribute("Milk"),
						}),
					}),
					"updateToken": events.NewStringAttribute("abc-123"),
					"createTime":  events.NewStringAttribute(string(content)),
					"updateTime":  events.NewStringAttribute(string(content)),
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
					"updateToken": events.NewStringAttribute("abc-123"),
					"createTime":  events.NewStringAttribute(string(content)),
					"updateTime":  events.NewStringAttribute(string(content)),
				},
			},
		}

		if updateHandler.Filter(modify) {
			t.Fatalf("Expected the record to skip: %v", modify)
		}

		modify = events.DynamoDBEventRecord{
			EventName: "MODIFY",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute(fmt.Sprintf("%s:ShoppingList", accountId)),
					"SK": events.NewStringAttribute(itemId),
				},
				OldImage: map[string]events.DynamoDBAttributeValue{
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
				NewImage: map[string]events.DynamoDBAttributeValue{
					"PK":    events.NewStringAttribute(fmt.Sprintf("%s:ShoppingList", accountId)),
					"SK":    events.NewStringAttribute(itemId),
					"name":  events.NewStringAttribute("Giant"),
					"owner": events.NewStringAttribute("nobody@email.com"),
					"items": events.NewListAttribute([]events.DynamoDBAttributeValue{
						events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
							"name":      events.NewStringAttribute("Milk"),
							"completed": events.NewBooleanAttribute(true),
						}),
					}),
					"updateToken": events.NewStringAttribute("abc-123"),
					"createTime":  events.NewStringAttribute(string(content)),
					"updateTime":  events.NewStringAttribute(string(content)),
				},
			},
		}

		if !updateHandler.Filter(modify) {
			t.Fatalf("Expected the update handler to filter record %v", modify)
		}

		approvalStatus := data.APPROVED
		approvedShare, err := sharingData.Create(accountId, data.ShareRequestInputDTO{
			Requester:      aws.String("philip.cali@gmail.com"),
			RequesterId:    aws.String(accountId),
			Approver:       aws.String("philip.cali@example.com"),
			ApproverId:     aws.String(uuid.NewString()),
			ApprovalStatus: &approvalStatus,
		})
		if err != nil {
			t.Fatalf("Expected to create a share entry, got %v", err)
		}

		listData := shopping.NewShoppingListService(tableName, *client, marshaler)

		existingEntry, err := listData.CreateWithItemId(*approvedShare.ApproverId, data.ShoppingListInputDTO{
			Name:  aws.String("Giant"),
			Owner: aws.String("nobody@example,com"),
			Items: &[]data.ShoppingListItemDTO{
				{
					Name: "Giant",
				},
			},
		}, itemId)

		fmt.Printf("Approver ID: %s, %s", existingEntry.PK, *approvedShare.ApproverId)
		if err = updateHandler.Apply(modify); err != nil {
			t.Fatalf("Expected the modification to work, got %v", err)
		}

		updatedEntry, err := listData.Get(*approvedShare.ApproverId, existingEntry.SK)
		if err != nil {
			t.Fatalf("Expected the entry to still exist, %v", err)
		}

		if len(updatedEntry.Items) != 1 && !updatedEntry.Items[0].Completed {
			t.Fatalf("Expected the entry to sync, but got %v", updatedEntry)
		}

		if updatedEntry.UpdateToken == nil || !strings.EqualFold("abc-123", *updatedEntry.UpdateToken) {
			t.Fatalf("Expected the updated entry to have a token: %v", updatedEntry.UpdateToken)
		}
	})

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
		content, _ := millis.MarshalText()
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
