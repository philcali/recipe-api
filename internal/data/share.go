package data

import "time"

type ApprovalStatus string

const (
	REQUESTED ApprovalStatus = "REQUESTED"
	APPROVED  ApprovalStatus = "APPROVED"
	REJECTED  ApprovalStatus = "REJECTED"
)

type ShareRequestDTO struct {
	PK             string         `dynamodbav:"PK"`
	SK             string         `dynamodbav:"SK"`
	FirstIndex     string         `dynamodbav:"GS1-PK"`
	Requester      string         `dynamodbav:"requester"`
	Approver       string         `dynamodbav:"approver"`
	ApproverId     *string        `dynamodbav:"approverId"`
	ApprovalStatus ApprovalStatus `dynamodbav:"approvalStatus"`
	ExpiresIn      *int           `dynamodbav:"expiresIn"`
	CreateTime     time.Time      `dynamodbav:"createTime"`
	UpdateTime     time.Time      `dynamodbav:"updateTime"`
}

type ShareRequestInputDTO struct {
	Requester      *string         `dynamodbav:"requester"`
	Approver       *string         `dynamodbav:"approver"`
	ApproverId     *string         `dynamodbav:"approverId"`
	ApprovalStatus *ApprovalStatus `dynamodbav:"approvalStatus"`
	ExpiresIn      *int            `dynamodbav:"expiresIn"`
}

type ShareRequestRepository interface {
	Repository[ShareRequestDTO, ShareRequestInputDTO]
}
