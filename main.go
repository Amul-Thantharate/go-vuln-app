package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PageData struct {
	CommandOutput string
	CurrentDir    string
}

type CommandResponse struct {
	Output      string `json:"output"`
	Error       string `json:"error,omitempty"`
	CurrentDir  string `json:"currentDir"`
	CommandInfo string `json:"commandInfo,omitempty"`
}

type S3UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Key     string `json:"key,omitempty"`
}

var currentWorkingDir string

func init() {
	var err error
	currentWorkingDir, err = os.Getwd()
	if err != nil {
		currentWorkingDir = "/"
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/execute", executeCommandHandler)
	http.HandleFunc("/upload-to-s3", uploadToS3Handler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		CommandOutput: "",
		CurrentDir:    currentWorkingDir,
	}

	tmpl.Execute(w, data)
}

func executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	command := r.FormValue("command")
	if command == "" {
		http.Error(w, "Command cannot be empty", http.StatusBadRequest)
		return
	}

	response := handleCommand(command)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func handleCommand(command string) CommandResponse {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return CommandResponse{
			Error: "Empty command",
		}
	}

	// Handle cd command specially
	if parts[0] == "cd" {
		return handleCdCommand(parts)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// Set the working directory for the command
	cmd.Dir = currentWorkingDir

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	response := CommandResponse{
		Output:     outputStr,
		CurrentDir: currentWorkingDir,
	}

	// Add command-specific information
	switch parts[0] {
	case "wget", "curl":
		if err == nil && outputStr == "" {
			response.CommandInfo = "Download command executed. Check the current directory for downloaded files."
		}
	}

	if err != nil {
		response.Error = fmt.Sprintf("Error executing command: %v", err)
	} else if outputStr == "" {
		response.CommandInfo = "Command executed successfully with no output"
	}

	return response
}

func handleCdCommand(parts []string) CommandResponse {
	if len(parts) < 2 {
		return CommandResponse{
			Error:      "cd requires a directory argument",
			CurrentDir: currentWorkingDir,
		}
	}

	targetDir := parts[1]
	if !filepath.IsAbs(targetDir) {
		targetDir = filepath.Join(currentWorkingDir, targetDir)
	}

	// Check if directory exists and is accessible
	if info, err := os.Stat(targetDir); err != nil || !info.IsDir() {
		return CommandResponse{
			Error:      fmt.Sprintf("Cannot change to directory %s: %v", targetDir, err),
			CurrentDir: currentWorkingDir,
		}
	}

	currentWorkingDir = targetDir
	return CommandResponse{
		CommandInfo: fmt.Sprintf("Changed directory to: %s", targetDir),
		CurrentDir:  currentWorkingDir,
	}
}

// uploadToS3Handler handles file uploads to S3
func uploadToS3Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form with a 10MB limit
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithJSON(w, S3UploadResponse{
			Success: false,
			Message: "Failed to parse form: " + err.Error(),
		})
		return
	}

	// Get the bucket name from the form
	bucketName := r.FormValue("bucket")
	if bucketName == "" {
		respondWithJSON(w, S3UploadResponse{
			Success: false,
			Message: "Bucket name is required",
		})
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithJSON(w, S3UploadResponse{
			Success: false,
			Message: "Failed to get file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Generate a unique key for the file
	key := fmt.Sprintf("%s-%d", header.Filename, time.Now().Unix())

	// Upload the file to S3
	result, err := uploadFileToS3(r.Context(), file, bucketName, key)
	if err != nil {
		respondWithJSON(w, S3UploadResponse{
			Success: false,
			Message: "Failed to upload to S3: " + err.Error(),
		})
		return
	}

	respondWithJSON(w, S3UploadResponse{
		Success: true,
		Message: "File uploaded successfully to S3",
		Key:     result,
	})
}

// uploadFileToS3 uploads a file to an S3 bucket
func uploadFileToS3(ctx context.Context, file io.Reader, bucket, key string) (string, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)

	// Upload the file to S3
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return key, nil
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
