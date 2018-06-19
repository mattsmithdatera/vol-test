package lib

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

const (
	Checkmark   = `✔`
	Xmark       = `✗`
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	ctxt = context.Background()
	Done = make(chan int, 2)
)

type TestClient struct {
	Node     string
	Plugin   string
	PlugOpts map[string]string
	VolOpts  map[string]string
	dclient  *client.Client
	vol      string
	con      string
}

func NewTestClient(plugin string, plugOpts, volOpts map[string]string, nodeFile string) *TestClient {
	SourceFile(nodeFile)
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return &TestClient{
		Plugin:   plugin,
		Node:     nodeFile,
		PlugOpts: plugOpts,
		VolOpts:  volOpts,
		dclient:  cli,
	}
}

func applyColor(toBeColored, color string) string {
	suffix := "\x1b[0m"
	colors := map[string]string{
		"red":     "\x1b[31m",
		"green":   "\x1b[32m",
		"yellow":  "\x1b[33m",
		"cyan":    "\x1b[36m",
		"magenta": "\x1b[35m",
	}
	return fmt.Sprintf("%s%s%s", colors[color], toBeColored, suffix)
}

func GetFunctionName(i interface{}) string {
	long := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(long, `.`)
	last := parts[len(parts)-1]
	return strings.TrimRight(last, "-fm")
}

func RunTestFunc(fn func() error) string {
	testName := GetFunctionName(fn)
	result := fmt.Sprintf("Test: %s ", testName)
	if err := fn(); err != nil {
		result = fmt.Sprintf("%s ", applyColor(Xmark, "red")) + result
		result += fmt.Sprintf("Test %s failed. error: %s", testName, err)
		return result
	}
	result = fmt.Sprintf("%s ", applyColor(Checkmark, "green")) + result
	return result
}

func (c *TestClient) InstallPlugin() error {
	log.Debugf("Installing Plugin: %s\n", c.Plugin)
	reader, err := c.dclient.PluginInstall(ctxt, c.Plugin, types.PluginInstallOptions{
		RemoteRef:            c.Plugin,
		AcceptAllPermissions: true,
	})
	if reader != nil {
		defer reader.Close()
	}
	if err != nil {
		return err
	}
	// This forces us to wait until the install is complete
	_, err = io.Copy(ioutil.Discard, reader)
	return nil
}

