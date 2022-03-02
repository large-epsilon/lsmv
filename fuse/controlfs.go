package main

import (
	"context"
	iofs "io/fs"
	"log"
	"os"
	"strings"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// ************************************************************************** //
// ControlFS
//
// A directory containing special control files that constitute the api for
// this filesystem, along the lines of /proc or /sys.
// ************************************************************************** //

type ControlDir struct {
	nodes      *map[string]fs.Node
	controller Controller
}

func constructControlDir(controller *Controller) (ControlDir, error) {
	return ControlDir{
		nodes: &map[string]fs.Node{
			"head": headFile{
				mode:       0o664,
				controller: controller,
			},
		},
	}, nil
}

func (d ControlDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = 2
	attr.Mode = os.ModeDir | 0o555
	return nil
}

func (d ControlDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	node, ok := (*d.nodes)[name]
	if ok {
		return node, nil
	}
	return nil, syscall.ENOENT
}

func (d ControlDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	response := []fuse.Dirent{}
	for name, _ := range *d.nodes {
		response = append(
			response,
			fuse.Dirent{Inode: 0, Name: name, Type: fuse.DT_File},
		)
	}
	return response, nil
}

// headFile points to the commit that is the head of the repo right now.
// Reading this file returns the hash of that commit.
// Writing this file sets the head to the specified commit and updates the
// working tree accordingly.
type headFile struct {
	mode       iofs.FileMode
	controller *Controller
}

func (f headFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = 0
	attr.Mode = f.mode
	attr.Size = uint64(len((*f.controller).currentHead))
	return nil
}

func (f headFile) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte((*f.controller).currentHead + "\n"), nil
}

func (f headFile) Write(
	ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse,
) error {
	commitHash := strings.TrimSpace(string(req.Data))
	if len(commitHash) == 0 {
		// Ignore empty writes, which is a thing echo likes to do.
		log.Printf("skipping zero write")
		resp.Size = len((*f.controller).currentHead)
		return nil
	}

	err := f.controller.setHead(commitHash)
	if err != nil {
		log.Printf("Failed writing `head`: %v", err)
		return err
	}
	resp.Size = len(commitHash)
	return nil
}
