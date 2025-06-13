// Alpine.js components for minimal client-side interactions
// Only used for pure UI interactions that need immediate feedback

// Image Upload Component
function imageUploadComponent() {
    return {
        uploading: false,
        progress: 0,
        
        // Trigger file input click
        triggerUpload() {
            this.$refs.fileInput.click();
        },
        
        // Handle file selection and upload
        async handleFileUpload(stampId) {
            const file = this.$refs.fileInput.files[0];
            if (!file) return;
            
            // Client-side validation
            if (!file.type.startsWith('image/')) {
                alert('Please select an image file.');
                return;
            }
            
            if (file.size > 5 * 1024 * 1024) {
                alert('File size must be less than 5MB.');
                return;
            }
            
            this.uploading = true;
            this.progress = 0;
            
            try {
                const formData = new FormData();
                formData.append('image', file);
                
                // Create XMLHttpRequest for progress tracking
                const xhr = new XMLHttpRequest();
                
                // Track upload progress
                xhr.upload.addEventListener('progress', (e) => {
                    if (e.lengthComputable) {
                        this.progress = Math.round((e.loaded / e.total) * 100);
                    }
                });
                
                // Handle completion
                xhr.addEventListener('load', () => {
                    if (xhr.status === 200) {
                        const response = JSON.parse(xhr.responseText);
                        this.updateImageDisplay(response.image_url, stampId);
                        this.$refs.fileInput.value = '';
                    } else {
                        throw new Error('Upload failed');
                    }
                    this.uploading = false;
                });
                
                xhr.addEventListener('error', () => {
                    alert('Upload failed. Please try again.');
                    this.uploading = false;
                });
                
                xhr.open('POST', `/api/stamps/${stampId}/upload-image`);
                xhr.send(formData);
                
            } catch (error) {
                console.error('Upload error:', error);
                alert('Upload failed. Please try again.');
                this.uploading = false;
            }
        },
        
        // Update image display after successful upload
        updateImageDisplay(imageUrl, stampId) {
            // Update the image in the UI
            const imageContainer = document.querySelector('.stamp-detail-image-container');
            if (imageContainer) {
                const newImage = document.createElement('img');
                newImage.src = imageUrl + '?t=' + Date.now();
                newImage.alt = 'Stamp image';
                newImage.className = 'stamp-detail-img';
                newImage.id = 'stamp-image';
                
                newImage.onerror = () => {
                    newImage.style.display = 'none';
                    this.showImagePlaceholder();
                };
                
                imageContainer.innerHTML = '';
                imageContainer.appendChild(newImage);
                
                // Update controls
                this.updateImageControls(stampId);
            }
        },
        
        // Show placeholder when image fails to load
        showImagePlaceholder() {
            const imageContainer = document.querySelector('.stamp-detail-image-container');
            if (imageContainer) {
                imageContainer.innerHTML = `
                    <div class="stamp-detail-placeholder" id="image-placeholder">
                        <i class="bi bi-image" style="font-size: 3rem; opacity: 0.3;"></i>
                        <p class="text-muted mt-2">Image not available</p>
                    </div>
                `;
            }
        },
        
        // Update image controls after upload
        updateImageControls(stampId) {
            const controls = document.querySelector('.image-controls');
            if (controls) {
                controls.innerHTML = `
                    <button class="btn btn-sm btn-outline-secondary me-2" onclick="triggerImageUpload()">
                        <i class="bi bi-camera"></i> Change Image
                    </button>
                    <button class="btn btn-sm btn-outline-danger" 
                            hx-put="/api/stamps/${stampId}" 
                            hx-vals='{"image_url": ""}'
                            hx-confirm="Remove this image?"
                            hx-target="#stamp-view-content"
                            hx-indicator="#loading-spinner">
                        <i class="bi bi-trash"></i> Remove
                    </button>
                `;
            }
        }
    };
}

// Modal Component for general-purpose modals
function modalComponent() {
    return {
        open: false,
        
        show() {
            this.open = true;
        },
        
        hide() {
            this.open = false;
        },
        
        // Close on escape key
        handleKeydown(event) {
            if (event.key === 'Escape') {
                this.hide();
            }
        }
    };
}

// Form Validation Component for real-time feedback
function formValidationComponent() {
    return {
        valid: false,
        errors: {},
        
        // Validate field in real-time
        validateField(field, value, rules) {
            this.errors[field] = [];
            
            if (rules.required && (!value || value.trim() === '')) {
                this.errors[field].push(`${field} is required`);
            }
            
            if (rules.minLength && value.length < rules.minLength) {
                this.errors[field].push(`${field} must be at least ${rules.minLength} characters`);
            }
            
            if (rules.maxLength && value.length > rules.maxLength) {
                this.errors[field].push(`${field} must not exceed ${rules.maxLength} characters`);
            }
            
            this.updateValidationState();
        },
        
        // Update overall form validation state
        updateValidationState() {
            this.valid = Object.values(this.errors).every(errorArray => errorArray.length === 0);
        },
        
        // Get CSS classes for field validation state
        getFieldClasses(field) {
            if (!this.errors[field]) return '';
            return this.errors[field].length > 0 ? 'is-invalid' : 'is-valid';
        }
    };
}

// Make components globally available
window.imageUploadComponent = imageUploadComponent;
window.modalComponent = modalComponent;
window.formValidationComponent = formValidationComponent;