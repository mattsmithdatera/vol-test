## Introduction

vol-test is a set of integration tests that is intended to prove and test API support of volume plugins for Docker. vol-test is based upon BATS(https://github.com/sstephenson/bats.git) and depends on some helper libraries - bats-support and bats-assert which are linked as submodules.

vol-test supports testing against remote environments. Remote Docker hosts should have ssh keys configured for access without a password.

## Setup

- Set up docker machines (node1 & node2)
    ```
    ssh-copy-id user@node1
    docker-machine create --driver generic --generic-ip-address <node1-ip> --generic-ssh-key <your_ssh_key> --generic-ssh-user <node1-user> node1
    ssh-copy-id user@node2
    docker-machine create --driver generic --generic-ip-address <node2-ip> --generic-ssh-key <your_ssh_key> --generic-ssh-user <node2-user> node2
    ```

- Clone and build this repository (optionally, fork)
    ```
    git clone https://github.com/khudgins/vol-test
    cd vol-tests
    make local
    ```

## Running

- Configuration:

vol-test requires a few environment variables to be configured before running:

* VOLDRIVER - this should be set to the full path (store/vendor/pluginname:tag) of the volume driver to be tested
* PLUGINOPTS - Gets appended to the 'docker volume install' command for install-time plugin configuration
* CREATEOPTS - Optional. Used in 'docker volume create' commands in testing to pass options to the driver being tested
* PREFIX - Optional. Commandline prefix for remote testing. Usually set to 'ssh address_of_node1'
* PREFIX2 - Optional. Commandline prefix for remote testing. Usually set to 'ssh address_of_node2'


- To validate a volume plugin:

1. Export the name of the plugin that is referenced when creating a network as the environmental variable `$VOLDRIVER`.
2. Run the tests by running `./bin/vol-tests`

Example using the vieux/sshfs driver (replace `vieux/sshfs` with the name of the plugin/driver you wish to test):

```
$VOLDRIVER=vieux/sshfs
$CREATEOPTS="-o khudgins@192.168.99.1:~/tmp -o password=yourpw"

./bin/vol-test

INFO[0025] ✔ Test: InstallPlugin
INFO[0025] ✔ Test: CreateVolume
INFO[0025] ✔ Test: ConfirmVolume
...

30 tests, 0 failures
```
