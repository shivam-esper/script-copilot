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

type GeminiRequest struct {
	Contents []struct {
		Role  string `json:"role"`
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		Temperature     float64 `json:"temperature"`
		TopK            int     `json:"topK"`
		TopP            float64 `json:"topP"`
		MaxOutputTokens int     `json:"maxOutputTokens"`
	} `json:"generationConfig"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func main() {
	// Get Gemini API key from environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is not set")
	}

	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // Adjust if your frontend runs elsewhere
	config.AllowMethods = []string{"POST", "GET", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"} // Added Authorization if needed by Gemini or your setup
	r.Use(cors.New(config))

	// Existing generate-script endpoint
	r.POST("/generate-script", handleGenerateScript(apiKey))

	// Script execution endpoint
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

		geminiReq := GeminiRequest{
			Contents: []struct {
				Role  string `json:"role"`
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			}{
				{
					Role: "user",
					Parts: []struct {
						Text string `json:"text"`
					}{
						{
							Text: fmt.Sprintf(`You are an expert in writing shell scripts for Linux systems.
Generate a practical, secure, and efficient shell script based on this request: %s

Please include:
1. Helpful comments explaining what the script does
2. Proper error handling
3. Input validation where appropriate

Format the response as a bash script within markdown code blocks, like this:
`+"```bash"+`
#!/bin/bash
# script content here
`+"```"+``, req.Prompt),
						},
					},
				},
			},
			GenerationConfig: struct {
				Temperature     float64 `json:"temperature"`
				TopK            int     `json:"topK"`
				TopP            float64 `json:"topP"`
				MaxOutputTokens int     `json:"maxOutputTokens"`
			}{
				Temperature:     0.7,
				TopK:            40,
				TopP:            0.95,
				MaxOutputTokens: 2048,
			},
		}

		jsonData, err := json.Marshal(geminiReq)
		if err != nil {
			log.Printf("Error marshaling request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		log.Printf("Request body: %s", string(jsonData))

		// *** FIX: Update the model name in the URL to a more current one ***
		// Using "gemini-1.5-flash-latest" as a common and recent model.
		// You might need to adjust this based on Google's current model offerings.
		apiUrl := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=%s", apiKey)
		// Note: Some newer models might use the v1beta endpoint.

		httpReq, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to AI"})
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body) // In Go 1.16+ you'd use io.ReadAll
		if err != nil {
			log.Printf("Error reading response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read AI response"})
			return
		}

		log.Printf("API Response Status: %s", resp.Status)
		log.Printf("API Response Body: %s", string(body))

		if resp.StatusCode != http.StatusOK {
			var geminiError struct {
				Error struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
					Status  string `json:"status"`
				} `json:"error"`
			}
			if json.Unmarshal(body, &geminiError) == nil && geminiError.Error.Message != "" {
				log.Printf("API error: Code %d, Status %s, Message: %s", geminiError.Error.Code, geminiError.Error.Status, geminiError.Error.Message)
				c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("AI API error: %s", geminiError.Error.Message)})
				return
			}
			c.JSON(resp.StatusCode, gin.H{"error": "AI API error"})
			return
		}

		var geminiResp GeminiResponse
		if err := json.Unmarshal(body, &geminiResp); err != nil {
			log.Printf("Error parsing response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse AI response"})
			return
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response from AI"})
			return
		}

		script := geminiResp.Candidates[0].Content.Parts[0].Text
		c.JSON(http.StatusOK, ScriptResponse{Script: script})
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
		// Using os.CreateTemp for better control over file permissions and naming
		tmpfile, err := os.CreateTemp("", "script-*.sh")
		if err != nil {
			log.Printf("Error creating temp file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary script file"})
			return
		}
		// Ensure the file is removed even if there are panics or early returns
		defer os.Remove(tmpfile.Name())

		// Write the script to the temp file
		if _, err := tmpfile.Write([]byte(req.Script)); err != nil {
			log.Printf("Error writing to temp file: %v", err)
			tmpfile.Close() // Attempt to close before returning
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write script to file"})
			return
		}

		// It's crucial to close the file before trying to execute it or change its permissions.
		if err := tmpfile.Close(); err != nil {
			log.Printf("Error closing temp file after writing: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close script file after writing"})
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
		// The script might return a non-zero exit code which is a valid execution but indicates an error in the script itself.
		// cmd.Run() will return an error in this case.
		if err != nil {
			log.Printf("Error executing script (or script returned non-zero exit code): %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
			c.JSON(http.StatusOK, ExecuteScriptResponse{ // Still OK from HTTP perspective, but script failed
				Success: false,
				Output:  stdout.String(), // Include stdout even on error
				Error:   fmt.Sprintf("Script execution failed: %v\n%s", err, stderr.String()),
			})
			return
		}

		log.Printf("Script executed successfully. Stdout: %s", stdout.String())
		c.JSON(http.StatusOK, ExecuteScriptResponse{
			Success: true,
			Output:  stdout.String(),
		})
	}
}
