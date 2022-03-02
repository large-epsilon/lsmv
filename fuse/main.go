package main

import (
	"flag"
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		panic("Incorrect number of arguments, expected 1 (mount path)")
	}
	mountLocation := flag.Arg(0)

	conn, err := fuse.Mount(
		mountLocation, fuse.FSName("lsmv"), fuse.Subtype("lsmvfs"))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	defer fuse.Unmount(mountLocation)

	srv := fs.New(conn, nil)

	filesystem, err := NewLsmvFS("repo", srv)
	if err != nil {
		log.Fatal(err)
	}

	err = srv.Serve(filesystem)
	if err != nil {
		log.Fatal(err)
	}
}
