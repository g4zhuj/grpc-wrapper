package config

//CliConfiguration config of client
type CliConfiguration struct {
}

//RegistryConfig configures the etcd cluster.
type RegistryConfig struct {
	RegistryType string   `yaml:"registry_type"` //etcd default
	HostPorts    []string `yaml:"host_ports"`
	UserName     string   `yaml:"user_name"`
	Pass         string   `yaml:"pass"`
}

//ServConfiguration config of server
type ServConfiguration struct {
}

//ServiceConfig configures the etcd cluster.
type ServiceConfig struct {
	ServiceName       string `yaml:"service_name"`
	ListenAddress     string `yaml:"listene_address"`
	AdvertisedAddress string `yaml:"advertised_address"`
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
