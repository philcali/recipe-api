package exceptions

import "fmt"

type ServiceError struct {
	StatusCode int
	Cause      error
}

func (se *ServiceError) Error() string {
	return se.Cause.Error()
}

type RequestError interface {
	ToServiceError() *ServiceError
	Error() string
}

type ConflictError struct {
	Resource string
	Id       string
}

func (ce *ConflictError) Error() string {
	return fmt.Sprintf("Found conflicting %s with id: %s", ce.Resource, ce.Id)
}

func (ce *ConflictError) ToServiceError() *ServiceError {
	return &ServiceError{
		StatusCode: 409,
		Cause:      ce,
	}
}

func Conflict(resource string, id string) *ConflictError {
	return &ConflictError{
		Resource: resource,
		Id:       id,
	}
}

type NotFoundError struct {
	Resource string
	Id       string
}

func (nfe *NotFoundError) Error() string {
	return fmt.Sprintf("Could not find a %s with id: %s", nfe.Resource, nfe.Id)
}

func (nfe *NotFoundError) ToServiceError() *ServiceError {
	return &ServiceError{
		StatusCode: 404,
		Cause:      nfe,
	}
}

func NotFound(resource string, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		Id:       id,
	}
}

type InvalidInputError struct {
	Message string
}

func (ie *InvalidInputError) Error() string {
	return ie.Message
}

func (ie *InvalidInputError) ToServiceError() *ServiceError {
	return &ServiceError{
		StatusCode: 400,
		Cause:      ie,
	}
}

func InvalidInput(message string) *InvalidInputError {
	return &InvalidInputError{
		Message: message,
	}
}
