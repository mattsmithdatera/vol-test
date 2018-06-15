package lib

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type TestClient struct {
	node    string
	dclient *client.Client
}

var ctxt = context.Background()

func PrintEnv() {
	for _, e := range os.Environ() {
		fmt.Println(e)
	}
}

func SourceFile(filename string) {
	fmt.Printf("Sourcing file: %s\n", filename)
	out, err := exec.Command("cat", filename).CombinedOutput()
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		if strings.Contains(s.Text(), "=") {
			kv := strings.SplitN(s.Text(), "=", 2)
			if len(kv) == 2 {
				// remove 'export ' from the key when setting
				fmt.Printf("Setting k=%s, v=%s\n", kv[0][7:len(kv[0])], kv[1][1:len(kv[1])-1])
				os.Setenv(kv[0][7:len(kv[0])], kv[1][1:len(kv[1])-1])
			}
		}
		if os.Getenv("DOCKER_API_VERSION") == "" {
			os.Setenv("DOCKER_API_VERSION", "1.37")
		}
	}

}

func NewTestClient(nodeFile string) *TestClient {
	SourceFile(nodeFile)
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return &TestClient{
		node:    nodeFile,
		dclient: cli,
	}
}

func (c *TestClient) PrintContainers() {
	containers, err := c.dclient.ContainerList(ctxt, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
}

func (c *TestClient) PrintPlugins() {
	plugins, err := c.dclient.PluginList(ctxt, filters.Args{})
	if err != nil {
		panic(err)
	}

	for _, plugin := range plugins {
		fmt.Printf("%s %s\n", plugin.ID[:10], plugin.Name)
	}
}

func (c *TestClient) InstallPlugin(plugName string) error {
	fmt.Printf("Installing Plugin: %s\n", plugName)
	reader, err := c.dclient.PluginInstall(ctxt, plugName, types.PluginInstallOptions{
		RemoteRef:            plugName,
		AcceptAllPermissions: true,
	})
	if reader != nil {
		defer reader.Close()
	}
	if err != nil {
		return err
	}
	// This forces us to wait until the install is complete
	_, err := io.Copy(ioutil.Discard, reader)
	return nil
}
