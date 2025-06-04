// Global variable to store available boxes
let availableBoxes = [];

// Load available boxes when page loads
document.addEventListener('DOMContentLoaded', function() {
    loadAvailableBoxes();
});

// Load available storage boxes for dropdowns
function loadAvailableBoxes() {
    fetch('/api/boxes')
        .then(response => response.json())
        .then(boxes => {
            availableBoxes = boxes;
            populateBoxDropdowns();
        })
        .catch(error => {
            console.error('Error loading boxes:', error);
        });
}

// Populate all box dropdowns with current data
function populateBoxDropdowns() {
    const boxSelects = document.querySelectorAll('select[data-field="box_id"]');
    boxSelects.forEach(select => {
        const currentValue = select.value;
        
        // Clear existing options except "Unboxed"
        select.innerHTML = '<option value="">Unboxed</option>';
        
        // Add all available boxes
        availableBoxes.forEach(box => {
            const option = document.createElement('option');
            option.value = box.id;
            option.textContent = box.name;
            if (box.id === currentValue) {
                option.selected = true;
            }
            select.appendChild(option);
        });
    });
}

// Save instance field changes
function saveInstanceField(element) {
    const field = element.dataset.field;
    const instanceId = element.dataset.instanceId;
    let value;
    
    console.log('Saving instance field:', field, 'for instance:', instanceId);
    
    // Add visual feedback
    element.classList.add('saving');
    
    if (element.type === 'number') {
        value = parseInt(element.value) || 0;
    } else {
        value = element.value;
    }
    
    console.log('Field value:', value);
    
    const data = {};
    data[field] = value;
    
    console.log('Sending data:', JSON.stringify(data));
    
    // Update the instance via API
    fetch(`/api/instances/${instanceId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
    })
    .then(response => {
        console.log('Response status:', response.status);
        
        if (response.status === 204) {
            // Instance was deleted (quantity set to 0)
            removeInstanceRow(instanceId);
            return null;
        }
        
        if (!response.ok) {
            return response.text().then(text => {
                console.error('Error response body:', text);
                throw new Error(`HTTP ${response.status}: ${text}`);
            });
        }
        return response.json();
    })
    .then(data => {
        if (data) {
            console.log('Success response:', data);
        }
        
        // Visual feedback for successful save
        element.classList.remove('saving');
        element.classList.add('saved');
        setTimeout(() => {
            element.classList.remove('saved');
        }, 1000);
    })
    .catch(error => {
        console.error('Error updating instance:', error);
        
        // Visual feedback for error
        element.classList.remove('saving');
        element.classList.add('error');
        setTimeout(() => {
            element.classList.remove('error');
        }, 2000);
        
        alert(`Failed to save changes: ${error.message}`);
    });
}

// Adjust instance quantity using +/- buttons
function adjustInstanceQuantity(instanceId, delta) {
    const input = document.querySelector(`input[data-instance-id="${instanceId}"][data-field="quantity"]`);
    const currentValue = parseInt(input.value) || 0;
    const newValue = Math.max(0, currentValue + delta);
    input.value = newValue;
    saveInstanceField(input);
}

// Delete an entire instance group
function deleteInstance(instanceId) {
    if (confirm('Are you sure you want to delete this group of copies?')) {
        fetch(`/api/instances/${instanceId}`, {
            method: 'DELETE'
        })
        .then(response => {
            if (response.ok) {
                removeInstanceRow(instanceId);
            } else {
                return response.text().then(text => {
                    throw new Error(`Failed to delete: ${text}`);
                });
            }
        })
        .catch(error => {
            console.error('Error deleting instance:', error);
            alert(`Failed to delete instance: ${error.message}`);
        });
    }
}

// Remove instance row from the table
function removeInstanceRow(instanceId) {
    const row = document.querySelector(`tr[data-instance-id="${instanceId}"]`);
    if (row) {
        row.remove();
        
        // Check if table is now empty
        const tableBody = document.getElementById('copies-table-body');
        if (tableBody.children.length === 0) {
            // Show empty state
            showEmptyState();
        }
        
        // Update the count in the header
        updateInstanceCount();
    }
}

// Show empty state when no instances exist
function showEmptyState() {
    const copiesSection = document.querySelector('.your-copies-section');
    const stampId = new URLSearchParams(window.location.search).get('id') || 
                    document.querySelector('[data-stamp-id]')?.dataset.stampId;
    
    copiesSection.innerHTML = `
        <div class="section-header">
            <h4 class="section-title">
                <i class="bi bi-collection"></i> Your Copies
                <span class="total-count">(0 groups)</span>
            </h4>
            <button class="btn btn-sm btn-primary add-copy-btn" onclick="addNewCopy('${stampId}')">
                <i class="bi bi-plus-circle"></i> Add Copies
            </button>
        </div>
        <div class="no-copies-message">
            <div class="empty-state">
                <i class="bi bi-inbox" style="font-size: 3rem; opacity: 0.3;"></i>
                <h5>No copies in your collection</h5>
                <p class="text-muted">Add some copies of this stamp to track your inventory.</p>
                <button class="btn btn-primary" onclick="addNewCopy('${stampId}')">
                    <i class="bi bi-plus-circle"></i> Add Your First Copy
                </button>
            </div>
        </div>
    `;
}

// Update the instance count in the header
function updateInstanceCount() {
    const rows = document.querySelectorAll('#copies-table-body tr').length;
    const countSpan = document.querySelector('.total-count');
    if (countSpan) {
        countSpan.textContent = `(${rows} ${rows === 1 ? 'group' : 'groups'})`;
    }
}

// Add new copy group
function addNewCopy(stampId) {
    // Simple prompt-based approach for now
    const condition = prompt('Enter condition (Mint, Used, etc.):') || '';
    const quantityStr = prompt('How many copies?') || '1';
    const quantity = parseInt(quantityStr) || 1;
    
    if (quantity < 1) {
        alert('Quantity must be at least 1');
        return;
    }
    
    const newInstance = {
        condition: condition,
        box_id: '', // Unboxed by default
        quantity: quantity
    };
    
    fetch(`/api/stamps/${stampId}/instances`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(newInstance)
    })
    .then(response => {
        if (!response.ok) {
            return response.text().then(text => {
                throw new Error(text);
            });
        }
        return response.json();
    })
    .then(instance => {
        // Reload the page to show the new instance
        const currentStampId = stampId;
        htmx.ajax('GET', `/views/stamps/detail/${currentStampId}`, '#stamp-view-content');
    })
    .catch(error => {
        console.error('Error adding instance:', error);
        if (error.message.includes('UNIQUE constraint')) {
            alert('You already have copies with this condition and box combination. Try editing the existing group instead.');
        } else {
            alert(`Failed to add copies: ${error.message}`);
        }
    });
}

function saveField(element) {
    const field = element.dataset.field;
    const stampId = element.dataset.stampId;
    let value;
    
    console.log('Saving field:', field, 'for stamp:', stampId);
    
    if (element.type === 'checkbox') {
        value = element.checked;
    } else if (element.type === 'number') {
        value = parseInt(element.value) || 0;
    } else {
        value = element.value || element.textContent.trim();
    }
    
    console.log('Field value:', value);
    
    const data = {};
    data[field] = value;
    
    console.log('Sending data:', JSON.stringify(data));
    
    // Update the stamp via API
    fetch(`/api/stamps/${stampId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
    })
    .then(response => {
        console.log('Response status:', response.status);
        console.log('Response headers:', response.headers);
        
        if (!response.ok) {
            return response.text().then(text => {
                console.error('Error response body:', text);
                throw new Error(`HTTP ${response.status}: ${text}`);
            });
        }
        return response.json();
    })
    .then(data => {
        console.log('Success response:', data);
        // Visual feedback for successful save
        element.style.backgroundColor = '#d4edda';
        setTimeout(() => {
            element.style.backgroundColor = '';
        }, 1000);
    })
    .catch(error => {
        console.error('Error updating stamp:', error);
        alert(`Failed to save changes: ${error.message}`);
        // Visual feedback for error
        element.style.backgroundColor = '#f8d7da';
        setTimeout(() => {
            element.style.backgroundColor = '';
        }, 2000);
    });
}

