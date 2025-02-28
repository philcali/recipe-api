package events

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"philcali.me/recipes/internal/data"
)

func _convertStreamAttribute(attr events.DynamoDBAttributeValue) types.AttributeValue {
	switch attr.DataType() {
	case events.DataTypeBoolean:
		return &types.AttributeValueMemberBOOL{
			Value: attr.Boolean(),
		}
	case events.DataTypeString:
		return &types.AttributeValueMemberS{
			Value: attr.String(),
		}
	case events.DataTypeBinary:
		return &types.AttributeValueMemberB{
			Value: attr.Binary(),
		}
	case events.DataTypeList:
		ls := make([]types.AttributeValue, len(attr.List()))
		for i, item := range attr.List() {
			ls[i] = _convertStreamAttribute(item)
		}
		return &types.AttributeValueMemberL{
			Value: ls,
		}
	case events.DataTypeNull:
		return &types.AttributeValueMemberNULL{
			Value: attr.IsNull(),
		}
	case events.DataTypeBinarySet:
		return &types.AttributeValueMemberBS{
			Value: attr.BinarySet(),
		}
	case events.DataTypeNumber:
		return &types.AttributeValueMemberN{
			Value: attr.Number(),
		}
	case events.DataTypeStringSet:
		return &types.AttributeValueMemberNS{
			Value: attr.NumberSet(),
		}
	case events.DataTypeMap:
		ms := make(map[string]types.AttributeValue, len(attr.Map()))
		for field, value := range attr.Map() {
			ms[field] = _convertStreamAttribute(value)
		}
		return &types.AttributeValueMemberM{
			Value: ms,
		}
	}
	return nil
}

func _convertStreamImageToItem(otherAccountId string, image map[string]events.DynamoDBAttributeValue) map[string]types.AttributeValue {
	converted := make(map[string]types.AttributeValue, len(image))
	for field, value := range image {
		if field == "PK" {
			parts := strings.Split(value.String(), ":")
			converted["PK"] = &types.AttributeValueMemberS{
				Value: fmt.Sprintf("%s:%s", otherAccountId, parts[1]),
			}
		} else if field == "shared" {
			converted["shared"] = &types.AttributeValueMemberBOOL{
				Value: true,
			}
		} else {
			converted[field] = _convertStreamAttribute(value)
		}
	}
	return converted
}

type CopySharingResourceHandler struct {
	Setting   data.SettingsRepository
	Sharing   data.ShareRequestRepository
	DynamoDB  *dynamodb.Client
	TableName string
}

func (ch *CopySharingResourceHandler) Filter(record events.DynamoDBEventRecord) bool {
	pk := record.Change.Keys["PK"]
	parts := strings.Split(pk.String(), ":")
	return record.EventName == "INSERT" && (parts[1] == "Recipe" || parts[1] == "ShoppingList") && record.Change.NewImage["shared"].IsNull()
}

func (ch *CopySharingResourceHandler) Apply(record events.DynamoDBEventRecord) error {
	pk := record.Change.Keys["PK"]
	parts := strings.Split(pk.String(), ":")
	ownerId := parts[0]
	resourceType := parts[1]
	s, err := ch.Setting.Get(ownerId, "Global")
	if err != nil {
		return nil
	}
	if resourceType == "Recipe" && !s.AutoShareRecipes {
		return nil
	}
	if resourceType == "ShoppingList" && !s.AutoShareLists {
		return nil
	}
	var nextToken *string
	truncated := true
	for truncated {
		sharing, err := ch.Sharing.List(ownerId, data.QueryParams{
			Limit:     100,
			NextToken: nextToken,
		})

		if err != nil {
			return err
		}

		for _, item := range sharing.Items {
			if item.ApprovalStatus != data.APPROVED {
				continue
			}

			otherAccountId := item.RequesterId
			if strings.EqualFold(*item.RequesterId, ownerId) {
				otherAccountId = item.ApproverId
			}

			converted := _convertStreamImageToItem(*otherAccountId, record.Change.NewImage)
			_, err := ch.DynamoDB.PutItem(context.TODO(), &dynamodb.PutItemInput{
				Item:      converted,
				TableName: aws.String(ch.TableName),
			})

			if err != nil {
				return err
			}
		}

		nextToken = sharing.NextToken
		truncated = nextToken != nil
	}
	return nil
}
