// new-stamp.js - Handles creating new stamps from the form

function clearPlaceholderOnFirstClick(element) {
    // Check if this is the first click on a placeholder field
    if (element.dataset.isPlaceholder === 'true') {
        const currentText = element.textContent.trim();
        
        // Clear the placeholder text if it matches our default values
        if (currentText === 'New Stamp' || currentText === 'Click to add Scott number') {
            // Clear the content and set cursor position
            element.innerHTML = '';
            element.focus();
            element.classList.remove('required-field');
        }
        
        // Mark as no longer a placeholder
        element.dataset.isPlaceholder = 'false';
    }
}

// Handle Enter key specifically for new stamp form
function handleNewStampEnterKey(event, element) {
    if (event.key === 'Enter') {
        event.preventDefault();
        element.blur(); // This will trigger the save
    }
}

function saveNewStampField(element) {
    const field = element.dataset.field;
    let value;
    
    if (element.type === 'date') {
        value = element.value;
    } else if (element.contentEditable === 'true') {
        value = element.textContent.trim();
        
        // Don't save placeholder values
        if (value === 'New Stamp' || value === 'Click to add Scott number') {
            value = '';
        }
        
        // Clear placeholder styling if user entered real content
        if (value && value.length > 0) {
            element.classList.remove('required-field');
            element.dataset.isPlaceholder = 'false';
        }
    } else {
        value = element.value.trim();
    }
    
    // Store the value in our temporary data object
    if (window.newStampData) {
        window.newStampData[field] = value;
    }
    
    console.log('Saved field:', field, 'Value:', value);
    
    // Visual feedback
    element.style.backgroundColor = '#d4edda';
    setTimeout(() => {
        element.style.backgroundColor = '';
    }, 1000);
}

function addNewTagToNewStamp() {
    const tagName = prompt('Enter tag name:');
    if (tagName && tagName.trim()) {
        const trimmedTag = tagName.trim();
        
        // Add to our data structure
        if (!window.newStampData.tags) {
            window.newStampData.tags = [];
        }
        
        if (!window.newStampData.tags.includes(trimmedTag)) {
            window.newStampData.tags.push(trimmedTag);
            
            // Add to the UI
            const container = document.getElementById('tags-container');
            const addButton = container.querySelector('.add-tag-btn');
            
            const tagElement = document.createElement('span');
            tagElement.className = 'tag-pill editable-tag';
            tagElement.innerHTML = `${trimmedTag} <button class="tag-remove" onclick="removeTagFromNewStamp('${trimmedTag}', this)">×</button>`;
            
            container.insertBefore(tagElement, addButton);
        }
    }
}

function removeTagFromNewStamp(tagName, buttonElement) {
    // Remove from data structure
    if (window.newStampData && window.newStampData.tags) {
        window.newStampData.tags = window.newStampData.tags.filter(tag => tag !== tagName);
    }
    
    // Remove from UI
    buttonElement.parentElement.remove();
}

function createStampFromForm() {
    // Validate required fields
    if (!window.newStampData.name || window.newStampData.name === 'New Stamp' || window.newStampData.name === '') {
        alert('Please enter a stamp name before creating the stamp.');
        const nameField = document.querySelector('[data-field="name"]');
        if (nameField) {
            nameField.focus();
            nameField.classList.add('required-field');
        }
        return;
    }
    
    // Prepare the stamp data
    const stampData = {
        name: window.newStampData.name,
        scott_number: window.newStampData.scott_number || null,
        issue_date: window.newStampData.issue_date || null,
        series: window.newStampData.series || null,
        tags: window.newStampData.tags || []
    };
    
    // Filter out empty values
    Object.keys(stampData).forEach(key => {
        if (stampData[key] === '' || stampData[key] === 'Click to add Scott number') {
            stampData[key] = null;
        }
    });
    
    console.log('Creating stamp with data:', stampData);
    
    // Show loading state
    const createButton = document.querySelector('.btn-edit');
    const originalText = createButton.innerHTML;
    createButton.disabled = true;
    createButton.innerHTML = '<span class="spinner-border spinner-border-sm" role="status"></span> Creating...';
    
    // Send to API
    fetch('/api/stamps', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(stampData)
    })
    .then(response => {
        if (!response.ok) {
            return response.text().then(text => {
                throw new Error(`HTTP ${response.status}: ${text}`);
            });
        }
        return response.json();
    })
    .then(data => {
        console.log('Stamp created successfully:', data);
        
        // Navigate to the new stamp's detail page
        htmx.ajax('GET', `/views/stamps/detail/${data.id}`, '#stamp-view-content');
        
        // Clear our temporary data
        window.newStampData = null;
    })
    .catch(error => {
        console.error('Error creating stamp:', error);
        alert(`Failed to create stamp: ${error.message}`);
        
        // Reset button
        createButton.disabled = false;
        createButton.innerHTML = originalText;
    });
}