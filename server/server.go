package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/reflection"
)

//Server wrapper of grpc server
type ServerWrapper struct {
	grpc.Server
	opt      ServOption
	resolver naming.Resolver
}

func NewServerWrapper() *ServerWrapper {

	return &ServerWrapper{}
}

//Start start running server
func (s *ServerWrapper) Start() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		fmt.Printf("nameing err %v\n", err)
		return
	}
	r := &etcdnaming.GRPCResolver{Client: cli}
	r.Update(context.TODO(), "my-service", naming.Update{Op: naming.Add, Addr: "127.0.0.1:50052", Metadata: "..."})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//Stop stop tht server
func (s *ServerWrapper) Stop() {
	s.resolver.Resolve
	s.Server.Stop()
}
