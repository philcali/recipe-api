package subscriptions

import (
	"philcali.me/recipes/internal/data"
)

type Subscription struct {
	Endpoint string `json:"endpoint"`
	Protocol string `json:"protocol"`
	Id       string `json:"subscriberId"`
}

type SubscriptionInput struct {
	Endpoint *string `json:"endpoint"`
	Protocol *string `json:"protocol"`
}

func (s *SubscriptionInput) toData() data.SubscriptionInputDTO {
	return data.SubscriptionInputDTO{
		Endpoint: s.Endpoint,
		Protocol: s.Protocol,
	}
}

func NewSubscription(entry data.SubscriptionDTO) Subscription {
	return Subscription{
		Endpoint: entry.Endpoint,
		Protocol: entry.Protocol,
		Id:       entry.SK,
	}
}
