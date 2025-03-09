package audits

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type AuditService struct {
	data      data.AuditRepository
	indexName string
}

func NewRouteWithIndex(data data.AuditRepository, indexName string) routes.Service {
	return &AuditService{
		data:      data,
		indexName: indexName,
	}
}

func NewRoute(data data.AuditRepository) routes.Service {
	return NewRouteWithIndex(data, os.Getenv("INDEX_NAME_1"))
}

func _stripFields(event events.APIGatewayV2HTTPRequest) func(d data.AuditDTO) Audit {
	return func(d data.AuditDTO) Audit {
		newValues := d.NewValues
		oldValues := d.OldValues
		params := event.QueryStringParameters
		if params != nil {
			if stripFields, ok := params["stripFields"]; ok {
				fieldNames := strings.Split(stripFields, ",")
				for _, fieldName := range fieldNames {
					if newValues != nil {
						delete(*newValues, fieldName)
					}
					if oldValues != nil {
						delete(*oldValues, fieldName)
					}
				}
			}
		}
		return _convertAudit(data.AuditDTO{
			CreateTime:   d.CreateTime,
			UpdateTime:   d.UpdateTime,
			Action:       d.Action,
			ResourceType: d.ResourceType,
			ResourceId:   d.ResourceId,
			NewValues:    newValues,
			OldValues:    oldValues,
			SK:           d.SK,
		})
	}
}

func _convertAudit(auditDTO data.AuditDTO) Audit {
	return Audit{
		CreateTime:   auditDTO.CreateTime,
		UpdateTime:   auditDTO.UpdateTime,
		Action:       auditDTO.Action,
		ResourceType: auditDTO.ResourceType,
		ResourceId:   auditDTO.ResourceId,
		NewValues:    auditDTO.NewValues,
		OldValues:    auditDTO.OldValues,
		Id:           auditDTO.SK,
	}
}

func (as *AuditService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/audits":             util.AuthorizedRoute(as.ListAudits),
		"DELETE:/audits/:auditId": util.AuthorizedRoute(as.DeleteAudit),
	}
}

func (as *AuditService) ListAudits(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return util.SerializeListByIndex(as.data, _stripFields(event), as.indexName, event, ctx)
}

func (as *AuditService) DeleteAudit(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	err := as.data.Delete(util.Username(ctx), util.RequestParam(ctx, "auditId"))
	return util.SerializeResponseNoContent(err)
}
