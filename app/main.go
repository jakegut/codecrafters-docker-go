package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
	// Uncomment this block to pass the first stage!
	// "os"
	// "os/exec"
)

func handleError(e error) {
	if e != nil {
		log.Fatalf("%+v", e)
	}
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	image := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	docker := NewDockerAPI(image)
	handleError(docker.Auth())
	paths, err := docker.DownloadImage()
	handleError(err)

	tmpChroot, err := ioutil.TempDir("", "")
	handleError(err)

	handleError(copyExecToDir(tmpChroot, command))
	handleError(extractTarsToDir(tmpChroot, paths))

	handleError(createDevNull(tmpChroot))

	handleError(syscall.Chroot(tmpChroot))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Err: %+v", err)
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		} else {
			os.Exit(1)
		}
	}
}

func copyExecToDir(chootDir, execPath string) error {
	execPathInDir := path.Join(chootDir, execPath)

	if err := os.MkdirAll(path.Dir(execPathInDir), 0750); err != nil {
		return err
	}

	return copyFile(execPath, execPathInDir)
}

func extractTarsToDir(chootDir string, paths []string) error {
	for _, path := range paths {
		fmt.Printf("\tExtracting '%s'\n", path)
		cmd := exec.Command("tar", "xf", path, "-C", chootDir)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dest string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, srcStat.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func createDevNull(chroot string) error {
	if err := os.MkdirAll(path.Join(chroot, "dev"), 0750); err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(chroot, "dev", "null"), []byte{}, 0644)
}
