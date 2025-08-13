package main

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func deploy(deploymentId string, image string, labels, env, volumes H) error {
	containerName := formatDockerResourceName(deploymentId)

	if err := deleteContainerByName(dockerClient, containerName); err != nil {
		return fmt.Errorf("failed to clean existing container: %w", err)
	}

	finalLabels := H{
		"tuna.deployment.id": deploymentId,
	}

	maps.Copy(finalLabels, labels)

	if _, err := runContainer(dockerClient, image, containerName, env, finalLabels, volumes); err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}
	fmt.Printf("Deployment %s completed successfully\n", deploymentId)
	return nil
}

func deployFromImage(deployment *Deployment) error {
	fmt.Println("deployFromImage called with deployment:", deployment.ID)
	if deployment.ModeData.Mode != ModeImage {
		return fmt.Errorf("deployFromImage called w`ith non-image mode: %s", deployment.ModeData.Mode)
	}

	if err := pullImage(dockerClient, deployment.ModeData.Image); err != nil {
		return fmt.Errorf("failed to pull image %s: %w", deployment.ModeData.Image, err)
	}

	err := deploy(deployment.ID, deployment.ModeData.Image, deployment.Labels, deployment.Env, deployment.Volumes)
	if err != nil {
		fmt.Printf("Failed to deploy: %v\n", err)
	}

	fmt.Printf("Deployment %s from image completed successfully\n", deployment.ID)
	return nil
}

func deployFromGit(deployment *Deployment) error {
	fmt.Println("deployFromGit called with deployment:", deployment.ID)
	if deployment.ModeData.Mode != ModeTemplate && deployment.ModeData.Mode != ModeDockerfile {
		return fmt.Errorf("deployFromGit called with non-git mode: %s", deployment.ModeData.Mode)
	}

	tmpDir, err := os.MkdirTemp("", "tuna-workdir-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. Clone repository
	if _, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:           deployment.ModeData.GitData.Repository,
		ReferenceName: plumbing.ReferenceName(deployment.ModeData.GitData.Branch),
		SingleBranch:  true,
	}); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	dockerfile := "Dockerfile"
	if deployment.ModeData.Mode == ModeDockerfile && deployment.ModeData.DockerFile != "" {
		dockerfile = deployment.ModeData.DockerFile
	}

	imageTag := formatBuildImageName(deployment.ID)

	// buildContext, err := archive.TarWithOptions(tmpDir, &archive.TarOptions{})
	// if err != nil {
	// 	return fmt.Errorf("failed to create build context: %v", err)
	// }

	// buildResponse, err := dockerClient.ImageBuild(context.Background(), buildContext, build.ImageBuildOptions{
	// 	Dockerfile: dockerfile,
	// 	Tags:       []string{imageTag},
	// 	Remove:     true,
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to build image: %v", err)
	// }
	// defer buildResponse.Body.Close()

	exec.Command("docker", "build", "-t", imageTag, "-f", path.Join(tmpDir, dockerfile), tmpDir).Run()

	err = deploy(deployment.ID, imageTag, deployment.Labels, deployment.Env, deployment.Volumes)
	if err != nil {
		fmt.Printf("Failed to deploy: %v\n", err)
	}

	fmt.Printf("Deployment %s from git completed successfully\n", deployment.ID)
	return nil
}
