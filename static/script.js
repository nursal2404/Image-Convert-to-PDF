document.addEventListener('DOMContentLoaded', function() {
    const fileInput = document.getElementById('fileInput');
    const convertBtn = document.getElementById('convertBtn');
    const statusEl = document.getElementById('status');
    const dropZone = document.getElementById('dropZone');
    const fileList = document.getElementById('fileList');
    const previewContainer = document.getElementById('previewContainer');
    const btnText = document.getElementById('btnText');
    const btnLoader = document.getElementById('btnLoader');
    
    // Handle file selection
    fileInput.addEventListener('change', function() {
        updateFileList(this.files);
    });
    
    // Drag and drop functionality
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
    });
    
    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }
    
    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, highlight, false);
    });
    
    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, unhighlight, false);
    });
    
    function highlight() {
        dropZone.classList.add('dragover');
    }
    
    function unhighlight() {
        dropZone.classList.remove('dragover');
    }
    
    dropZone.addEventListener('drop', function(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        fileInput.files = files;
        updateFileList(files);
    });
    
    // Update file list preview
    function updateFileList(files) {
        fileList.innerHTML = '';
        
        if (files.length > 0) {
            previewContainer.style.display = 'block';
            convertBtn.disabled = false;
            
            for (let i = 0; i < files.length; i++) {
                const file = files[i];
                const listItem = document.createElement('li');
                listItem.className = 'preview-item';
                
                const fileName = document.createElement('span');
                fileName.className = 'file-name';
                fileName.textContent = file.name;
                
                const fileSize = document.createElement('span');
                fileSize.className = 'file-size';
                fileSize.textContent = formatFileSize(file.size);
                
                listItem.appendChild(fileName);
                listItem.appendChild(fileSize);
                fileList.appendChild(listItem);
            }
        } else {
            previewContainer.style.display = 'none';
            convertBtn.disabled = true;
        }
    }
    
    // Format file size
    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    }
    
    // Convert button click handler
convertBtn.addEventListener('click', async function() {
        if (fileInput.files.length === 0) return;
        
        // Show loading state
        btnText.textContent = 'Processing...';
        btnLoader.style.display = 'block';
        convertBtn.disabled = true;
        statusEl.textContent = '';
        statusEl.className = '';
        
        try {
            const formData = new FormData();
            for (let i = 0; i < fileInput.files.length; i++) {
                formData.append('images', fileInput.files[i]);
            }
            
            const response = await fetch('/convert', {
                method: 'POST',
                body: formData
            });
            
            const result = await response.json();
            
            if (!response.ok) {
                throw new Error(result.error || 'Conversion failed');
            }
            
            // Show success message with link to PDF
            statusEl.innerHTML = `PDF berhasil dibuat! <a href="/pdfs/${result.pdf}" target="_blank" style="color: var(--primary-color); text-decoration: underline;">Lihat PDF</a>`;
            statusEl.className = 'success-message';
            
        } catch (error) {
            statusEl.textContent = error.message;
            statusEl.className = 'error-message';
            console.error('Error:', error);
        } finally {
            // Reset button state
            btnText.textContent = 'Convert to PDF';
            btnLoader.style.display = 'none';
            convertBtn.disabled = false;
            
            // Reset file input
            fileInput.value = '';
            previewContainer.style.display = 'none';
        }
    });
});