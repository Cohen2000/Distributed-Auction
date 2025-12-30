package main

import (
	"log"
	"net"

	"distributed-auction/internal/coordinator"
	pb "distributed-auction/proto"

	"google.golang.org/grpc"
)

// Coordinator process: exposes a small gRPC service that provides a centralized lock.
// Nodes use this lock to serialize local state updates.
func main() {
	lis, err := net.Listen("tcp", ":7000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterCoordinatorServiceServer(server, coordinator.NewServer())

	log.Println("[coordinator] listening on :7000")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
