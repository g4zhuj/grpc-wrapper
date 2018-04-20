package config

import (
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
	regCfg *RegistryConfig `yaml:"registry_config"`
	logCfg *LoggerConfig   `yaml:"logger_config"`
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

//
func convertLogLevel(levelStr string) (level zapcore.Level) {
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

//NewDefaultLoggerConfig create a default config
func NewDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      "debug",
		Filename:   "./logs",
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 10,
	}
}

// 日志时间格式
// func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
// 	enc.AppendString(t.Format("2006-01-02 15:04:05"))
// }
//NewLogger create logger by config
func (lconf *LoggerConfig) NewLogger() *plugins.ZapLogger {
	if lconf.Filename == "" {
		logger, _ := zap.NewProduction(zap.AddCallerSkip(2))
		return plugins.NewZapLogger(logger)
	}

	enCfg := zap.NewProductionEncoderConfig()
	if lconf.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	encoder := zapcore.NewJSONEncoder(enCfg)
	if lconf.Encoding == "console" {
		zapcore.NewConsoleEncoder(enCfg)
	}

	//zapWriter := zapcore.
	zapWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   lconf.Filename,
		MaxSize:    lconf.MaxSize,
		MaxAge:     lconf.MaxAge,
		MaxBackups: lconf.MaxBackups,
		LocalTime:  lconf.LocalTime,
	})

	newCore := zapcore.NewCore(encoder, zapWriter, zap.NewAtomicLevelAt(convertLogLevel(lconf.Level)))
	opts := []zap.Option{zap.ErrorOutput(zapWriter)}
	opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(2))
	logger := zap.New(newCore, opts...)
	return plugins.NewZapLogger(logger)
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
