package main

import (
	"context"
	iofs "io/fs"
	"log"

	"bazil.org/fuse"
)

type File struct {
	// TODO: don't hold contents in memory, cache to disk.
	// Pointer to slice is used here to make the File struct hashable.
	content    *[]byte
	inode      uint64
	mode       iofs.FileMode
	hash       string
	loaded     bool
	controller *Controller
}

func (f File) maybeLoad() error {
	if f.loaded {
		return nil
	}

	blob, err := f.controller.getFile(f.hash)
	if err != nil {
		return err
	}

	*f.content = blob.Content
	f.loaded = true
	return nil
}

func (f File) Attr(ctx context.Context, attr *fuse.Attr) error {
	// Load content during Attr to get accurate size. If we don't do this, the
	// first read on this file will return nothing.
	// The implication of this is that all the files in a directory will be
	// downloaded when you ls that directory. Maybe that's okay, we're not
	// trying to support very large files anyways.
	err := f.maybeLoad()
	if err != nil {
		log.Printf("failed to load file: %v", err)
		return err
	}

	attr.Inode = f.inode
	attr.Mode = f.mode
	if f.content == nil {
		attr.Size = 0
	} else {
		attr.Size = uint64(len(*f.content))
	}
	return nil
}

func (f File) ReadAll(ctx context.Context) ([]byte, error) {
	err := f.maybeLoad()
	if err != nil {
		log.Printf("failed to load file: %v", err)
		return nil, err
	}
	return *f.content, nil
}
