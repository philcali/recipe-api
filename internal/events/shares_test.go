package events

import (
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/shares"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/test"
)

func NewSharingRepository(t *testing.T) data.ShareRequestRepository {
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
	return shares.NewShareService(tableName, *client, marshaler)
}

func TestSharingEvents(t *testing.T) {
	sharingData := NewSharingRepository(t)

	t.Run("SharingEventHandlers", func(t *testing.T) {
		copyHandler := DefaultCopyApprovedRequestHandler(sharingData)
		deleteHandler := DefaultDeleteAssociatedHandler(sharingData)

		id := uuid.NewString()
		accountId := uuid.NewString()
		approvalStatus := data.REQUESTED
		request, err := sharingData.CreateWithItemId(accountId, data.ShareRequestInputDTO{
			Requester:      aws.String("nobody"),
			Approver:       aws.String("nobody2@email.com"),
			RequesterId:    aws.String(accountId),
			ApprovalStatus: &approvalStatus,
		}, id)
		if err != nil {
			t.Fatalf("Failed to insert a test record %v", err)
		}

		modify := events.DynamoDBEventRecord{
			EventName: "MODIFY",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute(fmt.Sprintf("%s:ShareRequest", accountId)),
				},
				NewImage: map[string]events.DynamoDBAttributeValue{
					"PK":             events.NewStringAttribute(fmt.Sprintf("%s:ShareRequest", accountId)),
					"SK":             events.NewStringAttribute(id),
					"approverId":     events.NewStringAttribute("nobody2"),
					"requester":      events.NewStringAttribute("nobody"),
					"approvalStatus": events.NewStringAttribute("APPROVED"),
				},
			},
		}

		remove := events.DynamoDBEventRecord{
			EventName: "REMOVE",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"PK": events.NewStringAttribute(fmt.Sprintf("%s:ShareRequest", "nobody2")),
					"SK": events.NewStringAttribute(id),
				},
				NewImage: map[string]events.DynamoDBAttributeValue{
					"PK":             events.NewStringAttribute(fmt.Sprintf("%s:ShareRequest", accountId)),
					"SK":             events.NewStringAttribute(id),
					"approverId":     events.NewStringAttribute("nobody2"),
					"requesterId":    events.NewStringAttribute("nobody"),
					"approvalStatus": events.NewStringAttribute("APPROVED"),
				},
			},
		}

		if !copyHandler.Filter(modify) {
			t.Fatalf("Failed to filter an approved record %v", modify)
		}

		if copyHandler.Filter(remove) {
			t.Fatalf("Failed to filter a remove record %v", remove)
		}

		err = copyHandler.Apply(modify)
		if err != nil {
			t.Fatalf("Failed to copy request %v", err)
		}

		newRequest, err := sharingData.Get("nobody2", id)
		if err != nil {
			t.Fatalf("Failed to copy share request %v, %v", modify, err)
		}

		if request.Requester != newRequest.Requester {
			t.Fatalf("Expected copy %s, got was %s", request.Requester, newRequest.Requester)
		}

		if deleteHandler.Filter(modify) {
			t.Fatalf("Failed on filter expectation for record %v", modify)
		}

		err = deleteHandler.Apply(remove)
		if err != nil {
			t.Fatalf("Failed to remove copied record: %v", err)
		}

		_, err = sharingData.Get("nobody", id)
		if err == nil {
			t.Fatalf("Expected a not found error for %v", remove)
		}
	})
}
