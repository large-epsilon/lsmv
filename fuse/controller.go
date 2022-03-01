package main

import (
	"context"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"google.golang.org/grpc"

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

func (fs *LsmvFS) setRootTree(hash string) error {
	fs.repoRoot = &Dir{
		inode:    3,
		hash:     hash,
		mode:     os.ModeDir | 0o555,
		files:    &map[string]File{},
		children: &map[string]Dir{},
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
	versioningServerAddress string
	filesystem              *LsmvFS
	currentHead             string
}

func (c *Controller) setHead(hash string) error {
	conn, err := grpc.Dial(c.versioningServerAddress)
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

	err = c.filesystem.setRootTree(response.Root.Hash)
	if err != nil {
		log.Printf("Failed to set root tree: %v", err)
		return err
	}
	return nil
}
