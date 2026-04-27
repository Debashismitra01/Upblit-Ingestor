package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	supabaseURL = os.Getenv("SUPABASE_URL")
	apiKey      = os.Getenv("SUPABASE_API_KEY")
)

type APIClient struct {
	ID            int64  `json:"id"`
	APIKey        string `json:"api_key"`
	ProjectID     int64  `json:"project_id"`
	ApplicationID int64  `json:"application_id"`
}

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")

		if apiKey == "" {
			c.JSON(401, gin.H{"error": "missing api key"})
			c.Abort()
			return
		}

		client, err := GetAPIClientByKey(apiKey)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid api key"})
			c.Abort()
			return
		}

		c.Set("project_id", client.ProjectID)
		c.Set("application_id", client.ApplicationID)

		c.Next()
	}
}
func GetAPIClientByKey(key string) (*APIClient, error) {
	url := fmt.Sprintf("%s/rest/v1/api_client?api_key=eq.%s&limit=1", supabaseURL, key)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []APIClient
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("client not found")
	}

	return &result[0], nil
}
