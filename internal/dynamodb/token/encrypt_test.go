package token_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"philcali.me/recipes/internal/dynamodb/token"
)

func TestEncryptionMarshaler(t *testing.T) {
	marshaler := token.NewGCM()
	accountId := "012345678912"
	lastKey := map[string]types.AttributeValue{
		"name": &types.AttributeValueMemberS{Value: "Philip"},
	}

	t.Run("thing==Unmarshal(Marshal(thing))", func(t *testing.T) {
		token, err := marshaler.Marshal(accountId, lastKey)
		if err != nil {
			t.Fatalf("Failed to marshal token: %s", lastKey)
		}
		otherKey, err := marshaler.Unmarshal(accountId, token)
		if err != nil {
			t.Fatalf("Failed to unmarshal token: %s", err)
		}
		if value, ok := otherKey["name"]; ok {
			if svalue, ok := value.(*types.AttributeValueMemberS); ok {
				if svalue.Value != "Philip" {
					t.Errorf("otherKey name is %s", svalue.Value)
				}
			} else {
				t.Error("otherKey name is not an S type")
			}
		} else {
			t.Errorf("otherKey does not contain name: %s", otherKey)
		}
	})

	t.Run("len(token)==nil", func(t *testing.T) {
		var emptyMap map[string]types.AttributeValue
		token, err := marshaler.Marshal(accountId, emptyMap)
		if err != nil {
			t.Fatalf("Threw an error on marshal: %s", err)
		}
		if token != nil {
			t.Fatalf("Whoa %s is not nil!", token)
		}
	})

	t.Run("acountA!=accountB", func(t *testing.T) {
		token, err := marshaler.Marshal(accountId, lastKey)
		if err != nil {
			t.Fatalf("Failed to marshal token: %s", lastKey)
		}
		otherKey, err := marshaler.Unmarshal("987654321012", token)
		if err == nil {
			t.Fatalf("Expected an err but received, %v", otherKey)
		}
		if otherKey != nil {
			t.Fatalf("Should not have decrypted %s", otherKey)
		}
	})
}
