package kafka

import (
	"context"

	kafka "github.com/segmentio/kafka-go"
)

// DecodeRequestFunc extracts a user-domain message object from a producer
// message object. It's designed to be used in Kafka consumers, for consumer-side
// endpoints. One straightforward DecodeRequestFunc could be something that
// JSON decodes from the message body to the concrete response type.
type DecodeRequestFunc func(context.Context, kafka.Message) (message interface{}, err error)

// EncodeRequestFunc encodes the passed message object into the Kafka message
// object. It's designed to be used in Kafka producers, for producer-side
// endpoints. One straightforward EncodeRequestFunc could something that JSON
// encodes the object directly to the message payload.
type EncodeRequestFunc func(context.Context, *kafka.Message, interface{}) error

// EncodeResponseFunc encodes the passed response object to the consumer reply.
// It's designed to be used in Kafka consumers, for consumer-side
// endpoints. One straightforward EncodeResponseFunc could be something that
// JSON encodes the object directly to the response body.
type EncodeResponseFunc func(context.Context, *kafka.Writer, interface{}) error

// DecodeResponseFunc extracts a user-domain response object from an Kafka
// response object. It's designed to be used in Kafka producer, for producer-side
// endpoints. One straightforward DecodeResponseFunc could be something that
// JSON decodes from the response payload to the concrete response type.
type DecodeResponseFunc func(context.Context, kafka.Message) (response interface{}, err error)
