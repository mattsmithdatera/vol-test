package main

import (
	"os"

	"github.com/docker/vol-test/pkg/lib"
)

func Main() int {
	cli1 := lib.NewTestClient("node1")
	cli2 := lib.NewTestClient("node2")

	cli1.InstallPlugin("dateraiodev/docker-driver:latest")
	cli2.InstallPlugin("dateraiodev/docker-driver:latest")

	println("This is a triumph!!!")
	return 0
}

func main() {
	os.Exit(Main())
}
