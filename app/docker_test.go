package main

import (
	"testing"
)

func TestDockerAuth(t *testing.T) {
	dockerAPI := NewDockerAPI("ubuntu:latest")

	if dockerAPI.repo != "library/ubuntu" {
		t.Errorf("wanted: '%s', got '%s'", "library/ubuntu", dockerAPI.repo)
	}

	if dockerAPI.ref != "latest" {
		t.Errorf("wanted: '%s', got '%s'", "latest", dockerAPI.ref)
	}
	err := dockerAPI.Auth()
	if err != nil {
		t.Errorf("Auth() error: %+v", err)
	}

	if dockerAPI.authToken == "" {
		t.Error("auth token not set")
	}
}

func TestDockerGetManifest(t *testing.T) {
	dockerAPI := NewDockerAPI("ubuntu:latest")
	err := dockerAPI.Auth()
	if err != nil {
		t.Errorf("Auth() error: %+v", err)
	}
	_, err = dockerAPI.GetManifest()
	if err != nil {
		t.Errorf("GetManifest() error: %+v", err)
	}
}
