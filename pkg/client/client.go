package client

import (
	"fmt"
	"sync"

	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"google.golang.org/grpc"
)

type Client struct {
	cfg Config

	mu   sync.Mutex
	conn *grpc.ClientConn
}

type Config struct {
	SdkUDS string
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
	}

}

func (c *Client) getConn() (*grpc.ClientConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		var err error
		c.conn, err = grpcserver.Connect(
			c.cfg.SdkUDS,
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(correlation.ContextUnaryClientInterceptor),
			})
		if err != nil {
			return nil, fmt.Errorf("Failed to connect to SDK unix domain socket %s: %v", c.cfg.SdkUDS, err)
		}
	}

	return c.conn, nil
}
