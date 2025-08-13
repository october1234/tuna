package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func listDeployments(c *gin.Context) {
	var deployments []Deployment
	if err := db.Find(&deployments).Error; err != nil {
		c.JSON(500, gin.H{"error": "DB error"})
		return
	}
	c.JSON(200, deployments)
}

func createDeployment(c *gin.Context) {
	var req Deployment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	fmt.Println(req)

	err := validateDeployment(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid deployment: " + err.Error()})
		return
	}

	if err := db.Create(&req).Error; err != nil {
		c.JSON(500, gin.H{"error": "DB error"})
		return
	}
}

func deleteDeployment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "ID is required"})
		return
	}

	if err := db.Delete(&Deployment{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "DB error"})
		return
	}
	c.Status(http.StatusNotImplemented)
}

func deployHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "ID is required"})
		return
	}

	var deployment Deployment
	if err := db.First(&deployment, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Deployment not found"})
		return
	}

	switch deployment.ModeData.Mode {
	case ModeImage:
		if err := deployFromImage(&deployment); err != nil {
			c.JSON(500, gin.H{"error": "Failed to deploy from image: " + err.Error()})
			return
		}
	case ModeTemplate, ModeDockerfile:
		if err := deployFromGit(&deployment); err != nil {
			c.JSON(500, gin.H{"error": "Failed to deploy from git: " + err.Error()})
			return
		}
	default:
		c.JSON(419, gin.H{"error": "Unsupported deployment mode: " + deployment.ModeData.Mode})
		return
	}

	c.Status(204)
}
