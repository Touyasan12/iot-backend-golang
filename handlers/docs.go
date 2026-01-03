package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// ServeOpenAPIYAML serves the OpenAPI YAML file
func ServeOpenAPIYAML(c *gin.Context) {
	c.Header("Content-Type", "application/x-yaml; charset=utf-8")
	c.File("./openapi.yaml")
}

// ServeOpenAPIJSON converts and serves OpenAPI spec as JSON
func ServeOpenAPIJSON(c *gin.Context) {
	// Read YAML file
	yamlData, err := os.ReadFile("./openapi.yaml")
	if err != nil {
		log.Printf("Error reading openapi.yaml: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read OpenAPI specification"})
		return
	}

	// Parse YAML
	var data interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		log.Printf("Error parsing YAML: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse OpenAPI specification"})
		return
	}

	// Serve as JSON
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, data)
}