// Handle Enter key for contenteditable and input fields
function handleEnterKey(event, element) {
    if (event.key === 'Enter') {
        event.preventDefault();
        element.blur();
    }
}

// Quantity adjustment
//function adjustQuantity(stampId, delta) {
//    const input = document.querySelector(`input[data-stamp-id="${stampId}"][data-field="quantity"]`);
//    const currentValue = parseInt(input.value) || 0;
//    const newValue = Math.max(0, currentValue + delta);
//    input.value = newValue;
//    saveField(input);
//}

// Tag management functions
function removeTag(stampId, tagName) {
    // Get current tags from the stamp
    const currentTags = Array.from(document.querySelectorAll('.editable-tag'))
        .map(tag => tag.textContent.replace('×', '').trim())
        .filter(tag => tag !== tagName);
    
    console.log('Removing tag:', tagName, 'Current tags after removal:', currentTags);
    
    fetch(`/api/stamps/${stampId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ tags: currentTags })
    })
    .then(response => {
        if (response.ok) {
            // Remove the tag from the UI
            document.querySelector(`[onclick="removeTag('${stampId}', '${tagName}')"]`).parentElement.remove();
        } else {
            throw new Error('Failed to remove tag');
        }
    })
    .catch(error => {
        console.error('Error removing tag:', error);
        alert('Failed to remove tag');
    });
}

function addNewTag(stampId) {
    const tagName = prompt('Enter tag name:');
    if (tagName && tagName.trim()) {
        // Get current tags
        const currentTags = Array.from(document.querySelectorAll('.editable-tag'))
            .map(tag => tag.textContent.replace('×', '').trim());
        
        currentTags.push(tagName.trim());
        
        console.log('Adding tag:', tagName, 'All tags:', currentTags);
        
        fetch(`/api/stamps/${stampId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ tags: currentTags })
        })
        .then(response => {
            if (response.ok) {
                // Reload the page to show the new tag
                htmx.ajax('GET', `/views/stamps/detail/${stampId}`, '#stamp-view-content');
            } else {
                throw new Error('Failed to add tag');
            }
        })
        .catch(error => {
            console.error('Error adding tag:', error);
            alert('Failed to add tag');
        });
    }
}

