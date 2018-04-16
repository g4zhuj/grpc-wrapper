package client

import (
	"sync"

	"google.golang.org/grpc"
)

//Client client
type Client struct {
	sync.RWMutex
	connPool map[string]*grpc.ClientConn
	dialOpts []grpc.DialOption
}

//NewClient create a new client
func NewClient(opts []grpc.DialOption) *Client {
	return &Client{
		connPool: make(map[string]*grpc.ClientConn),
		dialOpts: opts,
	}
}

//GetConn get grpc conn with service name
func (cli *Client) GetConn(serviceName string) (*grpc.ClientConn, error) {
	cli.RLock()
	if conn, ok := cli.connPool[serviceName]; ok {
		cli.RUnlock()
		return conn, nil
	}
	cli.RUnlock()

	cli.Lock()
	defer cli.Unlock()

	conn, err := grpc.Dial(serviceName, cli.dialOpts...)
	if err != nil {
		return nil, err
	}
	cli.connPool[serviceName] = conn
	return conn, nil
}

//Close close specific service's client
func (cli *Client) Close(serviceName string) (err error) {
	if conn, ok := cli.connPool[serviceName]; ok {
		return conn.Close()
	}
	return nil
}
