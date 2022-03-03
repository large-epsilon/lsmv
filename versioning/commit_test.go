package versioning

import (
	"net"
	"testing"
	"time"

	data_pb "lsmv/proto/data"
	pb "lsmv/proto/versioning"
)

func buildRequest() pb.PushCommitRequest {
	return pb.PushCommitRequest{
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

func TestCommit(t *testing.T) {
	objectstoreListener, err := net.Listen(
		"unix", "/tmp/versioning_test_TestPushCommit_objectstore.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer objectstoreListener.Close()

	server := &VersioningServerImpl{
		objectstoreAddress: "unix:///tmp/versioning_test_TestPushCommit.sock",
	}

	_, err = versioningClient.PushCommit(
		context.background(),
		&request,
	)
	if err != nil {
		t.Fatal(err)
	}

}
