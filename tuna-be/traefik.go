package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func ensureTraefik() error {
	_, err := dockerClient.ContainerInspect(context.Background(), "tuna-traefik")
	if err == nil {
		fmt.Println("Traefik is already running, skipping setup.")
		return nil
	}
	if !client.IsErrNotFound(err) {
		return fmt.Errorf("failed to inspect Traefik container: %w", err)
	}

	_, err = dockerClient.ContainerCreate(context.Background(),
		&container.Config{
			Image:  "traefik",
			Labels: H{"traefik.enable": "false"},
			Cmd: []string{
				"--providers.docker",
				"--entrypoints.web.address=:80",
				"--entrypoints.websecure.address=:443",
				fmt.Sprintf("--certificatesresolvers.myresolver.acme.email=%s", os.Getenv("ACME_EMAIL")),
				"--certificatesresolvers.myresolver.acme.storage=/letsencrypt/acme.json",
				"--certificatesresolvers.myresolver.acme.httpchallenge=true",
				"--certificatesresolvers.myresolver.acme.httpchallenge.entrypoint=web",
			},
		},
		&container.HostConfig{
			Binds: []string{"/var/run/docker.sock:/var/run/docker.sock:ro"},
			PortBindings: map[nat.Port][]nat.PortBinding{
				"80/tcp":  {{HostPort: "80"}},
				"443/tcp": {{HostPort: "443"}},
			},
		}, nil, nil, "tuna-traefik")

	if err != nil {
		return err
	}

	return dockerClient.ContainerStart(context.Background(), "tuna-traefik", container.StartOptions{})
}
