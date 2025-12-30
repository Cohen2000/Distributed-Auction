package coordinator

import (
	"context"

	pb "distributed-auction/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LockService interface for dependency injection.
type LockService interface {
	Acquire(ctx context.Context, nodeID string) error
	Release(ctx context.Context, nodeID string) error
}

// Client implements LockService using gRPC.
type Client struct {
	client pb.CoordinatorServiceClient
	conn   *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: pb.NewCoordinatorServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) Acquire(ctx context.Context, nodeID string) error {
	_, err := c.client.Acquire(ctx, &pb.LockRequest{NodeId: nodeID})
	return err
}

func (c *Client) Release(ctx context.Context, nodeID string) error {
	_, err := c.client.Release(ctx, &pb.LockRequest{NodeId: nodeID})
	return err
}

func (c *Client) Close() error {
	return c.conn.Close()
}
