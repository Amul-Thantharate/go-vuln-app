document.addEventListener('DOMContentLoaded', function() {
    const commandForm = document.getElementById('commandForm');
    const commandInput = document.getElementById('commandInput');
    const commandOutput = document.getElementById('commandOutput');
    const currentDirSpan = document.getElementById('currentDir');
    const uploadForm = document.getElementById('uploadForm');
    const uploadResult = document.getElementById('uploadResult');
    
    commandForm.addEventListener('submit', function(e) {
        e.preventDefault();
        const command = commandInput.value.trim();
        
        if (!command) {
            alert('Please enter a command');
            return;
        }
        
        // Show loading indicator
        commandOutput.textContent = 'Executing command...';
        
        fetch('/execute', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: `command=${encodeURIComponent(command)}`
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Server responded with status: ${response.status}`);
            }
            return response.text();
        })
        .then(text => {
            try {
                // Try to parse as JSON
                const data = JSON.parse(text);
                
                // Format the output
                let outputText = '';
                
                if (data.error) {
                    outputText += `Error: ${data.error}\n\n`;
                }
                
                if (data.commandInfo) {
                    outputText += `${data.commandInfo}\n`;
                }
                
                if (data.output) {
                    outputText += data.output;
                }
                
                // Update the current directory if provided
                if (data.currentDir) {
                    currentDirSpan.textContent = data.currentDir;
                }
                
                // Update the output
                if (outputText.trim() === '') {
                    commandOutput.textContent = 'Command executed with no output';
                } else {
                    commandOutput.textContent = outputText;
                }
            } catch (err) {
                // If JSON parsing fails, display the raw text
                console.error('Error parsing JSON:', err);
                commandOutput.textContent = text || 'No response from server';
            }
            
            // Scroll to the output
            commandOutput.scrollIntoView({ behavior: 'smooth' });
        })
        .catch(error => {
            commandOutput.textContent = `Error: ${error.message}`;
            console.error('Fetch error:', error);
        });
    });
    
    // Handle S3 file upload
    if (uploadForm) {
        uploadForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const bucketName = document.getElementById('bucketInput').value.trim();
            const fileInput = document.getElementById('fileInput');
            
            if (!bucketName) {
                alert('Please enter a bucket name');
                return;
            }
            
            if (!fileInput.files || fileInput.files.length === 0) {
                alert('Please select a file to upload');
                return;
            }
            
            // Show loading indicator
            uploadResult.textContent = 'Uploading file to S3...';
            uploadResult.className = 'upload-result uploading';
            
            const formData = new FormData();
            formData.append('bucket', bucketName);
            formData.append('file', fileInput.files[0]);
            
            fetch('/upload-to-s3', {
                method: 'POST',
                body: formData
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Server responded with status: ${response.status}`);
                }
                return response.json();
            })
            .then(data => {
                if (data.success) {
                    uploadResult.textContent = `Success: ${data.message}. File key: ${data.key}`;
                    uploadResult.className = 'upload-result success';
                } else {
                    uploadResult.textContent = `Error: ${data.message}`;
                    uploadResult.className = 'upload-result error';
                }
            })
            .catch(error => {
                uploadResult.textContent = `Error: ${error.message}`;
                uploadResult.className = 'upload-result error';
                console.error('Upload error:', error);
            });
        });
    }
    
    // Focus the input field on page load
    commandInput.focus();
});