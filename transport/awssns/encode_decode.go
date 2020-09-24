package awssns

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sns"
)

// EncodeNotificationFunc encodes the passed request object into the aws service.PublishInput
// object. It's designed to be used in sns publisher, for publisher-side
// endpoints. One straightforward EncodeNotificationFunc could be something that JSON
// encodes the object directly to the PublishInput Message field with MessageStructure=json.
type EncodeNotificationFunc func(context.Context, *sns.PublishInput, interface{}) error
