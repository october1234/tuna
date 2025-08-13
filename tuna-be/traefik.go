package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
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
			NetworkMode: container.NetworkMode("tuna-ingress"),
		}, nil, nil, "tuna-traefik")

	if err != nil {
		return err
	}

	return dockerClient.ContainerStart(context.Background(), "tuna-traefik", container.StartOptions{})
}

func ensureTraefikNetwork() error {
	filter := filters.NewArgs()
	filter.Add("name", "tuna-ingress")

	networks, err := dockerClient.NetworkList(context.Background(), network.ListOptions{
		Filters: filter,
	})
	if len(networks) != 0 {
		fmt.Println("Docker network tuna-ingress already exists, skipping creation.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to list docker networks: %w", err)
	}

	_, err = dockerClient.NetworkCreate(context.Background(), "tuna-ingress", network.CreateOptions{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: "172.28.0.0/16",
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create docker network: %w", err)
	}

	return nil
}
