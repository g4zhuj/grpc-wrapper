# grpc-wrapper

详细实现文档见 docs

grpc的封装扩展,集成通用的组件,形成一个微服务通讯框架.

## 1.支持的扩展

* 服务注册与发现</br>
etcd [OK] 

* 结构化日志 </br>
zap [OK] 

* 服务调用链追踪</br>
zipkin [OK]
jaeger [OK]

* 服务指标监控</br>
falcon-plus [TODO] 


## 2.使用方式
#### client端
```go
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

	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	serviceName = "HelloClient"
	servName    = "HelloServer"
)

//NewJaegerTracer New Jaeger for opentracing
func NewJaegerTracer(serviceName string) (tracer opentracing.Tracer, 
closer io.Closer, err error) {
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
	if tracer != nil {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(plugins.OpenTracingClientInterceptor(tracer)))
	}

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


```
#### server端

```go
package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/g4zhuj/grpc-wrapper/config"
	"github.com/g4zhuj/grpc-wrapper/plugins"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
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
		pctx := parent.Context()
		if tracer := opentracing.GlobalTracer(); tracer != nil {
			mysqlSpan := tracer.StartSpan("FindUserTable", opentracing.ChildOf(pctx))

			//do mysql operations
			time.Sleep(time.Millisecond * 100)

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
	if tracer != nil {
		servOpts = append(servOpts, grpc.UnaryInterceptor(plugins.OpentracingServerInterceptor(tracer)))
	}

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

```


## 实例
使用示例见examples
