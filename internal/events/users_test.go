package events

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/dynamodb/users"
	"philcali.me/recipes/internal/test"
)

func NewUserService(t *testing.T) data.UserService {
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
	return users.NewUserService(tableName, *client, marshaler)
}

func TestUsers(t *testing.T) {
	userData := NewUserService(t)

	t.Run("ManagedUserHandler", func(t *testing.T) {
		handler := ManageGlobalUserHandler{
			Users: userData,
		}

		insert := events.DynamoDBEventRecord{
			EventName: "INSERT",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute("012345678912:Subscription"),
					"SK": events.NewStringAttribute("abc-123"),
				},
				NewImage: map[string]events.DynamoDBAttributeValue{
					"endpoint": events.NewStringAttribute("nobody@example.com"),
				},
			},
		}

		insertFailed := events.DynamoDBEventRecord{
			EventName: "INSERT",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute("012345678912:Recipe"),
					"SK": events.NewStringAttribute("abc-123"),
				},
			},
		}

		update := events.DynamoDBEventRecord{
			EventName: "UPDATE",
		}

		remove := events.DynamoDBEventRecord{
			EventName: "REMOVE",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute("012345678912:Subscription"),
					"SK": events.NewStringAttribute("abc-123"),
				},
				OldImage: map[string]events.DynamoDBAttributeValue{
					"endpoint": events.NewStringAttribute("nobody@example.com"),
				},
			},
		}

		t.Run("Filter", func(t *testing.T) {
			if !handler.Filter(insert) {
				t.Fatalf("Expected insert to filter")
			}

			if handler.Filter(update) || handler.Filter(insertFailed) {
				t.Fatalf("Expected update not to filter")
			}

			if !handler.Filter(remove) {
				t.Fatalf("Expected remove to filter")
			}
		})

		t.Run("Apply", func(t *testing.T) {
			if err := handler.Apply(insert); err != nil {
				t.Fatalf("Unexpected failure for insert: %v", err)
			}

			user, err := userData.Get(users.GLOBAL_ACCOUNT, "nobody@example.com")
			if err != nil {
				t.Fatalf("Failed to retrieve a global user %v", err)
			}

			if user.AccountId != "012345678912" {
				t.Fatalf("User does not have the correct account: %v", user)
			}

			if handler.Apply(insert) == nil {
				t.Errorf("Expected the duplicate to fail")
			}

			err = handler.Apply(remove)
			if err != nil {
				t.Fatalf("Failed to remove global user: %v", err)
			}

			results, err := userData.List(users.GLOBAL_ACCOUNT, data.QueryParams{
				Limit: 10,
			})

			if err != nil {
				t.Errorf("Failed to list users: %v", err)
			}

			if len(results.Items) != 0 {
				t.Errorf("Expected the user to be removed, but got: %v", results.Items)
			}

		})
	})
}
