document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('commandForm');
    const commandInput = document.getElementById('commandInput');
    const commandOutput = document.getElementById('commandOutput');
    const currentDirSpan = document.getElementById('currentDir');
    
    form.addEventListener('submit', function(e) {
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
    
    // Focus the input field on page load
    commandInput.focus();
});