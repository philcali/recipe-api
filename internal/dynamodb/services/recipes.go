package services

import (
	"context"
	"fmt"
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

type RecipeDynamoDBService struct {
	DynamoDB       dynamodb.Client
	TableName      string
	TokenMarshaler token.TokenMarshaler
}

func NewRecipeService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.RecipeDataService {
	return &RecipeDynamoDBService{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
	}
}

func getKey(dto data.RecipeDTO) (map[string]types.AttributeValue, error) {
	pk, err := attributevalue.Marshal(dto.PK)
	if err != nil {
		return nil, err
	}
	sk, err := attributevalue.Marshal(dto.SK)
	if err != nil {
		return nil, err
	}
	return map[string]types.AttributeValue{"PK": pk, "SK": sk}, nil
}

func _getPrimaryKey(accountId string) string {
	return fmt.Sprintf("%s:Recipe", accountId)
}

func (rs *RecipeDynamoDBService) ListRecipes(accountId string, params data.QueryParams) (data.QueryResults[data.RecipeDTO], error) {
	keyEx := expression.Key("PK").Equal(expression.Value(_getPrimaryKey(accountId)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return data.QueryResults[data.RecipeDTO]{}, err
	}
	var items []data.RecipeDTO
	var startKey map[string]types.AttributeValue
	startKey, err = rs.TokenMarshaler.Unmarshal(accountId, params.NextToken)
	if err != nil {
		return data.QueryResults[data.RecipeDTO]{}, err
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
		return data.QueryResults[data.RecipeDTO]{}, err
	}
	err = attributevalue.UnmarshalListOfMaps(output.Items, &items)
	if err != nil {
		return data.QueryResults[data.RecipeDTO]{}, err
	}
	token, err := rs.TokenMarshaler.Marshal(accountId, output.LastEvaluatedKey)
	if err != nil {
		return data.QueryResults[data.RecipeDTO]{}, err
	}
	return data.QueryResults[data.RecipeDTO]{
		Items:     items,
		NextToken: token,
	}, nil
}

func (rs *RecipeDynamoDBService) GetRecipe(accountId string, recipeId string) (data.RecipeDTO, error) {
	shim := data.RecipeDTO{PK: _getPrimaryKey(accountId), SK: recipeId}
	key, err := getKey(shim)
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
		return shim, exceptions.NotFound("recipe", recipeId)
	}
	err = attributevalue.UnmarshalMap(response.Item, &shim)
	return shim, err
}

func (rs *RecipeDynamoDBService) CreateRecipe(accountId string, input data.RecipeInputDTO) (data.RecipeDTO, error) {
	gid, err := uuid.NewUUID()
	if err != nil {
		return data.RecipeDTO{}, err
	}
	now := time.Now()
	shim := data.RecipeDTO{
		PK:                 _getPrimaryKey(accountId),
		SK:                 gid.String(),
		Name:               *input.Name,
		Instructions:       *input.Instructions,
		Ingredients:        *input.Ingredients,
		PrepareTimeMinutes: input.PrepareTimeMinutes,
		CreateTime:         now,
		UpdateTime:         now,
	}
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
			return shim, exceptions.Conflict("recipe", shim.SK)
		}
		return shim, err
	}
	return shim, err
}

func (rs *RecipeDynamoDBService) UpdateRecipe(accountId string, recipeId string, input data.RecipeInputDTO) (data.RecipeDTO, error) {
	shim := data.RecipeDTO{PK: _getPrimaryKey(accountId), SK: recipeId}
	key, err := getKey(shim)
	if err != nil {
		return shim, err
	}
	updateTime := time.Now()
	update := expression.Set(expression.Name("updateTime"), expression.Value(updateTime))
	condition := expression.Name("PK").AttributeExists().And(expression.Name("SK").AttributeExists())
	if input.Name != nil {
		update.Set(expression.Name("name"), expression.Value(input.Name))
	}
	if input.Instructions != nil {
		update.Set(expression.Name("instructions"), expression.Value(input.Instructions))
	}
	if input.Ingredients != nil {
		update.Set(expression.Name("ingredients"), expression.Value(input.Ingredients))
	}
	if input.PrepareTimeMinutes != nil {
		update.Set(expression.Name("prepareTimeMinutes"), expression.Value(input.PrepareTimeMinutes))
	}
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
			return shim, exceptions.NotFound("recipe", recipeId)
		}
		return shim, err
	}
	err = attributevalue.UnmarshalMap(response.Attributes, &shim)
	return shim, err
}

func (rs *RecipeDynamoDBService) DeleteRecipe(accountId string, recipeId string) error {
	shim := data.RecipeDTO{PK: _getPrimaryKey(accountId), SK: recipeId}
	key, err := getKey(shim)
	if err != nil {
		return err
	}
	_, err = rs.DynamoDB.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key:       key,
		TableName: aws.String(rs.TableName),
	})
	return err
}
