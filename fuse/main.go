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

	// TODO: smarter controller construction
	filesystem := LsmvFS{}
	controller := Controller{
		filesystem:              &filesystem,
		currentHead:             "asdffdsa",
		versioningServerAddress: "localhost:7886",
	}
	control, err := constructControlDir(&controller)
	if err != nil {
		log.Fatalf("Failed to construct control directory: %v", err)
	}
	filesystem.control = &control

	// TODO: defer unmount
	conn, err := fuse.Mount(
		mountLocation, fuse.FSName("lsmv"), fuse.Subtype("lsmvfs"))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = fs.Serve(conn, filesystem)
	if err != nil {
		log.Fatal(err)
	}
}