// Image management functions
function triggerImageUpload() {
    document.getElementById('imageUpload').click();
}

// Handle file upload
function handleImageUpload(stampId, input) {
    const file = input.files[0];
    if (!file) return;
    
    // Validate file type
    if (!file.type.startsWith('image/')) {
        alert('Please select an image file.');
        return;
    }
    
    // Validate file size (5MB)
    if (file.size > 5 * 1024 * 1024) {
        alert('File size must be less than 5MB.');
        return;
    }
    
    // Show upload progress
    document.getElementById('upload-progress').style.display = 'block';
    
    // Create FormData
    const formData = new FormData();
    formData.append('image', file);
    
    // Upload the file
    fetch(`/api/stamps/${stampId}/upload-image`, {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
            return response.text().then(text => {
                throw new Error(text);
            });
        }
        return response.json();
    })
    .then(data => {
        // Update the image display immediately
        updateImageDisplay(data.image_url, stampId);
        
        // Clear the file input
        input.value = '';
        
        console.log('Image uploaded successfully:', data.image_url);
    })
    .catch(error => {
        console.error('Error uploading image:', error);
        alert(`Failed to upload image: ${error.message}`);
    })
    .finally(() => {
        // Hide upload progress
        document.getElementById('upload-progress').style.display = 'none';
    });
}

// Function to update the image display without reloading the page
function updateImageDisplay(imageUrl, stampId) {
    const imageContainer = document.querySelector('.stamp-detail-image-container');
    
    // Create new image element
    const newImage = document.createElement('img');
    newImage.src = imageUrl + '?t=' + Date.now(); // Add timestamp to force reload
    newImage.alt = 'Stamp image';
    newImage.className = 'stamp-detail-img';
    newImage.id = 'stamp-image';
    
    // Handle image load error
    newImage.onerror = function() {
        this.style.display = 'none';
        showImagePlaceholder();
    };
    
    // Replace the current content with the new image
    imageContainer.innerHTML = '';
    imageContainer.appendChild(newImage);
    
    // Update the image controls to show "Change" and "Remove" buttons
    const controls = document.querySelector('.image-controls');
    controls.innerHTML = `
        <button class="btn btn-sm btn-outline-secondary me-2" onclick="triggerImageUpload()">
            <i class="bi bi-camera"></i> Change Image
        </button>
        <button class="btn btn-sm btn-outline-danger" onclick="removeImage('${stampId}')">
            <i class="bi bi-trash"></i> Remove
        </button>
    `;
}

// Function to show image placeholder
function showImagePlaceholder() {
    const imageContainer = document.querySelector('.stamp-detail-image-container');
    imageContainer.innerHTML = `
        <div class="stamp-detail-placeholder" id="image-placeholder">
            <i class="bi bi-image" style="font-size: 3rem; opacity: 0.3;"></i>
            <p class="text-muted mt-2">Image not available</p>
        </div>
    `;
}

