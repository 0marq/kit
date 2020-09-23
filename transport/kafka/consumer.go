package kafka

import (
	"context"
	"encoding/json"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kafka "github.com/segmentio/kafka-go"
)

// Consumer wraps an endpoint and provides message handler.
type Consumer struct {
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	enc          EncodeResponseFunc
	before       []RequestFunc
	after        []ConsumerResponseFunc
	errorEncoder ErrorEncoder
	finalizer    []ConsumerFinalizerFunc
	errorHandler transport.ErrorHandler
}

// NewConsumer constructs a new consumer, which provides message handler and wraps
// the provided endpoint.
func NewConsumer(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	options ...ConsumerOption,
) *Consumer {
	s := &Consumer{
		e:            e,
		dec:          dec,
		enc:          enc,
		errorEncoder: DefaultErrorEncoder,
		errorHandler: transport.NewLogErrorHandler(log.NewNopLogger()),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ConsumerOption sets an optional parameter for consumers.
type ConsumerOption func(*Consumer)

// ConsumerBefore functions are executed on the producer request object before the
// request is decoded.
func ConsumerBefore(before ...RequestFunc) ConsumerOption {
	return func(c *Consumer) { c.before = append(c.before, before...) }
}

// ConsumerAfter functions are executed on the consumer reply after the
// endpoint is invoked, but before anything is published to the reply.
func ConsumerAfter(after ...ConsumerResponseFunc) ConsumerOption {
	return func(c *Consumer) { c.after = append(c.after, after...) }
}

// ConsumerErrorEncoder is used to encode errors to the consumer reply
// whenever they're encountered in the processing of a request. Clients can
// use this to provide custom error formatting. By default,
// errors will be published with the DefaultErrorEncoder.
func ConsumerErrorEncoder(ee ErrorEncoder) ConsumerOption {
	return func(c *Consumer) { c.errorEncoder = ee }
}

// ConsumerErrorLogger is used to log non-terminal errors. By default, no errors
// are logged. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ConsumerErrorEncoder which has access to the context.
// Deprecated: Use ConsumerErrorHandler instead.
func ConsumerErrorLogger(logger log.Logger) ConsumerOption {
	return func(c *Consumer) { c.errorHandler = transport.NewLogErrorHandler(logger) }
}

// ConsumerErrorHandler is used to handle non-terminal errors. By default, non-terminal errors
// are ignored. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ConsumerErrorEncoder which has access to the context.
func ConsumerErrorHandler(errorHandler transport.ErrorHandler) ConsumerOption {
	return func(c *Consumer) { c.errorHandler = errorHandler }
}

// ConsumerFinalizer is executed at the end of every request from a producer through Kafka.
// By default, no finalizer is registered.
func ConsumerFinalizer(f ...ConsumerFinalizerFunc) ConsumerOption {
	return func(c *Consumer) { c.finalizer = f }
}

// HandleMsg calls handlers for received Kafka message
func (c Consumer) HandleMsg(msg kafka.Message, kw *kafka.Writer) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(c.finalizer) > 0 {
		defer func() {
			for _, f := range c.finalizer {
				f(ctx, &msg)
			}
		}()
	}

	for _, f := range c.before {
		ctx = f(ctx, &msg)
	}

	request, err := c.dec(ctx, msg)
	if err != nil {
		c.errorHandler.Handle(ctx, err)
		c.errorEncoder(ctx, err, kw)
		return
	}

	response, err := c.e(ctx, request)
	if err != nil {
		c.errorHandler.Handle(ctx, err)
		c.errorEncoder(ctx, err, kw)
		return
	}

	for _, f := range c.after {
		ctx = f(ctx, kw)
	}

	if err := c.enc(ctx, kw, response); err != nil {
		c.errorHandler.Handle(ctx, err)
		c.errorEncoder(ctx, err, kw)
		return
	}
}

// Start starts reading messages from reader and call c.HandleMsg for every received msg
func (c Consumer) Start(ctx context.Context, kr *kafka.Reader, kw *kafka.Writer) {
	for {
		msg, err := kr.ReadMessage(ctx)
		if err != nil {
			c.errorHandler.Handle(ctx, err)
		}
		c.HandleMsg(msg, kw)
	}
}

// ErrorEncoder is responsible for encoding an error to the consumer reply.
// Users are encouraged to use custom ErrorEncoders to encode errors to
// their replies, and will likely want to pass and check for their own error
// types.
type ErrorEncoder func(ctx context.Context, err error, kw *kafka.Writer)

// ConsumerFinalizerFunc can be used to perform work at the end of an request
// from a producer, after the response has been written to the producer. The principal
// intended use is for request logging.
type ConsumerFinalizerFunc func(ctx context.Context, msg *kafka.Message)

// NopRequestDecoder is a DecodeRequestFunc that can be used for requests that do not
// need to be decoded, and simply returns nil, nil.
func NopRequestDecoder(_ context.Context, _ kafka.Message) (interface{}, error) {
	return nil, nil
}

// EncodeJSONResponse is a EncodeResponseFunc that serializes the response as a
// JSON object to the consumer reply. Many JSON-over services can use it as
// a sensible default.
func EncodeJSONResponse(ctx context.Context, reply string, kw *kafka.Writer, response interface{}) error {
	b, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return kw.WriteMessages(ctx, kafka.Message{
		Value: b,
	})
}

// DefaultErrorEncoder writes the error to the consumer reply.
func DefaultErrorEncoder(ctx context.Context, err error, kw *kafka.Writer) {
	logger := log.NewNopLogger()

	type Response struct {
		Error string `json:"err"`
	}

	var response Response

	response.Error = err.Error()

	b, err := json.Marshal(response)
	if err != nil {
		logger.Log("err", err)
		return
	}
	if err := kw.WriteMessages(ctx, kafka.Message{
		Value: b,
	}); err != nil {
		logger.Log("err", err)
	}
}
