package server

import (
	"google.golang.org/grpc"

	wrapper "github.com/g4zhuj/grpc-wrapper"
)

//ServOption option of server
type ServOption struct {
	serviceName       string
	binding           string
	advertisedAddress string
	registry          wrapper.Registry
	grpcOpts          []grpc.ServerOption
}

type ServOptions func(o *ServOption)

//WithRegistry set registry
func WithRegistry(r wrapper.Registry) ServOptions {
	return func(o *ServOption) {
		o.registry = r
	}
}

//WithGRPCServOption set grpc options
func WithGRPCServOption(opts []grpc.ServerOption) ServOptions {
	return func(o *ServOption) {
		o.grpcOpts = opts
	}
}

//WithServiceName set service name
func WithServiceName(sn string) ServOptions {
	return func(o *ServOption) {
		o.serviceName = sn
	}
}

//WithBinding set binding
func WithBinding(binding string) ServOptions {
	return func(o *ServOption) {
		o.binding = binding
	}
}

//WithAdvertisedAddress set advertised address
func WithAdvertisedAddress(addr string) ServOptions {
	return func(o *ServOption) {
		o.advertisedAddress = addr
	}
}
