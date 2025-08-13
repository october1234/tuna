package main

type Deployment struct {
	ID          string             `json:"id" gorm:"primaryKey" binding:"required"`
	Name        string             `json:"name" gorm:"not null" binding:"required"`
	Description string             `json:"description"`
	Ingress     []string           `json:"ingress" gorm:"serializer:json;default:'[]'"`
	ModeData    DeploymentModeData `json:"mode_data" gorm:"embedded;embeddedPrefix:modedata_" binding:"required"`
	Env         H                  `json:"env" gorm:"serializer:json;default:'{}'"`
	Labels      H                  `json:"labels" gorm:"serializer:json;default:'{}'"`
	Volumes     H                  `json:"volumes" gorm:"serializer:json;default:'{}'"`
	Disabled    bool               `json:"disabled" gorm:"default:false"`
}
type DeploymentModeData struct {
	Mode       string            `json:"mode" gorm:"not null" binding:"required"`
	GitData    DeploymentGitData `json:"git_data" gorm:"embedded;embeddedPrefix:gitdata_"`
	DockerFile string            `json:"dockerfile"`
	Template   string            `json:"template"`
	Image      string            `json:"image"`
}
type DeploymentGitData struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
}

type H map[string]string
