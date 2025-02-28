package shares

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/services"
	"philcali.me/recipes/internal/dynamodb/token"
)

func NewShareService(tableName string, client dynamodb.Client, marshaler token.TokenMarshaler) data.Repository[data.ShareRequestDTO, data.ShareRequestInputDTO] {
	return &services.RepositoryDynamoDBService[data.ShareRequestDTO, data.ShareRequestInputDTO]{
		DynamoDB:       client,
		TableName:      tableName,
		TokenMarshaler: marshaler,
		Name:           "ShareRequest",
		Shim: func(pk, sk string) data.ShareRequestDTO {
			return data.ShareRequestDTO{PK: pk, SK: sk}
		},
		OnCreate: func(srid data.ShareRequestInputDTO, t time.Time, pk, sk string) data.ShareRequestDTO {
			request := data.ShareRequestDTO{
				PK:             pk,
				SK:             sk,
				Requester:      *srid.Requester,
				RequesterId:    srid.RequesterId,
				Approver:       srid.Approver,
				ApproverId:     srid.ApproverId,
				ApprovalStatus: *srid.ApprovalStatus,
				ExpiresIn:      srid.ExpiresIn,
				CreateTime:     t,
				UpdateTime:     t,
			}
			if srid.Approver != nil {
				request.FirstIndex = aws.String(fmt.Sprintf("%s:ShareRequest", *srid.Approver))
			} else {
				request.FirstIndex = aws.String("ShareRequest")
			}
			return request
		},
		OnUpdate: func(srid data.ShareRequestInputDTO, ub expression.UpdateBuilder) {
			if srid.ApprovalStatus != nil {
				ub.Set(expression.Name("approvalStatus"), expression.Value(srid.ApprovalStatus))
				ub.Remove(expression.Name("GS1-PK"))
				if strings.EqualFold(string(data.APPROVED), string(*srid.ApprovalStatus)) {
					ub.Set(expression.Name("approverId"), expression.Value(srid.ApproverId))
					ub.Remove(expression.Name("expiresIn"))
				}
			}
		},
	}
}
