package services

import (
	"time"

	"UpblitIngestor/dto"
	"UpblitIngestor/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Convert Telemetry DTO → []Trace (one per trace)
func ToTraceDocuments(t dto.Telemetry, projectID, applicationID int64) []models.Trace {
	serverTime := time.Now().UTC()

	docs := make([]models.Trace, 0, len(t.Traces))

	for _, tr := range t.Traces {
		doc := models.Trace{
			ID: primitive.NewObjectID(),

			// timestamps
			ClientTimestamp: t.Timestamp,
			ServerTimestamp: serverTime,

			// tenancy
			ProjectID:     projectID,
			ApplicationID: applicationID,

			// trace fields
			RequestMethod:  tr.RequestMethod,
			RequestURL:     tr.RequestURL,
			ResponseStatus: tr.ResponseStatus,
			TraceID:        tr.TraceID,
			SpanID:         tr.SpanID,
			ParentSpanID:   tr.ParentSpanID,
			DurationMs:     tr.DurationMs,
			Timestamp:      tr.Timestamp,
		}

		docs = append(docs, doc)
	}

	return docs
}
