package events

import "github.com/aws/aws-lambda-go/events"

type EventFilter interface {
	Filter(record events.DynamoDBEventRecord) bool
	Apply(record events.DynamoDBEventRecord) error
}
