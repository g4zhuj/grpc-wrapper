package config

import (
	"sync/atomic"
	"unsafe"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/g4zhuj/grpc-wrapper/plugins"
	"github.com/g4zhuj/grpc-wrapper/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"gopkg.in/natefinch/lumberjack.v2"

	wrapper "github.com/g4zhuj/grpc-wrapper"
)

//CliConfiguration config of client
type CliConfiguration struct {
}

//RegistryConfig configures the etcd cluster.
type RegistryConfig struct {
	RegistryType string   `yaml:"registry_type"` //etcd default
	Endpoints    []string `yaml:"endpoints"`
	UserName     string   `yaml:"user_name"`
	Pass         string   `yaml:"pass"`
}

//ServConfiguration config of server
type ServConfiguration struct {
}

//ServiceConfig configures the etcd cluster.
type ServiceConfig struct {
	ServiceName       string `yaml:"service_name"`
	Binding           string `yaml:"binding"`
	AdvertisedAddress string `yaml:"advertised_address"`
	RegistryTTL       int    `yaml:"registry_ttl"`
}

//TokenConfig config of token, default ttl:1 day, default token length 32 bytes.
type TokenConfig struct {
	StaticToken string `yaml:"static_token"`
	TokenTTL    string `yaml:"token_ttl"`
	TokenLength int    `yaml:"token_length"`
}

//LoggerConfig config of logger
type LoggerConfig struct {
	Level      string `yaml:"level"`       //debug  info  warn  error
	Encoding   string `yaml:"encoding"`    //json or console
	CallFull   bool   `yaml:"call_full"`   //whether full call path or short path, default is short
	Filename   string `yaml:"file_name"`   //log file name
	MaxSize    int    `yaml:"max_size"`    //max size of log.(MB)
	MaxAge     int    `yaml:"max_age"`     //time to keep, (day)
	MaxBackups int    `yaml:"max_backups"` //max file numbers
	LocalTime  bool   `yaml:"local_time"`  //(default UTC)
	Compress   bool   `yaml:"compress"`    //default false
}

//OpenTracingConfig support jaeger and zipkin
type OpenTracingConfig struct {
}

// 设置日志最低等级
func setLogConsole(levelStr string) (level zapcore.Level, err error) {
	switch levelStr {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}
	return
}

func (lconf *LoggerConfig) NewLogger() {
	proConf := zapcore.EncoderConfig{
		TimeKey:        config.TimeKey,
		LevelKey:       config.LevelKey,
		NameKey:        "logger",
		CallerKey:      config.CallerKey,
		MessageKey:     config.MessageKey,
		StacktraceKey:  config.TraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	enCfg := zap.NewProductionEncoderConfig()
	if lconf.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	encoder := zapcore.NewJSONEncoder(enCfg)
	if lconf.Encoding == "console" {
		zapcore.NewConsoleEncoder(enCfg)
	}

	if config.CallFull {
		proConf.EncodeCaller = zapcore.FullCallerEncoder
	} else {
		proConf.EncodeCaller = zapcore.ShortCallerEncoder
	}

	zapWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   lconf.Filename,
		MaxSize:    lconf.MaxSize,
		MaxAge:     lconf.MaxAge,
		MaxBackups: lconf.MaxBackups,
		LocalTime:  lconf.LocalTime,
	})
	// var writer zapcore.WriteSyncer
	// writers := []zapcore.WriteSyncer{}
	// writers = append(writers, zapWriter)
	// output := zapcore.NewMultiWriteSyncer(writers...)

	//设置日志等级
	newCore := zapcore.NewCore(encoder, zapWriter, zap.NewAtomicLevelAt(lconf.Level))
	opts := []zap.Option{zap.ErrorOutput(writer)}
	opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	if config.TraceOpen {
		opts = append(opts, zap.AddStacktrace(config.TraceLevel))
	}
	zaploger := zap.New(newCore, opts...)
	if ZapLogger == nil {
		ZapLogger = zaploger
	} else {
		oldLoger := ZapLogger
		atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&ZapLogger)), unsafe.Pointer(zaploger))
		oldLoger.Sync()
	}

}

//NewServer new server wrapper with config
func (servconf *ServiceConfig) NewServer(registry wrapper.Registry, opts ...grpc.ServerOption) *server.ServerWrapper {
	var servOpts []grpc.ServerOption
	servOpts = append(servOpts, opts...)
	serv := server.NewServerWrapper(
		server.WithAdvertisedAddress(servconf.AdvertisedAddress),
		server.WithBinding(servconf.Binding),
		server.WithRegistry(registry),
		server.WithServiceName(servconf.ServiceName),
		server.WithGRPCServOption(servOpts),
	)
	return serv
}

//NewResolver create a resolver for grpc
func (regconf *RegistryConfig) NewResolver() (naming.Resolver, error) {
	cli, err := etcd.New(etcd.Config{
		Endpoints: regconf.Endpoints,
		Username:  regconf.UserName,
		Password:  regconf.Pass,
	})
	if err != nil {
		return nil, err
	}
	return plugins.NewEtcdResolver(cli), nil
}

//NewRegisty create a reistry for registering server addr
func (regconf *RegistryConfig) NewRegisty() (wrapper.Registry, error) {

	cli, err := etcd.New(etcd.Config{
		Endpoints: regconf.Endpoints,
		Username:  regconf.UserName,
		Password:  regconf.Pass,
	})
	if err != nil {
		return nil, err
	}
	return plugins.NewEtcdRegisty(cli), nil
}
