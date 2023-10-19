package subscriptions

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/notifications"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type SubscriptionService struct {
	data          data.SubscriptionDataService
	notifications notifications.NotificationService
}

func NewRoute(data data.SubscriptionDataService, notifications notifications.NotificationService) routes.Service {
	return &SubscriptionService{
		data:          data,
		notifications: notifications,
	}
}

func (s *SubscriptionService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/subscriptions":                  util.AuthorizedRoute(s.ListSubscriptions),
		"GET:/subscriptions/:subscriberId":    util.AuthorizedRoute(s.GetSubscription),
		"POST:/subscriptions":                 util.AuthorizedRoute(s.CreateSubscription),
		"DELETE:/subscriptions/:subscriberId": util.AuthorizedRoute(s.DeleteSubscription),
	}
}

func (s *SubscriptionService) ListSubscriptions(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	return util.SerializeList[data.SubscriptionDTO, data.SubscriptionInputDTO](s.data, NewSubscription, event, ctx)
}

func (s *SubscriptionService) GetSubscription(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	item, err := s.data.Get(util.Username(ctx), util.RequestParam(ctx, "subscriberId"))
	return util.SerializeResponseOK(NewSubscription, item, err)
}

func (s *SubscriptionService) CreateSubscription(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := SubscriptionInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}

	subscription, err := s.notifications.Subscribe(notifications.SubscribeInput{
		Endpoint: input.Endpoint,
		Protocol: input.Protocol,
	})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InternalServer(err.Error())
	}

	created, err := s.data.Create(util.Username(ctx), data.SubscriptionInputDTO{
		Endpoint:      input.Endpoint,
		Protocol:      input.Protocol,
		SubscriberArn: &subscription.SubscriberId,
	})
	return util.SerializeResponseOK(NewSubscription, created, err)
}

func (s *SubscriptionService) DeleteSubscription(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	subscriber, err := s.data.Get(util.Username(ctx), util.RequestParam(ctx, "subscriberId"))
	if err != nil {
		_, ok := err.(*exceptions.NotFoundError)
		if ok {
			return util.SerializeResponseNoContent(nil)
		} else {
			return events.APIGatewayV2HTTPResponse{}, exceptions.InternalServer(err.Error())
		}
	}

	err = s.notifications.Unsubscribe(subscriber.SubscriberArn)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InternalServer(err.Error())
	}

	return util.SerializeResponseNoContent(s.data.Delete(util.Username(ctx), subscriber.SK))
}
