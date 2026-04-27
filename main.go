package main

import (
	"UpblitIngestor/config"
	"UpblitIngestor/controllers"
	"UpblitIngestor/middleware"
	"UpblitIngestor/worker"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	r := gin.Default()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("MONGO_URI =", os.Getenv("MONGO_URI"))
	client := config.ConnectMongo()

	// Select DB + Collection
	db := client.Database("observability")

	metricsCol := db.Collection("metrics")
	traceCol := db.Collection("traces")

	w := worker.NewWorker(
		"localhost:9092",
		"traces-topic",
		"metrics-group",
		metricsCol,
		traceCol,
	)

	go w.Start()
	r.Use(middleware.APIKeyMiddleware())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, Gin!",
		})
	})
	r.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{
			"user": name,
		})
	})
	r.POST("/ingest/traces", controllers.TracesController)

	r.Run(":9000")
}
