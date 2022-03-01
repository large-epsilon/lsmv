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
    // TODO: load the root tree instead of using a special inode
	root_dir := Dir{
		inode:    1,
		mode:     os.ModeDir | 0o555,
		files:    &map[string]File{},
		children: &map[string]Dir{},
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
