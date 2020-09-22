package kafka

import (
	"context"

	kafka "github.com/segmentio/kafka-go"
)

// RequestFunc may take information from a producer message and put it into a
// message context. In Consumers, RequestFuncs are executed prior to invoking the
// endpoint.
type RequestFunc func(context.Context, *kafka.Message) context.Context

// ConsumerResponseFunc may take information from a message context and use it to
// manipulate a Producer. ConsumerResponseFuncs are only executed in
// subscribers, after invoking the endpoint but prior to publishing a reply.
type ConsumerResponseFunc func(context.Context, *kafka.Writer) context.Context

// ProducerResponseFunc may take information from an Kafka message and make the
// response available for consumption. ClientResponseFuncs are only executed in
// clients, after a message has been made, but prior to it being decoded.
type ProducerResponseFunc func(context.Context, *kafka.Message) context.Context
