package data

import "time"

type SubscriptionDTO struct {
	PK            string    `dynamodbav:"PK"`
	SK            string    `dynamodbav:"SK"`
	Endpoint      string    `dynamodbav:"endpoint"`
	Protocol      string    `dynamodbav:"protocol"`
	SubscriberArn string    `dynamodbav:"subscriberArn"`
	CreateTime    time.Time `dynamodbav:"createTime"`
	UpdateTime    time.Time `dynamodbav:"updateTime"`
}

type SubscriptionInputDTO struct {
	Endpoint      *string `dynamodbav:"endpoint"`
	Protocol      *string `dynamodbav:"protocol"`
	SubscriberArn *string `dynamodbav:"subscriberArn"`
}

type SubscriptionDataService interface {
	Repository[SubscriptionDTO, SubscriptionInputDTO]
}
