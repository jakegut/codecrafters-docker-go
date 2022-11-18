package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/schollz/progressbar/v3"
)

func ensureLayerDownloaded(docker *DockerAPI, blobsum string) (string, error) {
	destPath := path.Join(os.TempDir(), "mydocker", "layers", blobsum)
	_, err := os.Stat(destPath)
	fmt.Printf("Checking %s\n", destPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = downloadLayer(docker, destPath, blobsum); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return destPath, nil
}

func downloadLayer(docker *DockerAPI, destPath, blobsum string) error {
	fmt.Printf("Downloading layer '%s'\n", blobsum)
	if err := os.MkdirAll(path.Dir(destPath), 0750); err != nil {
		return err
	}
	resp, err := docker.GetBlobResp(blobsum)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading...",
	)
	defer bar.Close()
	io.Copy(io.MultiWriter(f, bar), resp.Body)
	return nil
}
