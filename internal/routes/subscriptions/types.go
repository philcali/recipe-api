package subscriptions

import (
	"time"

	"philcali.me/recipes/internal/data"
)

type Subscription struct {
	Endpoint   string    `json:"endpoint"`
	Protocol   string    `json:"protocol"`
	Id         string    `json:"subscriberId"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

type SubscriptionInput struct {
	Endpoint *string `json:"endpoint"`
	Protocol *string `json:"protocol"`
}

func NewSubscription(entry data.SubscriptionDTO) Subscription {
	return Subscription{
		Endpoint:   entry.Endpoint,
		Protocol:   entry.Protocol,
		Id:         entry.SK,
		CreateTime: entry.CreateTime,
		UpdateTime: entry.UpdateTime,
	}
}
