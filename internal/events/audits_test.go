package events

import (
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/audits"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/test"
)

func NewAuditRepository(t *testing.T) data.AuditRepository {
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
	return audits.NewAuditService(tableName, *client, marshaler)
}

func TestAudits(t *testing.T) {
	auditData := NewAuditRepository(t)

	t.Run("AuditHandler", func(t *testing.T) {
		handler := DefaultAuditHandler(auditData)

		t.Run("RecipeAudit", func(t *testing.T) {
			id := uuid.NewString()
			accountId := uuid.NewString()
			pk := fmt.Sprintf("%s:Recipe", accountId)
			insert := events.DynamoDBEventRecord{
				EventName: "INSERT",
				Change: events.DynamoDBStreamRecord{
					NewImage: map[string]events.DynamoDBAttributeValue{
						"name": events.NewStringAttribute("A Tasty Treat"),
						"SK":   events.NewStringAttribute(id),
						"PK":   events.NewStringAttribute(pk),
					},
				},
			}

			modify := events.DynamoDBEventRecord{
				EventName: "MODIFY",
				Change: events.DynamoDBStreamRecord{
					NewImage: map[string]events.DynamoDBAttributeValue{
						"name": events.NewStringAttribute("A Very Tasty Treat"),
						"SK":   events.NewStringAttribute(id),
						"PK":   events.NewStringAttribute(pk),
					},
				},
			}

			remove := events.DynamoDBEventRecord{
				EventName: "REMOVE",
				Change: events.DynamoDBStreamRecord{
					OldImage: map[string]events.DynamoDBAttributeValue{
						"name": events.NewStringAttribute("A Very Tasty Treat"),
						"SK":   events.NewStringAttribute(id),
						"PK":   events.NewStringAttribute(pk),
					},
				},
			}

			records := []events.DynamoDBEventRecord{
				insert,
				modify,
				remove,
			}

			for _, record := range records {
				if !handler.Filter(record) {
					t.Fatalf("Expected true for %v", record)
				}
				err := handler.Apply(record)
				if err != nil {
					t.Fatalf("Failed to create audit entry for %v: %v", record, err)
				}
				listEntry, err := auditData.List(accountId, data.QueryParams{
					Limit: 1,
				})
				if err != nil {
					t.Fatalf("Failed to list audit entry for %v", err)
				}
				item := listEntry.Items[0]
				expectedMessage := handler.Formats["Recipe"](record)
				if item.Message != *expectedMessage {
					t.Fatalf("Expected '%s', but got '%s'", item.Message, *expectedMessage)
				}
				if err := auditData.Delete(accountId, item.SK); err != nil {
					t.Fatalf("Failed to delete %v", item)
				}
			}
		})
	})
}
