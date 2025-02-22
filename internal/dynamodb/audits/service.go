package audits

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

func NewAuditService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.AuditDTO, data.AuditInputDTO] {
	return &services.RepositoryDynamoDBService[data.AuditDTO, data.AuditInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "Audit",
		Shim: func(pk, sk string) data.AuditDTO {
			return data.AuditDTO{PK: pk, SK: sk}
		},
		OnCreate: func(aid data.AuditInputDTO, t time.Time, pk, sk string) data.AuditDTO {
			return data.AuditDTO{
				PK:         pk,
				SK:         sk,
				FirstIndex: fmt.Sprintf("%s:Audit", *aid.AccountId),
				Message:    *aid.Message,
				CreateTime: t,
				UpdateTime: t,
			}
		},
	}
}
