package awssns

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sns"
)

// RequestFunc may take information from a publisher request and put it into a
// request context. In Publishers, RequestFuncs are executed prior to invoking the
// publishing the notification.
type RequestFunc func(context.Context, *sns.PublishInput) context.Context

// PublisherResponseFunc may take information from an sns publish response and make the
// it available. PublisherResponseFunc are executed in publishers, after a request has
// been made, but prior to it being decoded.
type PublisherResponseFunc func(context.Context, *sns.PublishOutput) context.Context
