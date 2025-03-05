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

func _sharedResourceFilter(t string) bool {
	return t == "Recipe" || t == "ShoppingList"
}

func _copyShareResource(tableName string, ownerId string, condition string, record events.DynamoDBEventRecord, ddb *dynamodb.Client, shareRepo data.ShareRequestRepository) error {
	var nextToken *string
	truncated := true
	for truncated {
		sharing, err := shareRepo.List(ownerId, data.QueryParams{
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
			_, err := ddb.PutItem(context.TODO(), &dynamodb.PutItemInput{
				Item:                converted,
				TableName:           aws.String(tableName),
				ConditionExpression: aws.String(condition),
			})

			if err != nil {
				if _, ok := err.(*types.ConditionalCheckFailedException); ok {
					continue
				}
				return err
			}
		}

		nextToken = sharing.NextToken
		truncated = nextToken != nil
	}
	return nil
}

type UpdateSharedResourceHandler struct {
	Sharing   data.ShareRequestRepository
	DynamoDB  *dynamodb.Client
	TableName string
}

func (uh *UpdateSharedResourceHandler) Filter(record events.DynamoDBEventRecord) bool {
	pk := record.Change.Keys["PK"]
	parts := strings.Split(pk.String(), ":")
	updateToken, isTokenSet := record.Change.OldImage["updateToken"]
	return record.EventName == "MODIFY" &&
		_sharedResourceFilter(parts[1]) &&
		(!isTokenSet || updateToken.String() != record.Change.NewImage["updateToken"].String())
}

func (uh *UpdateSharedResourceHandler) Apply(record events.DynamoDBEventRecord) error {
	pk := record.Change.Keys["PK"]
	parts := strings.Split(pk.String(), ":")
	ownerId := parts[0]
	return _copyShareResource(
		uh.TableName,
		ownerId,
		"attribute_exists(PK) and attribute_exists(SK)",
		record,
		uh.DynamoDB,
		uh.Sharing,
	)
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
	return record.EventName == "INSERT" &&
		_sharedResourceFilter(parts[1]) &&
		(record.Change.NewImage["shared"].IsNull() || !record.Change.NewImage["shared"].Boolean())
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
	return _copyShareResource(
		ch.TableName,
		ownerId,
		"attribute_not_exists(PK) and attribute_not_exists(SK)",
		record,
		ch.DynamoDB,
		ch.Sharing,
	)
}
