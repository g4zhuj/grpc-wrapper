package wrapper

import (
	"github.com/opentracing/opentracing-go/ext"

	opentracing "github.com/opentracing/opentracing-go"
)

var (
	//TracingComponentTag tags
	TracingComponentTag = opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}
)
