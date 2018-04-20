package main

import (
	"context"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/g4zhuj/grpc-wrapper/config"

	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {

	cfg := config.RegistryConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}

	r, err := cfg.NewResolver()
	if err != nil {
		grpclog.Errorf("new registry err %v \n", err)
		return
	}

	//set logger
	logcfg := config.LoggerConfig{
		Level: "debug",

		// Filename: "./logs",
		// MaxSize:    1,
		// MaxAge:     1,
		// MaxBackups: 10,
	}
	grpclog.SetLoggerV2(logcfg.NewLogger())

	b := grpc.RoundRobin(r)
	//time.Sleep(time.Second * 1)
	conn, err := grpc.Dial("test", grpc.WithTimeout(time.Second*3), grpc.WithBalancer(b), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		grpclog.Errorf("Dial err %v\n", err)
		return
	}

	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := "defaultName"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	for i := 0; i < 20; i++ {
		rsp, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
		if err != nil {
			grpclog.Errorf("could not greet: %v", err)
		}
		grpclog.Infof("Greeting: %s", rsp.Message)
		time.Sleep(time.Second * 5)
	}
}
