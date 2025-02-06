package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"philcali.me/recipes/internal/data"
)

type EncryptMode func(cipher.Block) (cipher.AEAD, error)

type EncryptionTokenMarshaler struct {
	Mode EncryptMode
}

func NewGCM() *EncryptionTokenMarshaler {
	return &EncryptionTokenMarshaler{
		Mode: cipher.NewGCM,
	}
}

func _encodeNextToken(token []byte) string {
	return base64.URLEncoding.EncodeToString(token)
}

func _convertLastKeyToToken(lastKey map[string]types.AttributeValue) ([]byte, error) {
	if len(lastKey) == 0 {
		return nil, nil
	}
	token := make(data.NextToken, len(lastKey))
	for key, value := range lastKey {
		innerMap := make(map[string]string, 1)
		if sv, ok := value.(*types.AttributeValueMemberS); ok {
			innerMap["S"] = sv.Value
		}
		if nv, ok := value.(*types.AttributeValueMemberN); ok {
			innerMap["N"] = nv.Value
		}
		if bv, ok := value.(*types.AttributeValueMemberB); ok {
			innerMap["B"] = string(bv.Value)
		}
		token[key] = innerMap
	}
	return json.Marshal(token)
}

func _decodeNextToken(encToken []byte) ([]byte, error) {
	dec := make([]byte, base64.URLEncoding.DecodedLen(len(encToken)))
	n, err := base64.URLEncoding.Decode(dec, encToken)
	if err != nil {
		return nil, err
	}
	return dec[:n], err
}

func _convertTokenToLastKey(token []byte) (map[string]types.AttributeValue, error) {
	if len(token) == 0 {
		return nil, nil
	}
	var nextToken data.NextToken
	err := json.Unmarshal(token, &nextToken)
	if err != nil {
		return nil, err
	}
	lastKey := make(map[string]types.AttributeValue, len(nextToken))
	for field, innerMap := range nextToken {
		if sv, ok := innerMap["S"]; ok {
			lastKey[field] = &types.AttributeValueMemberS{
				Value: sv,
			}
		}
		if nv, ok := innerMap["N"]; ok {
			lastKey[field] = &types.AttributeValueMemberN{
				Value: nv,
			}
		}
		if bv, ok := innerMap["B"]; ok {
			lastKey[field] = &types.AttributeValueMemberB{
				Value: []byte(bv),
			}
		}
	}
	return lastKey, nil
}

func _hash(accountId string) []byte {
	hash := sha256.New()
	hash.Write([]byte(accountId))
	return hash.Sum(nil)
}

func _mode(marshaller *EncryptionTokenMarshaler, accountId string) (cipher.AEAD, error) {
	key, err := aes.NewCipher(_hash(accountId))
	if err != nil {
		return nil, err
	}
	return marshaller.Mode(key)
}

func (em *EncryptionTokenMarshaler) Marshal(accountId string, lastKey map[string]types.AttributeValue) ([]byte, error) {
	var err error
	var bytes []byte
	serialized, err := _convertLastKeyToToken(lastKey)
	if err != nil || serialized == nil {
		return serialized, err
	}
	aesgcm, err := _mode(em, accountId)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, serialized, nil)
	payload := map[string]string{
		"ciphertext": hex.EncodeToString(ciphertext),
		"nonce":      hex.EncodeToString(nonce),
	}
	if b, err := json.Marshal(payload); err == nil {
		s := _encodeNextToken(b)
		bytes = []byte(strings.TrimSpace(s))
	}
	return bytes, err
}

func (em *EncryptionTokenMarshaler) Unmarshal(accountId string, token []byte) (map[string]types.AttributeValue, error) {
	if len(token) == 0 {
		return nil, nil
	}
	decToken, err := _decodeNextToken(token)
	if err != nil {
		return nil, err
	}
	var payload map[string]string
	if err := json.Unmarshal(decToken, &payload); err != nil {
		return nil, err
	}
	aesgcm, err := _mode(em, accountId)
	if err != nil {
		return nil, err
	}
	ciphertext, _ := hex.DecodeString(payload["ciphertext"])
	nonce, _ := hex.DecodeString(payload["nonce"])
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return _convertTokenToLastKey(plaintext)
}
