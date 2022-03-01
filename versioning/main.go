package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	objectstore_pb "lsmv/proto/objectstore"
	pb "lsmv/proto/versioning"
)

type VersioningServerImpl struct {
	objectstoreAddress string
}

func (v *VersioningServerImpl) PushCommit(
	ctx context.Context,
	request *pb.PushCommitRequest,
) (*pb.PushCommitResponse, error) {
	storeRequest := objectstore_pb.BatchedStoreObjectRequest{}
	storeRequest.Objects = append(
		storeRequest.Objects,
		&objectstore_pb.StoreObjectRequest{
			ToStore: &objectstore_pb.StoreObjectRequest_Tree{request.Root},
		},
	)
	storeRequest.Objects = append(
		storeRequest.Objects,
		&objectstore_pb.StoreObjectRequest{
			ToStore: &objectstore_pb.StoreObjectRequest_Commit{request.Commit},
		},
	)
	for _, tree := range request.Subtrees {
		storeRequest.Objects = append(
			storeRequest.Objects,
			&objectstore_pb.StoreObjectRequest{
				ToStore: &objectstore_pb.StoreObjectRequest_Tree{tree},
			},
		)
	}
	for _, blob := range request.Files {
		storeRequest.Objects = append(
			storeRequest.Objects,
			&objectstore_pb.StoreObjectRequest{
				ToStore: &objectstore_pb.StoreObjectRequest_Blob{blob},
			},
		)
	}

	conn, err := grpc.Dial(v.objectstoreAddress)
	if err != nil {
		log.Printf(
			"Failed to dial objectstore server at %s: %v",
			v.objectstoreAddress, err)
		return nil, err
	}
	defer conn.Close()

	objectstoreClient := objectstore_pb.NewObjectStoreClient(conn)

	_, err = objectstoreClient.BatchedStoreObject(ctx, &storeRequest)
	if err != nil {
		log.Printf("Failed to store commit objects: %v", err)
		return nil, err
	}

	return &pb.PushCommitResponse{}, nil
}

func (v *VersioningServerImpl) PullCommit(
	ctx context.Context,
	request *pb.PullCommitRequest,
) (*pb.PullCommitResponse, error) {
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
	pb.RegisterVersioningServer(grpcServer, &VersioningServerImpl{})
	log.Printf("Starting versioning server on %s:%d.", *host, *port)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
