package main

import (
	"context"
	"fmt"

	"github.com/g4zhuj/grpc-wrapper/config"
	"github.com/g4zhuj/grpc-wrapper/server"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

type grpcserver struct{}

// SayHello implements helloworld.GreeterServer
func (s *grpcserver) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("receive ctx %v, req : %v \n", ctx, *in)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	cfg := config.RegistryConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}

	serv := server.NewServerWrapper()
	pb.RegisterGreeterServer(serv.GetGRPCServer(), &grpcserver{})
	serv.Start()
}
