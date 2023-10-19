package data

import "time"

type UserDTO struct {
	PK         string    `dynamodbav:"PK"`
	SK         string    `dynamodbav:"SK"`
	AccountId  string    `dynamodbav:"accountId"`
	CreateTime time.Time `dynamodbav:"createTime"`
	UpdateTime time.Time `dynamodbav:"updateTime"`
}

type UserInputDTO struct {
	AccountId string `dynamodbav:"accountId"`
}

type UserService interface {
	Repository[UserDTO, UserInputDTO]
}
