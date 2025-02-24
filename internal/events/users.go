package events

import (
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/dynamodb/users"
)

func isSubscription(pk string) bool {
	return strings.HasSuffix(pk, ":Subscription")
}

type ManageGlobalUserHandler struct {
	Users data.UserService
}

func (mh *ManageGlobalUserHandler) Filter(record events.DynamoDBEventRecord) bool {
	switch record.EventName {
	case "INSERT":
		fallthrough
	case "REMOVE":
		return isSubscription(record.Change.Keys["PK"].String())
	}
	return false
}

func (mh *ManageGlobalUserHandler) Apply(record events.DynamoDBEventRecord) error {
	switch record.EventName {
	case "INSERT":
		user, err := mh.Users.CreateWithItemId(users.GLOBAL_ACCOUNT, data.UserInputDTO{
			AccountId: strings.Split(record.Change.Keys["PK"].String(), ":")[0],
		}, record.Change.NewImage["endpoint"].String())
		fmt.Printf("Successfully created user %s: %s", user.SK, user.AccountId)
		return err
	case "REMOVE":
		return mh.Users.Delete(users.GLOBAL_ACCOUNT, record.Change.OldImage["endpoint"].String())
	}
	return nil
}

func DefaultUserHandler(db data.UserService) *ManageGlobalUserHandler {
	return &ManageGlobalUserHandler{
		Users: db,
	}
}
