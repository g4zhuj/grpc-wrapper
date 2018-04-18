package config

import (
	etcd "github.com/coreos/etcd/clientv3"
	wrapper "github.com/g4zhuj/grpc-wrapper"
	"github.com/g4zhuj/grpc-wrapper/server"
	"google.golang.org/grpc"

	"github.com/g4zhuj/grpc-wrapper/plugins"
	"google.golang.org/grpc/naming"
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

//OpenTracingConfig support jaeger and zipkin
type OpenTracingConfig struct {
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
