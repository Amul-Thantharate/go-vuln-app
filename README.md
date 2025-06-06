# 🔍 Website Health Check Tool 🛠️

A Go-based web application for checking website health, executing system commands, and uploading files to AWS S3 buckets.

## 📋 Overview

This application provides a simple web interface that allows users to:
- ⚡ Execute system commands to check website health (ping, nslookup, etc.)
- 📤 Upload files to AWS S3 buckets
- 📂 Navigate through the file system using the command interface

## 🚀 Features

### Command Execution
- Execute system commands directly from the web interface
- View command output in real-time
- Navigate directories using the `cd` command
- Current working directory is displayed and maintained across commands

### S3 File Upload
- Upload files directly to specified AWS S3 buckets
- Automatic key generation for uploaded files
- Success/error feedback for upload operations

## 🔧 Technical Stack

- **Backend**: Go (Golang)
- **Frontend**: HTML, CSS, JavaScript
- **AWS Integration**: AWS SDK for Go v2
- **Server**: Native Go HTTP server

## 📦 Dependencies

- Go 1.x
- AWS SDK for Go v2
- AWS credentials configured on the host machine

## 🏗️ Project Structure

```
go-vuln-app/
├── main.go           # Main application code
├── static/           # Static web assets
│   ├── index.html    # Main HTML template
│   ├── script.js     # Frontend JavaScript
│   └── style.css     # CSS styling
└── README.md         # This documentation
```

## 🔌 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Serves the main application interface |
| `/execute` | POST | Executes system commands |
| `/upload-to-s3` | POST | Handles file uploads to S3 |
| `/static/*` | GET | Serves static assets |

## 🚦 Getting Started

### Prerequisites

1. Go installed on your system
2. AWS credentials configured (for S3 upload functionality)

### Running the Application

1. Clone the repository
2. Navigate to the project directory
3. Run the application:
   ```
   go run main.go
   ```
4. Open your browser and navigate to `http://localhost:8080`

## 🔐 Security Considerations

⚠️ **Warning**: This application allows execution of system commands and file uploads. In a production environment, implement proper authentication, authorization, and input validation to prevent security vulnerabilities.

## 🔄 Usage Examples

### Command Execution
- Check website connectivity: `ping example.com`
- Look up DNS records: `nslookup example.com`
- List directory contents: `ls -la`
- Change directory: `cd /path/to/directory`

### File Upload
1. Enter the name of your S3 bucket
2. Select a file to upload
3. Click "Upload to S3"
4. View the upload result message

## 📝 Notes

- The application maintains the working directory state between commands
- Special handling is provided for the `cd` command
- File uploads are limited to 10MB
- Unique keys are generated for S3 uploads based on filename and timestamp

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
