package server

import (
	"context"
	"log"
	"net"

	wrapper "github.com/g4zhuj/grpc-wrapper"

	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/reflection"
)

//Server wrapper of grpc server
type ServerWrapper struct {
	s        *grpc.Server
	sopts    ServOption
	registry wrapper.Registry
}

func NewServerWrapper(opts ...ServOptions) *ServerWrapper {
	var servWrapper ServerWrapper
	for _, opt := range opts {
		opt(&servWrapper.sopts)
	}
	return &servWrapper
}

func (sw *ServerWrapper) GetGRPCServer() *grpc.Server {
	return sw.s
}

//Start start running server
func (sw *ServerWrapper) Start() {
	lis, err := net.Listen("tcp", sw.sopts.binding)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	sw.s = grpc.NewServer()
	// Register reflection service on gRPC server.
	reflection.Register(sw.s)
	if err := sw.s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	//registry
	if sw.registry != nil {
		sw.registry.Register(context.TODO(), sw.sopts.serviceName,
			naming.Update{Op: naming.Add, Addr: sw.sopts.advertisedAddress, Metadata: "..."})
	}
}

//Stop stop tht server
func (sw *ServerWrapper) Stop() {
}
