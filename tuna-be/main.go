package main

import (
	"log"

	"github.com/docker/docker/client"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	ModeDockerfile       = "dockerfile"
	ModeTemplate         = "template"
	ModeImage            = "image"
	CheckIntervalSeconds = 10
)

var (
	dockerClient *client.Client
	db           *gorm.DB
)

func main() {
	initDocker()
	initDB()

	if err := ensureTraefik(); err != nil {
		log.Fatal("Failed to ensure Traefik:", err)
	}

	initUpdateHandler()
	startHttpServer()
}

func initDocker() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal("Failed to init Docker client:", err)
	}
}

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("data/main.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to init SQLite:", err)
	}
	if err := db.AutoMigrate(&Deployment{}); err != nil {
		log.Fatal("Failed to migrate DB:", err)
	}
}

func startHttpServer() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://modpack-manager.octsrv.org", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("static/index.html")
	})

	g := r.Group("/api")
	g.GET("/deployments", listDeployments)
	g.POST("/deployment", createDeployment)
	g.DELETE("/deployment/:id", deleteDeployment)
	g.PUT("/deployment/:id/deploy", deployHandler)

	log.Fatal(r.Run(":8080"))
}
