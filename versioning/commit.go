package versioning

import (
    "context"
    "log"
    "fmt"

    "google.golang.org/grpc"

    pb "lsmv/proto/versioning"
    objectstore_pb "lsmv/proto/objectstore"
)

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

	conn, err := grpc.Dial(v.ObjectstoreAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf(
			"Failed to dial objectstore server at %s: %v",
			v.ObjectstoreAddress, err)
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

	conn, err := grpc.Dial(v.ObjectstoreAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf(
			"Failed to dial objectstore server at %s: %v",
			v.ObjectstoreAddress, err)
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
