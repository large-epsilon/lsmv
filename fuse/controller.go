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
	fuse     *fs.Server
}

func (RootDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = 1
	attr.Mode = os.ModeDir | 0o555
	return nil
}

func (r RootDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == ".control" {
		return *r.control, nil
	} else if name == r.repoName && r.repoRoot != nil {
		return *r.repoRoot, nil
	}
	return nil, syscall.ENOENT
}

func (r RootDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	response := []fuse.Dirent{{Inode: 2, Name: ".control", Type: fuse.DT_Dir}}
	if r.repoRoot != nil {
		response = append(
			response,
			fuse.Dirent{Inode: 3, Name: r.repoName, Type: fuse.DT_Dir},
		)
	}
	return response, nil
}

type LsmvFS struct {
	control  *ControlDir
	rootNode *RootDir
	server   *fs.Server
	repoName string
}

func NewLsmvFS(name string) (*LsmvFS, error) {
	fs := LsmvFS{
		repoName: name,
	}
	// TODO: take these values as flags, cache currentHead
	controller := Controller{
		filesystem:               &fs,
		currentHead:              "asdffdsa",
		versioningServerAddress:  "localhost:7886",
		objectstoreServerAddress: "localhost:7887",
	}
	control, err := constructControlDir(&controller)
	if err != nil {
		return nil, err
	}
	fs.control = &control
	fs.rootNode = &RootDir{
		repoName: name,
		control:  &control,
	}
	err = fs.setRootTree(controller.currentHead, &controller)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func (fs *LsmvFS) setRootTree(hash string, controller *Controller) error {
	if fs.rootNode.repoRoot == nil {
		fs.rootNode.repoRoot = &Dir{
			inode:      3,
			mode:       os.ModeDir | 0o555,
			controller: controller,
		}
	}
	fs.rootNode.repoRoot.files = &map[string]File{}
	fs.rootNode.repoRoot.children = &map[string]Dir{}
	fs.rootNode.repoRoot.hash = hash
	fs.rootNode.repoRoot.loaded = false

	return nil
}

func (fs LsmvFS) Root() (fs.Node, error) {
	return fs.rootNode, nil
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

	c.currentHead = hash
	return nil
}
