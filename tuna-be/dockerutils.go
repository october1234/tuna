package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

func deleteContainerByName(cli *client.Client, name string) error {
	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("name", name)

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return nil
	}

	c := containers[0]
	if err := cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", c.ID, err)
	}
	fmt.Printf("Removed container: %s\n", c.ID)
	return nil
}

func pullImage(cli *client.Client, img string) error {
	ctx := context.Background()
	_, err := cli.ImagePull(ctx, img, image.PullOptions{})
	return err
}

func runContainer(cli *client.Client, image string, name string, env H, labels H, volumes H) (string, error) {
	ctx := context.Background()

	// SUS!!!
	mounts := []mount.Mount{}
	for volName, volPath := range volumes {
		if _, err := createVolume(cli, formatDockerResourceName(volName)); err != nil {
			return "", fmt.Errorf("failed to create volume %s: %w", volName, err)
		}

		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeVolume,
			Source:   volName,
			Target:   volPath,
			ReadOnly: false,
		})
	}

	envArr := []string{}
	for k, v := range env {
		envArr = append(envArr, fmt.Sprintf("%s=%s", k, v))
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:  image,
		Env:    envArr,
		Labels: labels,
	}, &container.HostConfig{
		Mounts: mounts,
	}, nil, nil, name)
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func createVolume(cli *client.Client, name string) (string, error) {
	vol, err := cli.VolumeCreate(context.Background(), volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return vol.Name, nil
}

func stopContainerByName(cli *client.Client, name string) error {
	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("name", name)

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	c := containers[0]
	if err := cli.ContainerStop(ctx, c.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", c.ID, err)
	}
	fmt.Printf("Stopped container: %s\n", c.ID)
	return nil
}
