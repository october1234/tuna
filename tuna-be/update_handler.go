package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

var imageDigests = make(H)
var gitHashes = make(H)

func initUpdateHandler() {
	go func() {
		for range time.Tick(CheckIntervalSeconds * time.Second) {
			checkImageUpdates()
			checkGitUpdates()
		}
	}()
}

func checkImageUpdates() {
	var deployments []Deployment
	if err := db.Where("modedata_mode = 'image' AND disabled = false").Find(&deployments).Error; err != nil {
		fmt.Print("failed to fetch deployments: %w", err)
		return
	}

	for _, dep := range deployments {
		image := dep.ModeData.Image
		if image == "" {
			fmt.Printf("Deployment %s has no image set, skipping\n", dep.Name)
			continue
		}

		if err := checkImage(image, &dep); err != nil {
			fmt.Printf("Error checking image %s: %v\n", image, err)
			continue
		}
	}
}

func checkImage(image string, dep *Deployment) error {
	digest, err := getRemoteDigest(image)
	if err != nil {
		return fmt.Errorf("error checking digest: %v", err)
	}

	currentDigest, exists := imageDigests[image]
	if !exists {
		imageDigests[image] = digest
		fmt.Printf("Initial digest for %s: %s\n", image, digest)
		return nil
	}

	if currentDigest != "" && digest != currentDigest {
		fmt.Println("Image update detected!")
		deployFromImage(dep)
	}

	return nil
}

func getRemoteDigest(image string) (string, error) {
	distInspect, err := dockerClient.DistributionInspect(context.Background(), image, "")
	if err != nil {
		return "", err
	}
	return distInspect.Descriptor.Digest.String(), nil
}

func checkGitUpdates() {
	var deployments []Deployment
	if err := db.Where("(modedata_mode = 'dockerfile' OR modedata_mode = 'template') AND disabled = false").Find(&deployments).Error; err != nil {
		fmt.Printf("failed to fetch git deployments: %v\n", err)
		return
	}

	for _, dep := range deployments {
		if dep.ModeData.GitData.Repository == "" {
			fmt.Printf("Deployment %s has no git URL set, skipping\n", dep.Name)
			continue
		}

		if err := checkGit(&dep); err != nil {
			fmt.Printf("Error checking git for %s: %v\n", dep.Name, err)
			continue
		}
	}
}

func checkGit(dep *Deployment) error {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{dep.ModeData.GitData.Repository},
	})
	if remote == nil {
		return fmt.Errorf("error creating remote")
	}
	err := remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%v:refs/remotes/origin/%v", dep.ModeData.GitData.Branch, dep.ModeData.GitData.Branch)),
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch: %v", err)
	}
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list refs: %v", err)
	}
	for _, ref := range refs {
		if ref.Name().String() == "refs/heads/"+dep.ModeData.GitData.Branch {
			hash := ref.Hash().String()
			currentHash, exists := gitHashes[dep.Name]
			if !exists {
				gitHashes[dep.Name] = hash
				fmt.Printf("Initial hash for %s: %s\n", dep.Name, hash)
				return nil
			}

			if currentHash != "" && hash != currentHash {
				fmt.Println("Git update detected!")
				gitHashes[dep.Name] = hash
				err := deployFromGit(dep)
				if err != nil {
					return fmt.Errorf("failed to deploy from git: %v", err)
				}
			}

			return nil
		}
	}
	return fmt.Errorf("failed to fetch git %s", dep.ModeData.GitData.Repository)
}
