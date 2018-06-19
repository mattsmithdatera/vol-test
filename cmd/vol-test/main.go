package main

import (
	"flag"
	"os"
	"strings"

	"github.com/docker/vol-test/pkg/lib"
	log "github.com/sirupsen/logrus"
)

var (
	debug     = flag.Bool("debug", false, "")
	cleanOnly = flag.Bool("clean-only", false, "")
)

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func Main() int {
	cli1 := lib.NewTestClient("dateraiodev/docker-driver", "node1")
	cli2 := lib.NewTestClient("dateraiodev/docker-driver", "node2")

	if !*cleanOnly {

		log.Debug("Starting volume tests")
		tests := []func() error{
			cli1.InstallPlugin,
			cli2.InstallPlugin,

			cli1.CreateVolume,
			cli2.CreateVolume,

			cli1.ConfirmVolume,
			cli2.ConfirmVolume,

			cli1.InspectVolume,
			cli2.InspectVolume,

			cli1.CreateContainerWithVolume,
			cli2.CreateContainerWithVolume,
		}
		results := []string{}
		for _, v := range tests {
			results = append(results, lib.RunTestFunc(v))
		}

		fail := false
		for _, v := range results {
			log.Info(v)
			if strings.Contains(v, lib.Xmark) {
				fail = true
			}
		}
		if !fail {
			log.Info("\nThis is a triumph!!!\n")
		} else {
			log.Info("\nThis is a failure!!!\n")
		}
	} else {
		log.Info("Skipping tests, running just testbed cleaning")
	}

	clean := func() { go cli1.Clean(); go cli2.Clean() }
	clean()

	<-lib.Done
	<-lib.Done

	return 0
}

func main() {
	os.Exit(Main())
}
