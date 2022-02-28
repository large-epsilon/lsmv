package main

import (
    "context"
    iofs "io/fs"

    "bazil.org/fuse"
)


type File struct {
    // TODO: don't hold contents in memory, cache to disk.
    // Pointer to slice used here to make File struct hashable.
    content *[]byte
    inode uint64
    mode iofs.FileMode
}

func (f File) Attr(ctx context.Context, attr *fuse.Attr) error {
    attr.Inode = f.inode
    attr.Mode = f.mode
    attr.Size = uint64(len(*f.content))
    return nil
}

func (f File) ReadAll(ctx context.Context) ([]byte, error) {
    return *f.content, nil
}
