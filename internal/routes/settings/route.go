package settings

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type SettingsService struct {
	data data.SettingsRepository
}

func NewRoute(data data.SettingsRepository) routes.Service {
	return &SettingsService{
		data: data,
	}
}

func (s *SettingsService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/settings":  util.AuthorizedRoute(s.GetSettings),
		"POST:/settings": util.AuthorizedRoute(s.PutSettings),
	}
}

func _convertSettings(data data.SettingsDTO) Settings {
	return Settings{
		AutoShareLists:   data.AutoShareLists,
		AutoShareRecipes: data.AutoShareRecipes,
		CreateTime:       data.CreateTime,
		UpdateTime:       data.UpdateTime,
	}
}

func (s *SettingsService) GetSettings(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	item, err := s.data.Get(util.Username(ctx), "Global")
	if _, ok := err.(*exceptions.NotFoundError); ok {
		return util.SerializeResponseOK(_convertSettings, data.SettingsDTO{
			AutoShareLists:   false,
			AutoShareRecipes: false,
			CreateTime:       time.Now(),
			UpdateTime:       time.Now(),
		}, nil)
	}
	return util.SerializeResponseOK(_convertSettings, item, err)
}

func (s *SettingsService) PutSettings(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	updateItem := data.SettingsInputDTO{}
	if err := json.Unmarshal([]byte(event.Body), &updateItem); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	item, err := s.data.CreateWithItemId(util.Username(ctx), updateItem, "Global")
	if err == nil {
		return util.SerializeResponseOK(_convertSettings, item, nil)
	}
	if _, ok := err.(*exceptions.ConflictError); ok {
		item, err = s.data.Update(util.Username(ctx), "Global", updateItem)
		return util.SerializeResponseOK(_convertSettings, item, err)
	}
	return events.APIGatewayV2HTTPResponse{}, exceptions.InternalServer(err.Error())
}
