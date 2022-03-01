package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"google.golang.org/grpc"

	data_pb "lsmv/proto/data"
	objectstore_pb "lsmv/proto/objectstore"
	versioning_pb "lsmv/proto/versioning"
)

type RootDir struct {
	control  *ControlDir
	repoRoot *Dir
	repoName string
}

func (RootDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = 1
	attr.Mode = os.ModeDir | 0o555
	return nil
}

func (r RootDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == ".control" {
		return *r.control, nil
	} else if name == r.repoName {
		return *r.repoRoot, nil
	}
	return nil, syscall.ENOENT
}

func (r RootDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Inode: 2, Name: ".control", Type: fuse.DT_Dir},
		{Inode: 3, Name: r.repoName, Type: fuse.DT_Dir},
	}, nil
}

type LsmvFS struct {
	control  *ControlDir
	repoRoot *Dir
	repoName string
}

func (fs *LsmvFS) setRootTree(hash string, controller *Controller) error {
	fs.repoRoot = &Dir{
		inode:      3,
		hash:       hash,
		mode:       os.ModeDir | 0o555,
		files:      &map[string]File{},
		children:   &map[string]Dir{},
		controller: controller,
	}
	return nil
}

func (fs LsmvFS) Root() (fs.Node, error) {
	return RootDir{
		control:  fs.control,
		repoName: fs.repoName,
		repoRoot: fs.repoRoot,
	}, nil
}

// TODO: swap this to disk somewhere
type Controller struct {
	versioningServerAddress  string
	objectstoreServerAddress string
	filesystem               *LsmvFS
	currentHead              string
}

func (c *Controller) getFile(hash string) (*data_pb.Blob, error) {
	conn, err := grpc.Dial(c.objectstoreServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf(
			"Failed to dial objectstore server at %s: %v",
			c.objectstoreServerAddress, err)
		return nil, err
	}

	objectstoreClient := objectstore_pb.NewObjectStoreClient(conn)
	response, err := objectstoreClient.GetObject(
		context.TODO(), &objectstore_pb.GetObjectRequest{Hash: hash})
	if err != nil {
		return nil, err
	}

	switch x := response.ReturnedObject.(type) {
	case *objectstore_pb.GetObjectResponse_Blob:
		return x.Blob, nil
	default:
		return nil, fmt.Errorf(
			"incorrect type for object '%s': %T, expected blob",
			hash, response.ReturnedObject)
	}
}

func (c *Controller) getDir(hash string) (*data_pb.Tree, error) {
	conn, err := grpc.Dial(c.objectstoreServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf(
			"Failed to dial objectstore server at %s: %v",
			c.objectstoreServerAddress, err)
		return nil, err
	}

	objectstoreClient := objectstore_pb.NewObjectStoreClient(conn)
	response, err := objectstoreClient.GetObject(
		context.TODO(), &objectstore_pb.GetObjectRequest{Hash: hash})
	if err != nil {
		return nil, err
	}

	switch x := response.ReturnedObject.(type) {
	case *objectstore_pb.GetObjectResponse_Tree:
		return x.Tree, nil
	default:
		return nil, fmt.Errorf(
			"incorrect type for object '%s': %T, expected tree",
			hash, response.ReturnedObject)
	}
}

func (c *Controller) setHead(hash string) error {
	conn, err := grpc.Dial(c.versioningServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf(
			"Failed to dial versioning server at %s: %v",
			c.versioningServerAddress, err)
		return err
	}
	defer conn.Close()

	versioningClient := versioning_pb.NewVersioningClient(conn)
	response, err := versioningClient.PullCommit(
		context.TODO(),
		&versioning_pb.PullCommitRequest{Hash: hash},
	)
	if err != nil {
		log.Printf("Failed to pull commit '%s': %v", hash, err)
		return err
	}

	err = c.filesystem.setRootTree(response.Root.Hash, c)
	if err != nil {
		log.Printf("Failed to set root tree: %v", err)
		return err
	}
	return nil
}
