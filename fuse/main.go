package main

import (
    "flag"
    "log"
    "os"

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
    root_dir := Dir {
        inode: 1,
        mode: os.ModeDir | 0o555,
        files: &map[string]File{},
        children: &map[string]Dir{},
    }
    content := []byte("Hello you beautiful person!\n")
    (*root_dir.files)["hello"] = File{
        content: &content,
        inode: 2,
        mode: 0o444,
    }

    return root_dir, nil
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
