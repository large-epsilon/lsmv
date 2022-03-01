package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "lsmv/proto/versioning"
)

type DummyVersioningServer struct{}

func (d *DummyVersioningServer) GetHistory(
	ctx context.Context,
	request *pb.GetHistoryRequest,
) (*pb.GetHistoryResponse, error) {
	return nil, nil
}

func (d *DummyVersioningServer) Commit(
	ctx context.Context,
	request *pb.CommitRequest,
) (*pb.CommitResponse, error) {
	return nil, nil
}

func main() {
	var port = flag.Int("port", 7886, "Port for the objectstore server to listen on.")
	var host = flag.String("host", "localhost", "Hostname for this server.")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterVersioningServer(grpcServer, &DummyVersioningServer{})
	log.Printf("Starting versioning server on %s:%d.", *host, *port)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
