package main

import (
	"context"
	iofs "io/fs"
	"log"
	"math/rand"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	data_pb "lsmv/proto/data"
)

type Dir struct {
	// Map pointers are used here to make the dir struct hashable.
	files      *map[string]File
	children   *map[string]Dir
	hash       string
	inode      uint64
	mode       iofs.FileMode
	tree       *data_pb.Tree
	loaded     bool
	controller *Controller
}

func (d Dir) maybeLoad() error {
	if d.loaded {
		return nil
	}

	tree, err := d.controller.getDir(d.hash)
	if err != nil {
		return err
	}
	d.tree = tree

	for _, child := range d.tree.Children {
		switch child.Type {
		case data_pb.Tree_Child_BLOB:
			(*d.files)[child.Name] = File{
				content:    &[]byte{},
				hash:       child.Hash,
				inode:      uint64(rand.Int63()),
				mode:       0o444, // TODO store modes in tree protos
				controller: d.controller,
			}
		case data_pb.Tree_Child_SUBTREE:
			(*d.children)[child.Name] = Dir{
				files:      &map[string]File{},
				children:   &map[string]Dir{},
				hash:       child.Hash,
				inode:      uint64(rand.Int63()),
				mode:       os.ModeDir | 0o555, // TODO store modes in tree protos
				controller: d.controller,
			}
		}
	}
	d.loaded = true
	return nil
}

func (d Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = d.inode
	attr.Mode = d.mode
	return nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	err := d.maybeLoad()
	if err != nil {
		log.Printf("failed to load directory: %v", err)
		return nil, err
	}

	child, ok := (*d.children)[name]
	if ok {
		return child, nil
	}
	file, ok := (*d.files)[name]
	if ok {
		return file, nil
	}
	return nil, syscall.ENOENT
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	err := d.maybeLoad()
	if err != nil {
		log.Printf("failed to load directory: %v", err)
		return nil, err
	}

	response := []fuse.Dirent{}
	for name, child := range *d.children {
		response = append(response, fuse.Dirent{Inode: child.inode, Name: name, Type: fuse.DT_Dir})
	}
	for name, file := range *d.files {
		response = append(response, fuse.Dirent{Inode: file.inode, Name: name, Type: fuse.DT_File})
	}
	return response, nil
}
