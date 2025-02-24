package events

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"philcali.me/recipes/internal/data"
)

// Five years for things to enpire
const EXPIRY_LOG = int(time.Hour + (24 * 365 * 5))

type AuditMessageFormat func(record events.DynamoDBEventRecord) *string

func _getRecordImage(record events.DynamoDBEventRecord) map[string]events.DynamoDBAttributeValue {
	if record.Change.NewImage != nil {
		return record.Change.NewImage
	} else {
		return record.Change.OldImage
	}
}

func _formatRecipe(record events.DynamoDBEventRecord) *string {
	image := _getRecordImage(record)
	var action string
	switch record.EventName {
	case "INSERT":
		action = "created"
	case "MODIFY":
		action = "updated"
	case "REMOVE":
		action = "deleted"
	}
	name := image["name"].String()
	id := image["SK"].String()
	return aws.String(fmt.Sprintf("Recipe %s (%s) was %s", id, name, action))
}

func _formatApiToken(record events.DynamoDBEventRecord) *string {
	image := _getRecordImage(record)
	var action string
	switch record.EventName {
	case "INSERT":
		action = "created"
	case "MODIFY":
		action = "updated"
	case "REMOVE":
		action = "deleted"
	}
	name := image["name"].String()
	id := image["SK"].String()
	return aws.String(fmt.Sprintf("API token %s (%s) was %s", id, name, action))
}

func _formatSetting(record events.DynamoDBEventRecord) *string {
	action := "updated"
	if record.EventName == "INSERT" {
		action = "applied"
	}
	return aws.String(fmt.Sprintf("New settings were %s", action))
}

func _formatList(record events.DynamoDBEventRecord) *string {
	image := _getRecordImage(record)
	var action string
	switch record.EventName {
	case "INSERT":
		action = "was created"
	case "REMOVE":
		action = "was deleted"
		exp, ok := image["expiresIn"]
		if ok {
			millis, err := strconv.Atoi(exp.Number())
			if err != nil {
				return nil
			}
			if millis < int(time.Now().UnixMilli()) {
				action = "has expired"
			}
		}
	case "MODIFY":
		action = "was updated"
	}
	name := image["name"].String()
	id := image["SK"].String()
	return aws.String(fmt.Sprintf("Shopping list %s (%s) %s", id, name, action))
}

func _formatShare(record events.DynamoDBEventRecord) *string {
	image := _getRecordImage(record)
	var action string
	switch record.EventName {
	case "INSERT":
		action = "created"
	case "MODIFY":
		action = "updated"
	case "REMOVE":
		action = "deleted"
		exp, ok := image["expiresIn"]
		if ok {
			millis, err := strconv.Atoi(exp.Number())
			if err != nil {
				return nil
			}
			if millis < int(time.Now().UnixMilli()) {
				action = "expired"
			}
		}
	}
	approver := image["name"].String()
	status := image["approvalStatus"].String()
	id := image["SK"].String()
	return aws.String(fmt.Sprintf("Share request %s (%s) %s was %s", id, approver, status, action))
}

type CreateAuditEntryHandler struct {
	Audit   data.AuditRepository
	Formats map[string]AuditMessageFormat
}

func (ch *CreateAuditEntryHandler) Filter(record events.DynamoDBEventRecord) bool {
	pk := _getRecordImage(record)["PK"]
	parts := strings.Split(pk.String(), ":")
	_, ok := ch.Formats[parts[1]]
	return ok
}

func (ch *CreateAuditEntryHandler) Apply(record events.DynamoDBEventRecord) error {
	pk := _getRecordImage(record)["PK"]
	parts := strings.Split(pk.String(), ":")
	format := ch.Formats[parts[1]]
	message := format(record)
	if message == nil {
		return nil
	}
	_, err := ch.Audit.Create(parts[0], data.AuditInputDTO{
		Message:   message,
		AccountId: &parts[0],
		ExpiresIn: aws.Int(int(time.Now().Add(time.Duration(EXPIRY_LOG)).UnixMilli())),
	})
	return err
}

func DefaultAuditHandler(db data.AuditRepository) *CreateAuditEntryHandler {
	return &CreateAuditEntryHandler{
		Audit: db,
		Formats: map[string]AuditMessageFormat{
			"Recipe":       _formatRecipe,
			"ApiToken":     _formatApiToken,
			"Settings":     _formatSetting,
			"ShoppingList": _formatList,
			"ShareRequest": _formatShare,
		},
	}
}
