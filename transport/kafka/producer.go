package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kit/kit/endpoint"
	kafka "github.com/segmentio/kafka-go"
)

// Producer wraps a URL and provides a method that implements endpoint.Endpoint.
type Producer struct {
	writer  *kafka.Writer
	topic   string
	enc     EncodeRequestFunc
	dec     DecodeResponseFunc
	before  []RequestFunc
	after   []ProducerResponseFunc
	timeout time.Duration
}

// NewProducer constructs a usable Producer for a single remote method.
func NewProducer(
	writer *kafka.Writer,
	topic string,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...ProducerOption,
) *Producer {
	p := &Producer{
		writer:  writer,
		topic:   topic,
		enc:     enc,
		dec:     dec,
		timeout: 10 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// ProducerOption sets an optional parameter for clients.
type ProducerOption func(*Producer)

// ProducerBefore sets the RequestFuncs that are applied to the outgoing NATS
// request before it's invoked.
func ProducerBefore(before ...RequestFunc) ProducerOption {
	return func(p *Producer) { p.before = append(p.before, before...) }
}

// ProducerAfter sets the ClientResponseFuncs applied to the incoming NATS
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func ProducerAfter(after ...ProducerResponseFunc) ProducerOption {
	return func(p *Producer) { p.after = append(p.after, after...) }
}

// ProducerTimeout sets the available timeout for NATS request.
func ProducerTimeout(timeout time.Duration) ProducerOption {
	return func(p *Producer) { p.timeout = timeout }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Producer) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		msg := kafka.Message{}

		if err := p.enc(ctx, &msg, request); err != nil {
			return nil, err
		}

		for _, f := range p.before {
			ctx = f(ctx, &msg)
		}

		if err := p.writer.WriteMessages(ctx, msg); err != nil {
			return nil, err
		}

		for _, f := range p.after {
			// Since we do not have a response when publishing in Kafka,
			// We call after funcs with the message we sent.
			ctx = f(ctx, &msg)
		}

		return nil, nil
	}
}

// EncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Data of the Msg. Many JSON-over-NATS services can use it as
// a sensible default.
func EncodeJSONRequest(_ context.Context, msg *kafka.Message, request interface{}) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	msg.Value = b

	return nil
}
