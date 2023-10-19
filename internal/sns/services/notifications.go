package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"philcali.me/recipes/internal/notifications"
)

type NotificationSNSService struct {
	Sns      sns.Client
	TopicArn string
}

func (n *NotificationSNSService) Subscribe(input notifications.SubscribeInput) (*notifications.SubscribeOutput, error) {
	output, err := n.Sns.Subscribe(context.TODO(), &sns.SubscribeInput{
		Endpoint:              input.Endpoint,
		Protocol:              input.Protocol,
		TopicArn:              aws.String(n.TopicArn),
		ReturnSubscriptionArn: true,
	})

	if err != nil {
		return nil, err
	}

	return &notifications.SubscribeOutput{
		SubscriberId: *output.SubscriptionArn,
	}, nil
}

func (n *NotificationSNSService) Unsubscribe(subscriberId string) error {
	_, err := n.Sns.Unsubscribe(context.TODO(), &sns.UnsubscribeInput{
		SubscriptionArn: aws.String(subscriberId),
	})

	return err
}
