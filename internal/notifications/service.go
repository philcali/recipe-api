package notifications

type SubscribeInput struct {
	Endpoint *string
	Protocol *string
}

type SubscribeOutput struct {
	SubscriberId string
}

type NotificationService interface {
	Subscribe(input SubscribeInput) (*SubscribeOutput, error)
	Unsubscribe(subscriberId string) error
}
