package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"lsmv/objectstore/in_memory_store"
	pb "lsmv/proto/objectstore"
)

func main() {
	var port = flag.Int("port", 7887, "Port for the objectstore server to listen on.")
	var host = flag.String("host", "localhost", "Hostname for this server.")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterObjectStoreServer(grpcServer, in_memory_store.New())
	log.Printf("Starting object store server on %s:%d.", *host, *port)
	grpcServer.Serve(lis)
}
