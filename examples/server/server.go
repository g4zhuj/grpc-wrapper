package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc/grpclog"

	"github.com/g4zhuj/grpc-wrapper/config"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

type grpcserver struct{}

// SayHello implements helloworld.GreeterServer
func (s *grpcserver) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	grpclog.Infof("receive req : %v \n", *in)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	cfg := config.RegistryConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}

	//set zap logger
	logcfg := config.LoggerConfig{}
	grpclog.SetLoggerV2(logcfg.NewLogger())

	reg, err := cfg.NewRegisty()
	if err != nil {
		fmt.Printf("new registry err %v \n", err)
		return
	}

	servConf := config.ServiceConfig{
		ServiceName:       "test",
		Binding:           ":1234",
		AdvertisedAddress: "127.0.0.1:1234",
	}
	servWrapper := servConf.NewServer(reg)
	if coreServ := servWrapper.GetGRPCServer(); coreServ != nil {
		pb.RegisterGreeterServer(coreServ, &grpcserver{})
		servWrapper.Start()
	}
}
