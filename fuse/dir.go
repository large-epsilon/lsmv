package main

import (
    "context"
    iofs "io/fs"
    "syscall"

    "bazil.org/fuse"
    "bazil.org/fuse/fs"
)

type Dir struct {
    // Map pointers are used here to make the dir struct hashable.
    files *map[string]File
    children *map[string]Dir
    inode uint64
    mode iofs.FileMode
}

func (d Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
    attr.Inode = d.inode
    attr.Mode = d.mode
    return nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
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
    // TODO: handle lazy-loading subtrees
    response := []fuse.Dirent{}
    for name, child := range *d.children {
        response = append(response, fuse.Dirent{Inode: child.inode, Name: name, Type: fuse.DT_Dir})
    }
    for name, file := range *d.files {
        response = append(response, fuse.Dirent{Inode: file.inode, Name: name, Type: fuse.DT_File})
    }
    return response, nil
}
