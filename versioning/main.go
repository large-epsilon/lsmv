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

	conn, err := grpc.Dial(v.objectstoreAddress, grpc.WithInsecure())
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
	response := &pb.PullCommitResponse{}

	conn, err := grpc.Dial(v.objectstoreAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf(
			"Failed to dial objectstore server at %s: %v",
			v.objectstoreAddress, err)
		return nil, err
	}
	defer conn.Close()

	objectstoreClient := objectstore_pb.NewObjectStoreClient(conn)

	getObjectResponse, err := objectstoreClient.GetObject(
		ctx, &objectstore_pb.GetObjectRequest{Hash: request.Hash})
	if err != nil {
		log.Printf("Failed to get commit from objectstore: %v", err)
		return nil, err
	}

	switch x := getObjectResponse.ReturnedObject.(type) {
	case *objectstore_pb.GetObjectResponse_Commit:
		response.Commit = x.Commit
	default:
		return nil, fmt.Errorf(
			"incorrect type for object '%s': %T, expected commit",
			request.Hash, getObjectResponse.ReturnedObject)
	}

	getObjectResponse, err = objectstoreClient.GetObject(
		ctx, &objectstore_pb.GetObjectRequest{Hash: response.Commit.Tree})
	if err != nil {
		log.Printf("Failed to get tree from objectstore: %v", err)
		return nil, err
	}

	switch x := getObjectResponse.ReturnedObject.(type) {
	case *objectstore_pb.GetObjectResponse_Tree:
		response.Root = x.Tree
	default:
		return nil, fmt.Errorf(
			"incorrect type for object '%s': %T, expected tree",
			request.Hash, getObjectResponse.ReturnedObject)
	}

	return response, nil
}

func main() {
	var port = flag.Int(
		"port",
		7886,
		"Port for the objectstore server to listen on.",
	)
	var host = flag.String("host", "localhost", "Hostname for this server.")
	var objectstoreAddress = flag.String(
		"objectstore_address",
		"localhost:7887",
		"host:port address of the objectstore service to use.",
	)

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterVersioningServer(grpcServer, &VersioningServerImpl{
		objectstoreAddress: *objectstoreAddress,
	})
	log.Printf("Starting versioning server on %s:%d.", *host, *port)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
