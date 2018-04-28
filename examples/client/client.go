package main

import (
	"context"
	"io"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/g4zhuj/grpc-wrapper/config"
	"github.com/g4zhuj/grpc-wrapper/plugins"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"

	grpcmd "github.com/grpc-ecosystem/go-grpc-middleware"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	serviceName = "HelloClient"
	servName    = "HelloServer"
)

//NewJaegerTracer New Jaeger for opentracing
func NewJaegerTracer(serviceName string) (tracer opentracing.Tracer, closer io.Closer, err error) {
	cfg := jaegerCfg.Configuration{
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "192.168.1.105:6831",
		},
	}
	tracer, closer, err = cfg.New(
		serviceName,
		jaegerCfg.Logger(jaeger.StdLogger),
	)
	//defer closer.Close()

	if err != nil {
		return
	}
	opentracing.SetGlobalTracer(tracer)
	return
}

func main() {

	//service discovery
	cfg := config.RegistryConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}
	r, err := cfg.NewResolver()
	if err != nil {
		grpclog.Errorf("new registry err %v \n", err)
		return
	}
	b := grpc.RoundRobin(r)

	//set logger
	logcfg := config.LoggerConfig{
		Level: "debug",
	}
	grpclog.SetLoggerV2(logcfg.NewLogger())

	dialOpts := []grpc.DialOption{grpc.WithTimeout(time.Second * 3), grpc.WithBalancer(b), grpc.WithInsecure(), grpc.WithBlock()}

	//open tracing
	tracer, _, err := NewJaegerTracer(serviceName)
	if err != nil {
		grpclog.Errorf("new tracer err %v , continue", err)
	}

	//open falcon
	falconReporter := plugins.NewDefaultFalconReporter()

	chainInter := grpcmd.ChainUnaryClient(
		plugins.OpenTracingClientInterceptor(tracer),
		plugins.MetricClientInterceptor(falconReporter),
	)
	dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(chainInter))

	//time.Sleep(time.Second * 1)
	conn, err := grpc.Dial(servName, dialOpts...)
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

	rsp, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
	if err != nil {
		grpclog.Errorf("could not greet: %v", err)
	}
	grpclog.Infof("Greeting: %s", rsp.Message)
	time.Sleep(time.Second * 2)
}