func (c *TestClient) DeletePlugin() error {
	log.Debugf("Deleting Plugin: %s", c.Plugin)
	err := c.dclient.PluginDisable(ctxt, c.Plugin, types.PluginDisableOptions{
		Force: true,
	})
	if err != nil {
		return err
	}
	err = c.dclient.PluginRemove(ctxt, c.Plugin, types.PluginRemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *TestClient) CreateVolume() error {
	volName := c.Node + "-" + RandString(5)
	c.vol = volName
	log.Debugf("Creating volume: %s with driver %s\n", volName, c.Plugin)
	_, err := c.dclient.VolumeCreate(ctxt, volume.VolumesCreateBody{
		Driver:     c.Plugin,
		Name:       volName,
		DriverOpts: c.VolOpts,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *TestClient) ConfirmVolume() error {
	log.Debugf("Confirming volume: %s with driver %s\n", c.vol, c.Plugin)
	f := filters.KeyValuePair{
		Key:   "name",
		Value: c.vol,
	}
	vok, err := c.dclient.VolumeList(ctxt, filters.NewArgs(f))
	if err != nil {
		return err
	}
	if len(vok.Volumes) == 0 {
		return fmt.Errorf("Volume %s not found", c.vol)
	}
	return nil
}

func (c *TestClient) InspectVolume() error {
	volName := c.vol
	log.Debugf("Inspecting volume: %s with driver %s\n", volName, c.Plugin)
	vok, err := c.dclient.VolumeList(ctxt, filters.NewArgs())
	if err != nil {
		return err
	}
	found := ""
	for _, vol := range vok.Volumes {
		if strings.HasPrefix(vol.Name, volName) {
			found = vol.Name
		}
	}
	if found == "" {
		return fmt.Errorf("Volume %s not found", volName)
	}

	vol, err := c.dclient.VolumeInspect(ctxt, found)
	if err != nil {
		return err
	}
	driver := strings.Replace(vol.Driver, ":latest", "", 1)
	if driver != c.Plugin {
		return fmt.Errorf("Driver does not match.  %s != %s", driver, c.Plugin)
	}
	return nil
}

func (c *TestClient) DeleteVolume(volName string) error {
	log.Debugf("Deleting volume: %s\n", volName)
	return c.dclient.VolumeRemove(ctxt, volName, true)
}

func (c *TestClient) CreateContainerWithVolume() error {
	conName := c.Node + "-" + RandString(5)
	c.con = conName
	log.Debugf("Creating container [%s] with volume [%s]\n", c.con, c.vol)
	conf := &container.Config{
		Cmd: []string{`/bin/bash`},
	}
	hconf := &container.HostConfig{
		Binds:        []string{fmt.Sprintf("%s:/data", c.vol)},
		VolumeDriver: c.Plugin,
	}
	nconf := &network.NetworkingConfig{}
	_, err := c.dclient.ContainerCreate(ctxt, conf, hconf, nconf, conName)
	if err != nil {
		return err
	}
	return nil
}

// func (c *TestClient) WriteDataToVolume() {
// 	_, err := c.dclient.ContainerExecAttach(ctxt, c.con, types.ExecStartType{
// 		Detach: true,
// 		Tty:    true,
// 	})

// }

func (c *TestClient) CleanVolumes() {
	log.Debug("Cleaning volumes")
	vok, err := c.dclient.VolumeList(ctxt, filters.NewArgs())
	if err != nil {
		panic(err)
	}
	for _, vol := range vok.Volumes {
		err := c.DeleteVolume(vol.Name)
		if err != nil {
			panic(err)
		}
	}
}

func (c *TestClient) CleanContainers() error {
	log.Debug("Cleaning containers")
	containers, err := c.dclient.ContainerList(ctxt, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		log.Debugf("Cleaning Container: %s", container.ID)
		err = c.dclient.ContainerRemove(ctxt, container.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *TestClient) Clean() {
	log.Debugf("Cleaning up testbed: %s", c.Node)
	defer func() {
		if r := recover(); true {
			if r != nil {
				log.Errorf("Error cleaning up, %s", r)
			}
			Done <- 1
		}
	}()
	c.subClean()
}

func (c *TestClient) subClean() {
	c.CleanContainers()
	c.CleanVolumes()
	c.DeletePlugin()
	Done <- 1
}

func (c *TestClient) PrintContainers() {
	containers, err := c.dclient.ContainerList(ctxt, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		log.Debugf("%s %s\n", container.ID[:10], container.Image)
	}
}

func (c *TestClient) PrintVolumes() {
	vok, err := c.dclient.VolumeList(ctxt, filters.NewArgs())
	if err != nil {
		panic(err)
	}
	for _, vol := range vok.Volumes {
		log.Debugf("Volume name: %s\n", vol.Name)
	}
}

func (c *TestClient) PrintPlugins() {
	plugins, err := c.dclient.PluginList(ctxt, filters.Args{})
	if err != nil {
		panic(err)
	}

	for _, plugin := range plugins {
		log.Debugf("%s %s\n", plugin.ID[:10], plugin.Name)
	}
}

func PrintEnv() {
	for _, e := range os.Environ() {
		log.Debug(e)
	}
}

func SourceFile(filename string) {
	log.Debugf("Sourcing file: %s\n", filename)
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
				log.Debugf("Setting k=%s, v=%s\n", kv[0][7:len(kv[0])], kv[1][1:len(kv[1])-1])
				os.Setenv(kv[0][7:len(kv[0])], kv[1][1:len(kv[1])-1])
			}
		}
		if os.Getenv("DOCKER_API_VERSION") == "" {
			os.Setenv("DOCKER_API_VERSION", "1.37")
		}
	}

}

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
