package events

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"philcali.me/recipes/internal/data"
)

type DeleteAssociatedShareHandler struct {
	Sharing data.ShareRequestRepository
}

func (dh *DeleteAssociatedShareHandler) Filter(record events.DynamoDBEventRecord) bool {
	pk := record.Change.Keys["PK"]
	parts := strings.Split(pk.String(), ":")
	return record.EventName == "REMOVE" && parts[1] == "ShareRequest"
}

func (dh *DeleteAssociatedShareHandler) Apply(record events.DynamoDBEventRecord) error {
	pk := record.Change.Keys["PK"]
	parts := strings.Split(pk.String(), ":")
	requesterId := record.Change.OldImage["requesterId"].String()
	// What is being deleted is the request made by the requester
	if requesterId == parts[0] {
		// If the approver approved this request, then delete it's association
		if approverId, ok := record.Change.OldImage["approverId"]; ok {
			return dh.Sharing.Delete(approverId.String(), record.Change.Keys["SK"].String())
		}
	} else {
		// The approved request is being deleted, delete the requester associate
		return dh.Sharing.Delete(requesterId, record.Change.Keys["SK"].String())
	}
	// Do nothing
	return nil
}

type CopyApprovedShareRequestHandler struct {
	Sharing data.ShareRequestRepository
}

func (ch *CopyApprovedShareRequestHandler) Filter(record events.DynamoDBEventRecord) bool {
	pk := record.Change.Keys["PK"]
	parts := strings.Split(pk.String(), ":")
	return record.EventName == "MODIFY" &&
		parts[1] == "ShareRequest" &&
		!record.Change.NewImage["approverId"].IsNull() &&
		record.Change.NewImage["approvalStatus"].String() == "APPROVED"
}

func (ch *CopyApprovedShareRequestHandler) Apply(record events.DynamoDBEventRecord) error {
	approverAccount := record.Change.NewImage["approverId"].String()
	itemId := record.Change.NewImage["SK"].String()
	pk := record.Change.NewImage["PK"]
	parts := strings.Split(pk.String(), ":")
	status := data.APPROVED
	_, err := ch.Sharing.CreateWithItemId(approverAccount, data.ShareRequestInputDTO{
		Requester:      aws.String(record.Change.NewImage["requester"].String()),
		RequesterId:    aws.String(parts[0]),
		ApproverId:     aws.String(approverAccount),
		ApprovalStatus: &status,
		Approver:       aws.String(itemId),
	}, itemId)
	return err
}

func DefaultCopyApprovedRequestHandler(db data.ShareRequestRepository) *CopyApprovedShareRequestHandler {
	return &CopyApprovedShareRequestHandler{
		Sharing: db,
	}
}

func DefaultDeleteAssociatedHandler(db data.ShareRequestRepository) *DeleteAssociatedShareHandler {
	return &DeleteAssociatedShareHandler{
		Sharing: db,
	}
}
