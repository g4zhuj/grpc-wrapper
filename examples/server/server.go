package main

import (
	"context"
	"fmt"
	"io"
	"time"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/g4zhuj/grpc-wrapper/config"
	"github.com/g4zhuj/grpc-wrapper/plugins"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	grpcmd "github.com/grpc-ecosystem/go-grpc-middleware"
)

const (
	serviceName = "HelloServer"
)

type grpcserver struct{}

// SayHello implements helloworld.GreeterServer
func (s *grpcserver) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	grpclog.Infof("receive req : %v \n", *in)

	//start a new span, eg.(mysql)
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		n := rand.Intn(100)
		pctx := parent.Context()
		if tracer := opentracing.GlobalTracer(); tracer != nil {
			mysqlSpan := tracer.StartSpan("FindUserTable", opentracing.ChildOf(pctx))

			//do mysql operations
			time.Sleep(time.Millisecond * time.Duration(n))

			defer mysqlSpan.Finish()
		}
	}

	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

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

	//set zap logger
	logcfg := config.LoggerConfig{}
	grpclog.SetLoggerV2(logcfg.NewLogger())

	//service register
	cfg := config.RegistryConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}
	reg, err := cfg.NewRegisty()
	if err != nil {
		fmt.Printf("new registry err %v \n", err)
		return
	}

	var servOpts []grpc.ServerOption

	//open tracing
	tracer, _, err := NewJaegerTracer(serviceName)
	if err != nil {
		grpclog.Errorf("new tracer err %v , continue", err)
	}

	//open falcon
	falconReporter := plugins.NewDefaultFalconReporter()

	chainInter := grpcmd.ChainUnaryServer(
		plugins.OpentracingServerInterceptor(tracer),
		plugins.MetricServerInterceptor(falconReporter),
	)
	servOpts = append(servOpts, grpc.UnaryInterceptor(chainInter))

	servConf := config.ServiceConfig{
		ServiceName:       serviceName,
		Binding:           ":1234",
		AdvertisedAddress: "127.0.0.1:1234",
	}
	servWrapper := servConf.NewServer(reg, servOpts...)
	if coreServ := servWrapper.GetGRPCServer(); coreServ != nil {
		pb.RegisterGreeterServer(coreServ, &grpcserver{})
		servWrapper.Start()
	}
}
