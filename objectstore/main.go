package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "lsmv/proto/objectstore"
)

type InMemoryObjectStoreServer struct {
	blobs   map[string]*pb.Blob
	trees   map[string]*pb.Tree
	commits map[string]*pb.Commit
}

func (s *InMemoryObjectStoreServer) StoreObject(ctx context.Context, request *pb.StoreObjectRequest) (*pb.StoreObjectResponse, error) {
	_, err := s.BatchedStoreObject(ctx, &pb.BatchedStoreObjectRequest{Objects: []*pb.StoreObjectRequest{request}})
	return &pb.StoreObjectResponse{}, err
}

func (s *InMemoryObjectStoreServer) BatchedStoreObject(ctx context.Context, request *pb.BatchedStoreObjectRequest) (*pb.BatchedStoreObjectResponse, error) {
	for _, single := range request.Objects {
		switch x := (*single).ToStore.(type) {
		case *pb.StoreObjectRequest_Tree:
			s.trees[x.Tree.Hash] = x.Tree
		case *pb.StoreObjectRequest_Blob:
			s.blobs[x.Blob.Hash] = x.Blob
		case *pb.StoreObjectRequest_Commit:
			s.commits[x.Commit.Hash] = x.Commit
		}
	}
	return &pb.BatchedStoreObjectResponse{}, nil
}

func (s *InMemoryObjectStoreServer) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	resp, err := s.BatchedGetObject(ctx, &pb.BatchedGetObjectRequest{Requests: []*pb.GetObjectRequest{request}})
	if err != nil {
		return nil, err
	}
	if len(resp.Responses) != 1 {
		return nil, fmt.Errorf("Incorrect number of responses for a single hash: got %d", len(resp.Responses))
	}
	return resp.Responses[0], nil
}

func (s *InMemoryObjectStoreServer) BatchedGetObject(ctx context.Context, request *pb.BatchedGetObjectRequest) (*pb.BatchedGetObjectResponse, error) {
	var response *pb.BatchedGetObjectResponse
	for _, req := range request.Requests {
		tree, ok := s.trees[req.Hash]
		if ok {
			response.Responses = append(response.Responses, &pb.GetObjectResponse{ReturnedObject: &pb.GetObjectResponse_Tree{tree}})
		}
		blob, ok := s.blobs[req.Hash]
		if ok {
			response.Responses = append(response.Responses, &pb.GetObjectResponse{ReturnedObject: &pb.GetObjectResponse_Blob{blob}})
		}
		commit, ok := s.commits[req.Hash]
		if ok {
			response.Responses = append(response.Responses, &pb.GetObjectResponse{ReturnedObject: &pb.GetObjectResponse_Commit{commit}})
		}
	}
	return response, nil
}

func main() {
	var port = flag.Int("port", 7887, "Port for the objectstore server to listen on.")
	var host = flag.String("host", "localhost", "Hostname for this server.")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *port, *host))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterObjectStoreServer(grpcServer, &InMemoryObjectStoreServer{})
	log.Printf("Starting object store server on %s:%d.", *port, *host)
	grpcServer.Serve(lis)
}
