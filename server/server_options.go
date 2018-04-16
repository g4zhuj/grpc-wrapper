package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

//ServOption
type ServOption struct {
	serviceName       string
	binding           string
	advertisedAddress string
	resolver          naming.Resolver
	grpcOpts          []grpc.ServerOption
}

type ServOptions func(o *ServOption)



func With

func WithResolver(r naming.Resolver) ServOptions {
	return func(o *ServOption) {
		o.resolver = r
	}
}

func WithGRPCServOption(opts []grpc.ServerOption) ServOptions {
	return func(o *ServOption) {
		o.grpcOpts = opts
	}
}


