package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var regex = regexp.MustCompile(`^[a-z0-9][a-z0-9_]{2,}$`)

func createTar(dir string) (*os.File, error) {
	tarFile, err := os.CreateTemp("", "build-context-*.tar")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("tar", "-cf", tarFile.Name(), "-C", dir, ".")
	if err := cmd.Run(); err != nil {
		tarFile.Close()
		os.Remove(tarFile.Name())
		return nil, err
	}

	return tarFile, nil
}

func validateDeploymentGitData(gitData DeploymentGitData, mode string) error {
	if gitData.Repository == "" {
		return fmt.Errorf("Git repository is required for " + mode + " mode")
	}
	if gitData.Branch == "" {
		return fmt.Errorf("Git branch is required for " + mode + " mode")
	}
	return nil
}

func validateDeploymentID(id string) bool {
	return regex.MatchString(id)
}

func validateDeployment(deployment *Deployment) error {
	// validate id to be lowercase and underscore
	if !validateDeploymentID(deployment.ID) {
		return fmt.Errorf("invalid deployment id: %s", deployment.ID)
	}

	switch deployment.ModeData.Mode {
	case ModeDockerfile:
		if err := validateDeploymentGitData(deployment.ModeData.GitData, deployment.ModeData.Mode); err != nil {
			return err
		}
		if deployment.ModeData.DockerFile == "" {
			return fmt.Errorf("dockerfile mode requires a dockerfile path")
		}
		if strings.HasPrefix(deployment.ModeData.DockerFile, "/") {
			return fmt.Errorf("dockerfile path must be relative, not absolute")
		}
		if strings.Contains(deployment.ModeData.DockerFile, "..") {
			return fmt.Errorf("dockerfile path must not contain '..'")
		}
		return nil
	case ModeTemplate:
		if err := validateDeploymentGitData(deployment.ModeData.GitData, deployment.ModeData.Mode); err != nil {
			return err
		}
		if deployment.ModeData.Template == "" {
			return fmt.Errorf("template mode requires a template")
		}
		return nil
	case ModeImage:
		if deployment.ModeData.Image == "" {
			return fmt.Errorf("image mode requires an image name")
		}
		return nil
	default:
		return fmt.Errorf("invalid deployment mode: %s", deployment.ModeData.Mode)
	}
}

func formatDockerResourceName(deploymentId string) string {
	return "tuna-deployment-" + deploymentId
}

func formatBuildImageName(deploymentId string) string {
	return "tuna-deployment-build-" + deploymentId + ":latest"
}
