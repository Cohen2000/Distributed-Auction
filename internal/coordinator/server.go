package coordinator

import (
	"context"
	"log"

	pb "distributed-auction/proto"
)

// Server implements centralized mutual exclusion using a channel-based lock.
type Server struct {
	pb.UnimplementedCoordinatorServiceServer
	lockChan chan struct{}
}

func NewServer() *Server {
	s := &Server{
		lockChan: make(chan struct{}, 1),
	}
	s.lockChan <- struct{}{} // Lock starts as available
	return s
}

func (s *Server) Acquire(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	log.Printf("[coordinator] %s requesting lock", req.NodeId)

	select {
	case <-s.lockChan:
		log.Printf("[coordinator] %s acquired lock", req.NodeId)
		return &pb.LockResponse{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Server) Release(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	s.lockChan <- struct{}{}
	log.Printf("[coordinator] %s released lock", req.NodeId)
	return &pb.LockResponse{}, nil
}
