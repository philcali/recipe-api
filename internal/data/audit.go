package data

import "time"

type AuditDTO struct {
	PK           string                  `dynamodbav:"PK"`
	SK           string                  `dynamodbav:"SK"`
	FirstIndex   string                  `dynamodbav:"GS1-PK"`
	ResourceId   string                  `dynamodbav:"resourceId"`
	ResourceType string                  `dynamodbav:"resourceType"`
	Action       string                  `dynamodbav:"action"`
	NewValues    *map[string]interface{} `dynamodbav:"newValues"`
	OldValues    *map[string]interface{} `dynamodbav:"oldValues"`
	ExpiresIn    *int                    `dynamodbav:"expiresIn"`
	CreateTime   time.Time               `dynamodbav:"createTime"`
	UpdateTime   time.Time               `dynamodbav:"updateTime"`
}

type AuditInputDTO struct {
	AccountId    *string                 `dynamodbav:"accountId"`
	ResourceId   *string                 `dynamodbav:"resourceId"`
	ResourceType *string                 `dynamodbav:"resourceType"`
	Action       *string                 `dynamodbav:"action"`
	NewValues    *map[string]interface{} `dynamodbav:"newValues"`
	OldValues    *map[string]interface{} `dynamodbav:"oldValues"`
	ExpiresIn    *int                    `dynamodbav:"expiresIn"`
}

type AuditRepository interface {
	Repository[AuditDTO, AuditInputDTO]
}
