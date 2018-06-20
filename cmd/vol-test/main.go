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
	driver := os.Getenv("VOLDRIVER")
	// pluginOpts := os.Getenv("PLUGINOPTS")
	// createOpts := os.Getenv("CREATEOPTS")
	//TODO (_alastor_): Actually use pluginOpts and createOpts
	cli1 := lib.NewTestClient(driver, map[string]string{}, map[string]string{}, "node1")
	cli2 := lib.NewTestClient(driver, map[string]string{}, map[string]string{}, "node2")

	if !*cleanOnly {

		log.Debug("Starting volume tests")
		tests := []string{
			"InstallPlugin",

			"CreateVolume",

			"ConfirmVolume",

			"InspectVolume",

			"CreateContainerWithVolume",
		}
		results := []string{}
		for _, v := range tests {
			for _, t := range []*lib.TestClient{cli1, cli2} {
				results = append(results, lib.RunTestFunc(t, v))
			}
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
