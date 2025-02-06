package token

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type TokenMarshaler interface {
	Marshal(accountId string, lastKey map[string]types.AttributeValue) (*string, error)

	Unmarshal(accountId string, token *string) (map[string]types.AttributeValue, error)
}
