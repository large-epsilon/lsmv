package versioning

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"

	"lsmv/objectstore/in_memory_store"
	data_pb "lsmv/proto/data"
	objectstore_pb "lsmv/proto/objectstore"
	pb "lsmv/proto/versioning"
)

var now = uint64(time.Now().Unix())

func buildRequest() pb.PushCommitRequest {
	return pb.PushCommitRequest{
		Commit: &data_pb.Commit{
			Timestamp:      now,
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

func TestCommit(t *testing.T) {
	const objSock = "/tmp/lsmv_versioning_test_TestPushCommit_objectstore.sock"
	objectstoreListener, err := net.Listen(
		"unix", objSock)
	if err != nil {
		t.Fatal(err)
	}
	defer objectstoreListener.Close()
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	objectstore_pb.RegisterObjectStoreServer(grpcServer, in_memory_store.New())
	go func() {
		grpcServer.Serve(objectstoreListener)
	}()
	defer grpcServer.Stop()

	server := &VersioningServerImpl{
		ObjectstoreAddress: "unix://" + objSock,
	}

	request := buildRequest()
	_, err = server.PushCommit(
		context.Background(),
		&request,
	)
	if err != nil {
		t.Fatal(err)
	}

	pullRequest := pb.PullCommitRequest{Hash: "fakecommit"}
	response, err := server.PullCommit(context.Background(), &pullRequest)
	if err != nil {
		t.Fatal(err)
	}
	expectedResponse := &pb.PullCommitResponse{
		Commit: &data_pb.Commit{
			Timestamp:      now,
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
	}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Fatalf("Returned response was different from expected response")
	}

}
