package main

import (
    "context"
    "flag"
    "log"
    "os"
    "syscall"

    "bazil.org/fuse"
    "bazil.org/fuse/fs"
)

type FuseFS struct {}

func (FuseFS) Root() (fs.Node, error) {
    return Dir{}, nil
}

type Dir struct {}

func (Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
    attr.Inode = 1
    attr.Mode = os.ModeDir | 0o555
    return nil
}

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
    if name == "hello" {
        return File{}, nil
    }
    return nil, syscall.ENOENT
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
    return []fuse.Dirent{{Inode: 2, Name: "hello", Type: fuse.DT_File}}, nil
}

const content = "Hello, world!\n"
type File struct {}

func (File) Attr(ctx context.Context, attr *fuse.Attr) error {
    attr.Inode = 2
    attr.Mode = 0o444
    attr.Size = uint64(len(content))
    return nil
}

func (File) ReadAll(ctx context.Context) ([]byte, error) {
    return []byte(content), nil
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

    err = fs.Serve(conn, FuseFS{})
    if err != nil {
        log.Fatal(err)
    }
}