// Remove image
function removeImage(stampId) {
    if (confirm('Are you sure you want to remove this image?')) {
        fetch(`/api/stamps/${stampId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ image_url: '' })
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    throw new Error(text);
                });
            }
            return response.json();
        })
        .then(data => {
            // Update the image display to show placeholder
            showImagePlaceholder();
            
            // Update the controls to show "Upload" button
            const controls = document.querySelector('.image-controls');
            controls.innerHTML = `
                <button class="btn btn-sm btn-primary" onclick="triggerImageUpload()">
                    <i class="bi bi-upload"></i> Upload Image
                </button>
            `;
            
            console.log('Image removed successfully');
        })
        .catch(error => {
            console.error('Error removing image:', error);
            alert(`Failed to remove image: ${error.message}`);
        });
    }
}

// Delete stamp
function deleteStamp(stampId) {
    if (confirm('Are you sure you want to delete this stamp? This action cannot be undone.')) {
        fetch(`/api/stamps/${stampId}`, {
            method: 'DELETE'
        })
        .then(response => {
            if (response.ok) {
                // Navigate back to gallery
                htmx.ajax('GET', '/views/stamps/gallery', '#stamp-view-content');
            } else {
                return response.text().then(text => {
                    throw new Error(`Failed to delete: ${text}`);
                });
            }
        })
        .catch(error => {
            console.error('Error deleting stamp:', error);
            alert(`Failed to delete stamp: ${error.message}`);
        });
    }
}

// Storage box selection handling
// function handleBoxInput(event, element) {
//     if (event.key === 'Enter') {
//         event.preventDefault(); // Prevent form submission
//         element.blur();         // Trigger the onchange event
//     }
// }

// function handleBoxChange(element) {
//     const stampId = element.dataset.stampId;
//     const boxName = element.value.trim();
//     const datalist = document.getElementById(element.getAttribute('list'));

//     // Case 1: Input is cleared, so unassign the box.
//     if (boxName === '') {
//         updateStampBox(stampId, null, element);
//         return;
//     }

//     // Find if the entered box name exists in our datalist
//     const existingOption = Array.from(datalist.options).find(opt => opt.value === boxName);

//     // Case 2: User selected an existing box.
//     if (existingOption) {
//         const boxId = existingOption.dataset.id;
//         updateStampBox(stampId, boxId, element);
//     } 
//     // Case 3: User entered a new box name.
//     else {
//         // Create the new box first
//         createNewBox(boxName).then(newBox => {
//             if (newBox && newBox.id) {
//                 // Add the new box to the datalist so it's available next time
//                 const newOption = document.createElement('option');
//                 newOption.value = newBox.name;
//                 newOption.dataset.id = newBox.id;
//                 datalist.appendChild(newOption);

//                 // Now, update the stamp to use the newly created box
//                 updateStampBox(stampId, newBox.id, element);
                
//                 // Optional: Trigger an event to refresh the box list in the sidebar
//                 htmx.trigger(document.body, 'newBoxAdded');
//             }
//         });
//     }
// }

// async function createNewBox(boxName) {
//     try {
//         const response = await fetch('/api/boxes', {
//             method: 'POST',
//             headers: { 'Content-Type': 'application/json' },
//             body: JSON.stringify({ name: boxName })
//         });
//         if (!response.ok) {
//             throw new Error(`Server returned ${response.status}: ${await response.text()}`);
//         }
//         return await response.json();
//     } catch (error) {
//         console.error('Error creating new box:', error);
//         alert(`Failed to create new box: ${error.message}`);
//         return null;
//     }
// }

// function updateStampBox(stampId, boxId, element) {
//     const payload = { box_id: boxId };
    
//     fetch(`/api/stamps/${stampId}`, {
//         method: 'PUT',
//         headers: { 'Content-Type': 'application/json' },
//         body: JSON.stringify(payload)
//     })
//     .then(response => {
//         if (!response.ok) {
//             return response.text().then(text => { throw new Error(text) });
//         }
//         return response.json();
//     })
//     .then(data => {
//         console.log('Stamp box updated successfully:', data);
//         // Visual feedback
//         element.style.transition = 'background-color 0.3s ease';
//         element.style.backgroundColor = '#d4edda'; // light green
//         setTimeout(() => {
//             element.style.backgroundColor = '';
//         }, 1200);
//     })
//     .catch(error => {
//         console.error('Error updating stamp box:', error);
//         alert(`Failed to save changes: ${error.message}`);
//         // Visual error feedback
//         element.style.backgroundColor = '#f8d7da'; // light red
//         setTimeout(() => {
//             element.style.backgroundColor = '';
//         }, 2000);
//     });
// }

// // Update status labels when checkbox changes
// document.addEventListener('change', function(event) {
//     if (event.target.type === 'checkbox' && event.target.dataset.field === 'is_owned') {
//         const ownedText = document.querySelector('.owned-text');
//         const neededText = document.querySelector('.needed-text');
        
//         if (event.target.checked) {
//             ownedText.classList.add('active');
//             neededText.classList.remove('active');
//         } else {
//             ownedText.classList.remove('active');
//             neededText.classList.add('active');
//         }
//     }
// });