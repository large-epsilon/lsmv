package in_memory_store

import (
	"context"
	"fmt"

	data_pb "lsmv/proto/data"
	pb "lsmv/proto/objectstore"
)

type InMemoryObjectStoreServer struct {
	blobs   map[string]*data_pb.Blob
	trees   map[string]*data_pb.Tree
	commits map[string]*data_pb.Commit
}

func New() *InMemoryObjectStoreServer {
    return &InMemoryObjectStoreServer{
		blobs:   map[string]*data_pb.Blob{},
		trees:   map[string]*data_pb.Tree{},
		commits: map[string]*data_pb.Commit{},
	}
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
	if resp == nil {
		return nil, fmt.Errorf("No object was found for hash '%s'", request.Hash)
	}
	if len(resp.Responses) != 1 {
		return nil, fmt.Errorf("Incorrect number of responses for hash %s: got %d", request.Hash, len(resp.Responses))
	}
	return resp.Responses[0], nil
}

func (s *InMemoryObjectStoreServer) BatchedGetObject(ctx context.Context, request *pb.BatchedGetObjectRequest) (*pb.BatchedGetObjectResponse, error) {
	responses := []*pb.GetObjectResponse{}
	for _, req := range request.Requests {
		tree, ok := s.trees[req.Hash]
		if ok {
			responses = append(responses, &pb.GetObjectResponse{ReturnedObject: &pb.GetObjectResponse_Tree{tree}})
		}
		blob, ok := s.blobs[req.Hash]
		if ok {
			responses = append(responses, &pb.GetObjectResponse{ReturnedObject: &pb.GetObjectResponse_Blob{blob}})
		}
		commit, ok := s.commits[req.Hash]
		if ok {
			responses = append(responses, &pb.GetObjectResponse{ReturnedObject: &pb.GetObjectResponse_Commit{commit}})
		}
	}
	return &pb.BatchedGetObjectResponse{Responses: responses}, nil
}
