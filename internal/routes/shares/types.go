package shares

import (
	"time"

	"philcali.me/recipes/internal/data"
)

type ShareRequest struct {
	Id             string              `json:"id"`
	Approver       string              `json:"approver"`
	ApprovalStatus data.ApprovalStatus `json:"approvalStatus"`
	Requester      string              `json:"requester"`
	CreateTime     time.Time           `json:"createTime"`
	UpdateTime     time.Time           `json:"updateTime"`
}

type ShareRequestInput struct {
	Approver       *string              `json:"approver"`
	ApprovalStatus *data.ApprovalStatus `json:"approvalStatus"`
}
