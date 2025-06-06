package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
