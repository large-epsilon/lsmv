package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	data_pb "lsmv/proto/data"
	versioning_pb "lsmv/proto/versioning"
)

// Client to inject dummy data into an objectstore server for manual testing.

func buildRequest() versioning_pb.PushCommitRequest {
	return versioning_pb.PushCommitRequest{
		Commit: &data_pb.Commit{
			Timestamp:      uint64(time.Now().Unix()),
			Hash:           "fakecommit",
			Tree:           "faketree",
			Author:         "Raine Serrano",
			AuthorEmail:    "raine.h.serrano@gmail.com",
			Committer:      "Raine Serrano",
			CommitterEmail: "raine.h.serrano@gmail.com",
			Diff:           "[pretend this is a diff]",
		},
		Root: &data_pb.Tree{
			Hash: "faketree",
			Children: []*data_pb.Tree_Child{
				{
					Hash: "aaaa",
					Name: "fakefile",
					Type: data_pb.Tree_Child_BLOB,
				},
				{
					Hash: "bbbb",
					Name: "anotherfile",
					Type: data_pb.Tree_Child_BLOB,
				},
				{
					Hash: "cccc",
					Name: "banana",
					Type: data_pb.Tree_Child_BLOB,
				},
			},
		},
		Files: []*data_pb.Blob{
			{
				Hash:    "aaaa",
				Content: []byte("This is fakefile. It's a fake file.\n"),
			},
			{
				Hash: "bbbb",
				Content: []byte(
					"This is anotherfile. It's another file.\n"),
			},
			{
				Hash: "cccc",
				Content: []byte(
					"This is banana. It's not actually a banana.\n"),
			},
		},
	}
}

func main() {
	conn, err := grpc.Dial(
		"localhost:7886",
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf(
			"Failed to dial versioning server at localhost:7886: %v", err)
	}
	defer conn.Close()

	request := buildRequest()
	versioningClient := versioning_pb.NewVersioningClient(conn)
	_, err = versioningClient.PushCommit(
		context.TODO(),
		&request,
	)
	if err != nil {
		log.Fatalf("Failed to send dummy data: %v", err)
	}
}
