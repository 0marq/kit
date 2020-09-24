package awssns

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/go-kit/kit/endpoint"
)

// Publisher wraps a URL and provides a method that implements endpoint.Endpoint.
type Publisher struct {
	client   *sns.SNS
	topicArn string
	enc      EncodeNotificationFunc
	before   []RequestFunc
	after    []PublisherResponseFunc
}

// NewPublisher constructs a usable Publisher for a single remote method.
func NewPublisher(
	client *service.SNS,
	topicArn string,
	enc EncodeRequestFunc,
	options ...PublisherOption,
) *Publisher {
	p := &Publisher{
		client:   client,
		topicArn: topicArn,
		enc:      enc,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// PublisherOption sets an optional parameter for clients.
type PublisherOption func(*Publisher)

// PublisherBefore sets the RequestFuncs that are applied to the outgoing
// publish request before it's invoked.
func PublisherBefore(before ...RequestFunc) PublisherOption {
	return func(p *Publisher) { p.before = append(p.before, before...) }
}

// PublisherAfter sets the PublisherResponseFunc applied to the incoming sns
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func PublisherAfter(after ...PublisherResponseFunc) PublisherOption {
	return func(p *Publisher) { p.after = append(p.after, after...) }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Publisher) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		msg := &sns.PublishInput{TopicArn: aws.String(p.topicArn)}

		if err := p.enc(ctx, msg, request); err != nil {
			return nil, err
		}

		for _, f := range p.before {
			ctx = f(ctx, msg)
		}

		resp, err := p.client.PublishWithContext(ctx, msg)
		if err != nil {
			return nil, err
		}

		for _, f := range p.after {
			ctx = f(ctx, resp)
		}

		// sns.PublishOutput only provides the Unique ID assigned to the published message.
		return *output.MessageId, nil
	}
}

// EncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Data of the sns input. Many services can use it as a sensible default.
func EncodeJSONRequest(_ context.Context, input *sns.PublishInput, request interface{}) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	input.Message = aws.String(string(b))
	input.MessageStructure = aws.String("json")
	return nil
}
