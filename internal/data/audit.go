package data

import "time"

type AuditDTO struct {
	PK         string    `dynamodbav:"PK"`
	SK         string    `dynamodbav:"SK"`
	FirstIndex string    `dynamodbav:"GS1-PK"`
	Message    string    `dynamodbav:"message"`
	ExpiresIn  *int      `dynamodbav:"expiresIn"`
	CreateTime time.Time `dynamodbav:"createTime"`
	UpdateTime time.Time `dynamodbav:"updateTime"`
}

type AuditInputDTO struct {
	Message   *string `dynamodbav:"message"`
	AccountId *string `dynamodbav:"accountId"`
	ExpiresIn *int    `dynamodbav:"expiresIn"`
}

type AuditRepository interface {
	Repository[AuditDTO, AuditInputDTO]
}
