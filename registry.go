package wrapper

import (
	"context"
	"time"

	"google.golang.org/grpc/naming"
)

//DefaultRegInfTTL default ttl of server info in registry
const DefaultRegInfTTL = time.Second * 50

type RegistryOption struct {
	TTL time.Duration
}
type RegistryOptions func(o *RegistryOption)

//WithTTL set ttl
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
