package controllers

import (
	"UpblitIngestor/kafka"
	"net/http"

	"github.com/gin-gonic/gin"

	"UpblitIngestor/dto"
	"UpblitIngestor/services"
)

func TracesController(c *gin.Context) {
	// Extract IDs from context (set by middleware)
	appID := c.MustGet("application_id").(int64)
	projID := c.MustGet("project_id").(int64)

	// Parse request body
	var telemetry dto.Telemetry
	if err := c.ShouldBindJSON(&telemetry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
		})
		return
	}

	// Convert DTO → []TraceDocument
	docs := services.ToTraceDocuments(telemetry, projID, appID)
	var producer = kafka.NewProducer("localhost:9092", "traces-topic")
	err := producer.SendTraces(docs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to send to kafka",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":       "traces ingested successfully",
		"applicationId": appID,
		"projectId":     projID,
		"docs":          docs,
	})
}
