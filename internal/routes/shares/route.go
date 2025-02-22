package shares

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/exceptions"
	"philcali.me/recipes/internal/routes"
	"philcali.me/recipes/internal/routes/util"
)

type ShareRequestService struct {
	data      data.ShareRequestRepository
	indexName string
}

func NewRouteWithIndex(data data.ShareRequestRepository, indexName string) routes.Service {
	return &ShareRequestService{
		data:      data,
		indexName: indexName,
	}
}

func NewRoute(data data.ShareRequestRepository) routes.Service {
	return NewRouteWithIndex(data, os.Getenv("INDEX_NAME_1"))
}

func _convertShare(shareDTO data.ShareRequestDTO) ShareRequest {
	return ShareRequest{
		Id:             shareDTO.SK,
		Approver:       shareDTO.Approver,
		Requester:      shareDTO.Requester,
		CreateTime:     shareDTO.CreateTime,
		UpdateTime:     shareDTO.UpdateTime,
		ApprovalStatus: shareDTO.ApprovalStatus,
	}
}

func (s *ShareRequestService) GetRoutes() map[string]routes.Route {
	return map[string]routes.Route{
		"GET:/shares":             util.AuthorizedRoute(s.ListShares),
		"POST:/shares":            util.AuthorizedRoute(s.CreateShare),
		"PUT:/shares/:shareId":    util.AuthorizedRoute(s.UpdateShare),
		"DELETE:/shares/:shareId": util.AuthorizedRoute(s.DeleteShare),
	}
}

func (s *ShareRequestService) ListShares(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	if t, ok := event.QueryStringParameters["status"]; ok {
		if strings.EqualFold(string(data.REQUESTED), t) {
			claims := util.AuthorizationClaims(event)
			return util.SerializeListByIndexAndHash(s.data, _convertShare, s.indexName, event, claims["email"])
		}
	}
	return util.SerializeList(s.data, _convertShare, event, ctx)
}

func (s *ShareRequestService) DeleteShare(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	err := s.data.Delete(util.Username(ctx), util.RequestParam(ctx, "shareId"))
	return util.SerializeResponseNoContent(err)
}

func (s *ShareRequestService) CreateShare(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := ShareRequestInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	if strings.EqualFold(*input.Approver, util.Username(ctx)) {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput("Cannot share with yourself")
	}
	expiresIn := int(time.Now().Add(time.Hour + (24 * 7)).UnixMilli())
	claims := util.AuthorizationClaims(event)
	requested := data.REQUESTED
	created, err := s.data.Create(util.Username(ctx), data.ShareRequestInputDTO{
		Approver:       input.Approver,
		ApprovalStatus: &requested,
		ExpiresIn:      &expiresIn,
		Requester:      aws.String(claims["email"]),
	})
	return util.SerializeResponseOK(_convertShare, created, err)
}

func (s *ShareRequestService) UpdateShare(event events.APIGatewayV2HTTPRequest, ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	input := ShareRequestInput{}
	if err := json.Unmarshal([]byte(event.Body), &input); err != nil {
		return events.APIGatewayV2HTTPResponse{}, exceptions.InvalidInput(err.Error())
	}
	claims := util.AuthorizationClaims(event)
	shareId := util.RequestParam(ctx, "shareId")
	perform := true
	var nextToken *string
	for perform {
		results, err := s.data.ListByIndex(claims["email"], s.indexName, data.QueryParams{
			Limit:     100,
			NextToken: nextToken,
		})
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, exceptions.InternalServer(err.Error())
		}
		for _, item := range results.Items {
			if strings.EqualFold(shareId, item.SK) {
				parts := strings.Split(item.PK, ":")
				updated, err := s.data.Update(parts[0], item.SK, data.ShareRequestInputDTO{
					ApprovalStatus: input.ApprovalStatus,
					ApproverId:     aws.String(util.Username(ctx)),
				})
				return util.SerializeResponseOK(_convertShare, updated, err)
			}
		}
		nextToken = results.NextToken
		perform = nextToken != nil
	}
	return events.APIGatewayV2HTTPResponse{}, exceptions.NotFound("Share", shareId)
}
