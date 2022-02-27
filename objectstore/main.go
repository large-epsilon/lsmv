package main

import (
    "context"
    "flag"
    "log"
    "net"

    "google.golang.org/grpc"

    pb "lsmv/proto/objectstore"
)

type DummyObjectStoreServer struct {}

func (s *DummyObjectStoreServer) StoreObject(ctx context.Context, request *pb.StoreObjectRequest) (*pb.StoreObjectResponse, error) {
    return nil, nil
}


func (s *DummyObjectStoreServer) BatchedStoreObject(ctx context.Context, request *pb.BatchedStoreObjectRequest) (*pb.BatchedStoreObjectResponse, error) {
    return nil, nil
}


func (s *DummyObjectStoreServer) GetObject(ctx context.Context, request *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
    return nil, nil
}


func (s *DummyObjectStoreServer) BatchedGetObject(ctx context.Context, request *pb.BatchedGetObjectRequest) (*pb.BatchedGetObjectResponse, error) {
    return nil, nil
}

func main() {
    flag.Parse()

    lis, err := net.Listen("tcp", "localhost:7887")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    var opts []grpc.ServerOption

    grpcServer := grpc.NewServer(opts...)
    pb.RegisterObjectStoreServer(grpcServer, &DummyObjectStoreServer{})
    grpcServer.Serve(lis)
}
