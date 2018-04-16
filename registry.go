package wrapper

import (
	"context"
	"time"

	"google.golang.org/grpc/naming"
)

type RegistryOption struct {
	TTL time.Duration
}
type RegistryOptions func(o *RegistryOption)

func WithTTL(ttl time.Duration) RegistryOptions {
	return func(o *RegistryOption) {
		o.TTL = ttl
	}
}

//Registry registry
type Registry interface {
	Register(ctx context.Context, target string, update naming.Update, opts ...RegistryOptions) (err error)
	Close()
}
