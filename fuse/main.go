package main

import (
    "context"
    "flag"
    iofs "io/fs"
    "log"
    "syscall"

    "bazil.org/fuse"
    "bazil.org/fuse/fs"

    objectstore_pb "lsmv/proto/objectstore"
)

type LsmvFS struct {
    tree *objectstore_pb.Tree
    root *Dir
}

func (fs LsmvFS) setRootTree(tree *objectstore_pb.Tree) error {
    // TODO
    return nil
}

func (LsmvFS) Root() (fs.Node, error) {
    return Dir{}, nil
}

type Dir struct {
    files map[string]File
    children map[string]Dir
    inode uint64
    mode iofs.FileMode
}

func (d Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
    //attr.Inode = 1
    //attr.Mode = os.ModeDir | 0o555
    attr.Inode = d.inode
    attr.Mode = d.mode
    return nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
    child, ok := d.children[name]
    if ok {
        return child, nil
    }
    file, ok := d.files[name]
    if ok {
        return file, nil
    }
    return nil, syscall.ENOENT
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
    // TODO: handle lazy-loading subtrees
    response := []fuse.Dirent{}
    for name, child := range d.children {
        response = append(response, fuse.Dirent{Inode: child.inode, Name: name, Type: fuse.DT_Dir})
    }
    for name, file := range d.files {
        response = append(response, fuse.Dirent{Inode: file.inode, Name: name, Type: fuse.DT_File})
    }
    return response, nil
}

type File struct {
    // TODO: don't hold everything in memory, cache to disk.
    content []byte
    inode uint64
    mode iofs.FileMode
}

func (f File) Attr(ctx context.Context, attr *fuse.Attr) error {
    // attr.Inode = 2
    // attr.Mode = 0o444
    // attr.Size = uint64(len(content))
    attr.Inode = f.inode
    attr.Mode = f.mode
    attr.Size = uint64(len(f.content))
    return nil
}

func (f File) ReadAll(ctx context.Context) ([]byte, error) {
    return f.content, nil
}

func main() {
    flag.Parse()
    if flag.NArg() != 1 {
        panic("Incorrect number of arguments, expected 1 (mount path)")
    }
    mountLocation := flag.Arg(0)

    conn, err := fuse.Mount(mountLocation, fuse.FSName("helloworld"), fuse.Subtype("hellofs"))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    err = fs.Serve(conn, LsmvFS{})
    if err != nil {
        log.Fatal(err)
    }
}
