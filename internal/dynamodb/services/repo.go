package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/token"
	"philcali.me/recipes/internal/exceptions"
)

type RepositoryDynamoDBService[T interface{}, I interface{}] struct {
	DynamoDB       dynamodb.Client
	TableName      string
	TokenMarshaler token.TokenMarshaler
	Name           string
	Shim           func(pk string, sk string) T
	GetPK          func(T) string
	GetSK          func(T) string
	OnCreate       func(I, time.Time, string, string) T
	OnUpdate       func(I, expression.UpdateBuilder)
}

func _getPrimaryKey(accountId string, name string) string {
	return fmt.Sprintf("%s:%s", accountId, name)
}

func _getKey(pks string, sks string) (map[string]types.AttributeValue, error) {
	pk, err := attributevalue.Marshal(pks)
	if err != nil {
		return nil, err
	}
	sk, err := attributevalue.Marshal(sks)
	if err != nil {
		return nil, err
	}
	return map[string]types.AttributeValue{"PK": pk, "SK": sk}, nil
}

func (rs *RepositoryDynamoDBService[T, I]) List(accountId string, params data.QueryParams) (data.QueryResults[T], error) {
	keyEx := expression.Key("PK").Equal(expression.Value(_getPrimaryKey(accountId, rs.Name)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return data.QueryResults[T]{}, err
	}
	var items []T
	var startKey map[string]types.AttributeValue
	startKey, err = rs.TokenMarshaler.Unmarshal(accountId, params.NextToken)
	if err != nil {
		return data.QueryResults[T]{}, err
	}
	output, err := rs.DynamoDB.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(rs.TableName),
		Limit:                     params.GetLimit(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ExclusiveStartKey:         startKey,
	})
	if err != nil {
		return data.QueryResults[T]{}, err
	}
	err = attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		return data.QueryResults[T]{}, err
	}
	token, err := rs.TokenMarshaler.Marshal(accountId, output.LastEvaluatedKey)
	if err != nil {
		return data.QueryResults[T]{}, err
	}
	return data.QueryResults[T]{
		Items:     items,
		NextToken: token,
	}, nil
}

func (rs *RepositoryDynamoDBService[T, I]) Create(accountId string, input I) (T, error) {
	gid, _ := uuid.NewUUID()
	now := time.Now()
	shim := rs.OnCreate(input, now, _getPrimaryKey(accountId, rs.Name), gid.String())
	item, err := attributevalue.MarshalMap(shim)
	if err != nil {
		return shim, err
	}
	expr, err := expression.NewBuilder().WithCondition(expression.Name("PK").AttributeNotExists().And(expression.Name("SK").AttributeNotExists())).Build()
	if err != nil {
		return shim, err
	}
	_, err = rs.DynamoDB.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:                     item,
		TableName:                aws.String(rs.TableName),
		ConditionExpression:      expr.Condition(),
		ExpressionAttributeNames: expr.Names(),
	})
	if err != nil {
		if _, ok := err.(*types.ConditionalCheckFailedException); ok {
			return shim, exceptions.Conflict(strings.ToLower(rs.Name), rs.GetSK(shim))
		}
		return shim, err
	}
	return shim, err
}

func (rs *RepositoryDynamoDBService[T, I]) Update(accountId string, itemId string, input I) (T, error) {
	pk := _getPrimaryKey(accountId, rs.Name)
	shim := rs.Shim(pk, itemId)
	key, err := _getKey(pk, itemId)
	if err != nil {
		return shim, err
	}
	updateTime := time.Now()
	update := expression.Set(expression.Name("updateTime"), expression.Value(updateTime))
	condition := expression.Name("PK").AttributeExists().And(expression.Name("SK").AttributeExists())
	rs.OnUpdate(input, update)
	expr, err := expression.NewBuilder().WithCondition(condition).WithUpdate(update).Build()
	if err != nil {
		return shim, err
	}
	response, err := rs.DynamoDB.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 aws.String(rs.TableName),
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
		ReturnValues:              types.ReturnValueAllNew,
	})
	if err != nil {
		if _, ok := err.(*types.ConditionalCheckFailedException); ok {
			return shim, exceptions.NotFound(strings.ToLower(rs.Name), itemId)
		}
		return shim, err
	}
	err = attributevalue.UnmarshalMap(response.Attributes, &shim)
	return shim, err
}

func (rs *RepositoryDynamoDBService[T, I]) Get(accountId string, itemId string) (T, error) {
	pk := _getPrimaryKey(accountId, rs.Name)
	shim := rs.Shim(pk, itemId)
	key, err := _getKey(pk, itemId)
	if err != nil {
		return shim, err
	}
	response, err := rs.DynamoDB.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(rs.TableName),
		Key:       key,
	})
	if err != nil {
		return shim, err
	}
	if response.Item == nil {
		return shim, exceptions.NotFound(strings.ToLower(rs.Name), itemId)
	}
	err = attributevalue.UnmarshalMap(response.Item, &shim)
	return shim, err
}

func (rs *RepositoryDynamoDBService[T, I]) Delete(accountId string, itemId string) error {
	pk := _getPrimaryKey(accountId, rs.Name)
	key, err := _getKey(pk, itemId)
	if err != nil {
		return err
	}
	_, err = rs.DynamoDB.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key:       key,
		TableName: aws.String(rs.TableName),
	})
	return err
}
