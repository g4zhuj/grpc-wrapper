package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/reflection"
)

//Server wrapper of grpc server
type ServerWrapper struct {
	s     *grpc.Server
	sopts ServOption
	//registry wrapper.Registry
}

func NewServerWrapper(opts ...ServOptions) *ServerWrapper {
	var servWrapper ServerWrapper
	for _, opt := range opts {
		opt(&servWrapper.sopts)
	}
	servWrapper.s = grpc.NewServer()
	return &servWrapper
}

func (sw *ServerWrapper) GetGRPCServer() *grpc.Server {
	return sw.s
}

//Start start running server
func (sw *ServerWrapper) Start() error {
	lis, err := net.Listen("tcp", sw.sopts.binding)
	if err != nil {
		return err
	}

	//registry
	if sw.sopts.registry != nil {
		err := sw.sopts.registry.Register(context.TODO(), sw.sopts.serviceName,
			naming.Update{Op: naming.Add, Addr: sw.sopts.advertisedAddress, Metadata: "..."})
		if err != nil {
			return err
		}
	} else {
		fmt.Println("registry is nil")
	}

	// Register reflection service on gRPC server.
	reflection.Register(sw.s)
	if err := sw.s.Serve(lis); err != nil {
		return err
	}

	return nil
}

//Stop stop tht server
func (sw *ServerWrapper) Stop() {
}
