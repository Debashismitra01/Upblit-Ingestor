package dto

import (
	"time"
)

type Log struct {
	TraceID         string    `bson:"trace_id" json:"traceId"`
	Level           string    `bson:"level" json:"level"`
	Message         string    `bson:"message" json:"message"`
	Timestamp       time.Time `bson:"timestamp" json:"timestamp"`
	ClientTimestamp time.Time `json:"clientTimestamp" bson:"clientTimestamp"`
}
