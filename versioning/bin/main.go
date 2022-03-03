package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "lsmv/proto/versioning"
	"lsmv/versioning"
)

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
	pb.RegisterVersioningServer(grpcServer, &versioning.VersioningServerImpl{
		ObjectstoreAddress: *objectstoreAddress,
	})
	log.Printf("Starting versioning server on %s:%d.", *host, *port)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
