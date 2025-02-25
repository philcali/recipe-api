package events

import (
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"philcali.me/recipes/internal/data"
)

// Five years for things to enpire
const EXPIRY_LOG = int(time.Hour + (24 * 365 * 5))

func _getRecordImage(record events.DynamoDBEventRecord) map[string]events.DynamoDBAttributeValue {
	if record.Change.NewImage != nil {
		return record.Change.NewImage
	} else {
		return record.Change.OldImage
	}
}

func _convertAttribute(value events.DynamoDBAttributeValue) interface{} {
	switch value.DataType() {
	case events.DataTypeString:
		return value.String()
	case events.DataTypeNumber:
		return value.Number()
	case events.DataTypeBoolean:
		return value.Boolean()
	case events.DataTypeBinary:
		return value.Binary()
	case events.DataTypeStringSet:
		return value.StringSet()
	case events.DataTypeBinarySet:
		return value.BinarySet()
	case events.DataTypeNumberSet:
		return value.NumberSet()
	case events.DataTypeList:
		ls := make([]interface{}, len(value.List()))
		for _, item := range value.List() {
			ls = append(ls, _convertAttribute(item))
		}
		return ls
	case events.DataTypeMap:
		return _flattenResourceProperties(value.Map())
	}
	return nil
}

func _flattenResourceProperties(image map[string]events.DynamoDBAttributeValue) *map[string]interface{} {
	if image == nil {
		return nil
	}
	properties := make(map[string]interface{}, len(image))
	for field, value := range image {
		switch field {
		case "PK":
			fallthrough
		case "SK":
			fallthrough
		case "GS1-PK":
			continue
		default:
			properties[field] = _convertAttribute(value)
		}
	}
	return &properties
}

type CreateAuditEntryHandler struct {
	Audit         data.AuditRepository
	ResourceTypes []string
}

func (ch *CreateAuditEntryHandler) Filter(record events.DynamoDBEventRecord) bool {
	pk := _getRecordImage(record)["PK"]
	parts := strings.Split(pk.String(), ":")
	for _, t := range ch.ResourceTypes {
		if t == parts[1] {
			return true
		}
	}
	return false
}

func (ch *CreateAuditEntryHandler) Apply(record events.DynamoDBEventRecord) error {
	image := _getRecordImage(record)
	pk := image["PK"]
	resourceId := image["SK"].String()
	parts := strings.Split(pk.String(), ":")
	if parts[1] == "ApiToken" {
		parts = strings.Split(image["GS1-PK"].String(), ":")
	}
	var action string
	switch record.EventName {
	case "INSERT":
		action = "CREATED"
	case "MODIFY":
		action = "UPDATED"
	case "REMOVE":
		action = "DELETED"
	}
	accountId := parts[0]
	_, err := ch.Audit.Create(accountId, data.AuditInputDTO{
		AccountId:    &accountId,
		ResourceId:   &resourceId,
		ResourceType: aws.String(parts[1]),
		Action:       &action,
		ExpiresIn:    aws.Int(int(time.Now().UnixMilli()) + EXPIRY_LOG),
		NewValues:    _flattenResourceProperties(record.Change.NewImage),
		OldValues:    _flattenResourceProperties(record.Change.OldImage),
	})
	return err
}

func DefaultAuditHandler(db data.AuditRepository) *CreateAuditEntryHandler {
	return &CreateAuditEntryHandler{
		Audit: db,
		ResourceTypes: []string{
			"Recipe",
			"ApiToken",
			"Settings",
			"ShoppingList",
			"ShareRequest",
		},
	}
}
