package config

//ServOptions contorl the behavior of server
type ServOptions struct {
}

//CliOptions contorl the behavior of client
type CliOptions struct {
}

//ServOption a function sets options on ServOptions
type ServOption func(c *ServOptions)

//CliOption a function sets options on CliOptions
type CliOption func(c *CliOptions)
