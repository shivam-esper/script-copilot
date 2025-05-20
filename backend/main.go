package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type ScriptRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

type ScriptResponse struct {
	Script string `json:"script"`
}

type ExecuteScriptRequest struct {
	Script string `json:"script" binding:"required"`
}

type ExecuteScriptResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AnthropicRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system"`
	MaxTokens int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Id      string `json:"id"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model string `json:"model"`
	Role  string `json:"role"`
}

func main() {
	// Get Anthropic API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY") // keeping the same env var name for convenience
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"POST", "GET", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type"}
	r.Use(cors.New(config))

	// Existing generate-script endpoint
	r.POST("/generate-script", handleGenerateScript(apiKey))

	// New execute-script endpoint
	r.POST("/execute-script", handleExecuteScript())

	// Start server
	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func handleGenerateScript(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ScriptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Received prompt: %s", req.Prompt)

		// Create the Anthropic API request
		anthropicReq := AnthropicRequest{
			Model: "claude-3-opus-20240229",
			Messages: []Message{
				{
					Role:    "user",
					Content: req.Prompt,
				},
			},
			System: `You are an expert in writing shell scripts for Linux systems. 
					Generate practical, secure, and efficient shell scripts based on the user's requirements. 
					Include helpful comments explaining what the script does.
					Always include proper error handling and input validation where appropriate.`,
			MaxTokens: 4096,
		}

		jsonData, err := json.Marshal(anthropicReq)
		if err != nil {
			log.Printf("Error marshaling request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		log.Printf("Request body: %s", string(jsonData))

		// Create HTTP request to Anthropic API
		httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Set headers
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")

		// Send request
		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to AI"})
			return
		}
		defer resp.Body.Close()

		// Read response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Log the raw response for debugging
		log.Printf("Raw API Response: %s", string(body))

		if resp.StatusCode != http.StatusOK {
			log.Printf("API error (status %d): %s", resp.StatusCode, string(body))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("API error: %s", string(body))})
			return
		}

		// Parse response
		var anthropicResp AnthropicResponse
		if err := json.Unmarshal(body, &anthropicResp); err != nil {
			log.Printf("Error parsing response: %v\nResponse body: %s", err, string(body))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		if len(anthropicResp.Content) == 0 {
			log.Printf("Empty response content from AI")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No content in AI response"})
			return
		}

		// Get the text content from the response
		var scriptContent string
		for _, content := range anthropicResp.Content {
			if content.Type == "text" {
				scriptContent = content.Text
				break
			}
		}

		if scriptContent == "" {
			log.Printf("No text content found in response")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No text content in AI response"})
			return
		}

		c.JSON(http.StatusOK, ScriptResponse{
			Script: scriptContent,
		})
	}
}

func handleExecuteScript() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ExecuteScriptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Received script to execute: %s", req.Script)

		// Create a temporary script file
		tmpfile, err := ioutil.TempFile("", "script-*.sh")
		if err != nil {
			log.Printf("Error creating temp file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary script file"})
			return
		}
		defer os.Remove(tmpfile.Name())

		// Write the script to the temp file
		if _, err := tmpfile.Write([]byte(req.Script)); err != nil {
			log.Printf("Error writing to temp file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write script to file"})
			return
		}
		if err := tmpfile.Close(); err != nil {
			log.Printf("Error closing temp file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close script file"})
			return
		}

		// Make the script executable
		if err := os.Chmod(tmpfile.Name(), 0755); err != nil {
			log.Printf("Error making script executable: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make script executable"})
			return
		}

		// Execute the script
		cmd := exec.Command("bash", tmpfile.Name())
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			log.Printf("Error executing script: %v\nStderr: %s", err, stderr.String())
			c.JSON(http.StatusOK, ExecuteScriptResponse{
				Success: false,
				Error:   fmt.Sprintf("Error executing script: %v\n%s", err, stderr.String()),
			})
			return
		}

		c.JSON(http.StatusOK, ExecuteScriptResponse{
			Success: true,
			Output:  stdout.String(),
		})
	}
}
