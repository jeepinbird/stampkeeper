// settings.js - Handle settings page functionality and localStorage preferences

// Default preferences
const DEFAULT_PREFERENCES = {
    defaultView: 'gallery',
    defaultSort: 'name',
    sortDirection: 'ASC',
    itemsPerPage: 50
};

// Load preferences from localStorage
function loadPreferences() {
    try {
        const stored = localStorage.getItem('stampkeeper_preferences');
        if (stored) {
            const parsed = JSON.parse(stored);
            return { ...DEFAULT_PREFERENCES, ...parsed };
        }
    } catch (error) {
        console.error('Error loading preferences from localStorage:', error);
    }
    return DEFAULT_PREFERENCES;
}

// Save preferences to localStorage
function savePreferences(preferences) {
    try {
        localStorage.setItem('stampkeeper_preferences', JSON.stringify(preferences));
        console.log('Preferences saved:', preferences);
        return true;
    } catch (error) {
        console.error('Error saving preferences to localStorage:', error);
        return false;
    }
}

// Apply preferences to the current page
function applyPreferences() {
    const preferences = loadPreferences();
    
    try {
        // Apply view preference to the main page toggles
        const viewToggle = document.querySelector(`input[name="view_toggle"][id="view_${preferences.defaultView}"]`);
        if (viewToggle) {
            viewToggle.checked = true;
        }
        
        console.log('Applied preferences:', preferences);
    } catch (error) {
        console.error('Error applying preferences:', error);
    }
}

// Navigate to the user's preferred default view
function navigateToDefaultView() {
    const preferences = loadPreferences();
    const defaultPath = `/views/stamps/${preferences.defaultView}`;
    
    // Use HTMX to load the default view
    htmx.ajax('GET', defaultPath, '#stamp-view-content');
    
    // Update the view toggle to match
    const viewToggle = document.querySelector(`input[name="view_toggle"][id="view_${preferences.defaultView}"]`);
    if (viewToggle) {
        viewToggle.checked = true;
    }
}

// Update all HTMX links to use user's preferred view instead of hardcoded gallery
function updateHtmxLinksToPreferredView() {
    const preferences = loadPreferences();
    const preferredView = preferences.defaultView;
    
    try {
        // Update Quick Filters
        const filterLinks = document.querySelectorAll('input[name="owned_filter"]');
        filterLinks.forEach(input => {
            const label = document.querySelector(`label[for="${input.id}"]`);
            if (label) {
                const currentHxGet = label.getAttribute('hx-get');
                if (currentHxGet) {
                    // Replace gallery/list with preferred view
                    const newHxGet = currentHxGet.replace(/\/views\/stamps\/(gallery|list)/, `/views/stamps/${preferredView}`);
                    label.setAttribute('hx-get', newHxGet);
                }
            }
        });
        
        // Update box list links (these get refreshed via HTMX, so we'll handle them in the htmx:afterSwap event)
        
        console.log('Updated HTMX links to use preferred view:', preferredView);
    } catch (error) {
        console.error('Error updating HTMX links:', error);
    }
}

// Get preferences for use in HTMX requests
function getPreferencesForRequest() {
    const preferences = loadPreferences();
    return {
        sort: preferences.defaultSort,
        order: preferences.sortDirection,
        limit: preferences.itemsPerPage
    };
}

// Load preferences into the settings form
function loadSettingsForm() {
    const preferences = loadPreferences();
    
    try {
        // Default view
        const viewRadio = document.querySelector(`input[name="defaultView"][value="${preferences.defaultView}"]`);
        if (viewRadio) viewRadio.checked = true;
        
        // Items per page
        const itemsSelect = document.getElementById('itemsPerPage');
        if (itemsSelect) itemsSelect.value = preferences.itemsPerPage;
        
        // Default sort
        const sortSelect = document.getElementById('defaultSort');
        if (sortSelect) sortSelect.value = preferences.defaultSort;
        
        // Sort direction
        const directionRadio = document.querySelector(`input[name="sortDirection"][value="${preferences.sortDirection}"]`);
        if (directionRadio) directionRadio.checked = true;
        
        console.log('Settings form loaded with preferences:', preferences);
    } catch (error) {
        console.error('Error loading settings form:', error);
    }
}

// Save display preferences from the form
function saveDisplayPreferences() {
    try {
        const currentPrefs = loadPreferences();
        
        const formData = {
            defaultView: document.querySelector('input[name="defaultView"]:checked')?.value || currentPrefs.defaultView,
            itemsPerPage: parseInt(document.getElementById('itemsPerPage')?.value) || currentPrefs.itemsPerPage,
            defaultSort: document.getElementById('defaultSort')?.value || currentPrefs.defaultSort,
            sortDirection: document.querySelector('input[name="sortDirection"]:checked')?.value || currentPrefs.sortDirection
        };
        
        const newPrefs = { ...currentPrefs, ...formData };
        
        if (savePreferences(newPrefs)) {
            // Show success feedback
            const button = event.target;
            const originalText = button.innerHTML;
            button.innerHTML = '<i class="bi bi-check-circle me-1"></i>Saved!';
            button.classList.add('btn-success');
            button.classList.remove('btn-primary');
            
            setTimeout(() => {
                button.innerHTML = originalText;
                button.classList.remove('btn-success');
                button.classList.add('btn-primary');
            }, 2000);
            
            // Apply the new preferences immediately
            applyPreferences();
            
            // Navigate to the preferred default view if we're currently on settings
            if (document.querySelector('.settings-container')) {
                setTimeout(() => {
                    navigateToDefaultView();
                }, 500);
            };
        } else {
            alert('Failed to save preferences. Please try again.');
        }
    } catch (error) {
        console.error('Error saving display preferences:', error);
        alert('Failed to save preferences. Please try again.');
    }
}

// Reset all settings to defaults
function resetAllSettings() {
    if (confirm('Are you sure you want to reset all settings to their defaults? This cannot be undone.')) {
        try {
            localStorage.removeItem('stampkeeper_preferences');
            
            // Reload the settings form with defaults
            loadSettingsForm();
            
            // Apply defaults
            applyPreferences();
            
            alert('Settings have been reset to defaults.');
        } catch (error) {
            console.error('Error resetting settings:', error);
            alert('Failed to reset settings. Please try again.');
        }
    }
}

// Box management functions
function editBoxName(boxId) {
    const row = document.querySelector(`tr[data-box-id="${boxId}"]`);
    if (!row) return;
    
    const nameDisplay = row.querySelector('.box-name-display');
    const nameEdit = row.querySelector('.box-name-edit');
    const editBtn = row.querySelector('.edit-box-btn');
    const saveBtn = row.querySelector('.save-box-btn');
    const cancelBtn = row.querySelector('.cancel-box-btn');
    
    // Store original value before hiding
    const originalValue = nameDisplay.textContent.trim();
    nameEdit.value = originalValue;
    
    // Toggle visibility
    nameDisplay.classList.add('editing');
    nameEdit.classList.add('editing');
    nameEdit.focus();
    nameEdit.select();
    
    editBtn.style.display = 'none';
    saveBtn.style.display = 'inline-block';
    cancelBtn.style.display = 'inline-block';
}

function saveBoxName(boxId) {
    const row = document.querySelector(`tr[data-box-id="${boxId}"]`);
    if (!row) return;
    
    const nameEdit = row.querySelector('.box-name-edit');
    const newName = nameEdit.value.trim();
    
    if (!newName) {
        alert('Box name cannot be empty.');
        return;
    }
    
    // Save via API
    fetch(`/api/boxes/${boxId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newName })
    })
    .then(response => {
        if (!response.ok) throw new Error('Failed to update box name');
        return response.json();
    })
    .then(data => {
        // Update the display
        const nameDisplay = row.querySelector('.box-name-display');
        nameDisplay.textContent = newName;
        
        cancelBoxEdit(boxId);
        
        // Trigger box list refresh in sidebar
        htmx.trigger(document.body, 'newBoxAdded');
    })
    .catch(error => {
        console.error('Error updating box name:', error);
        alert('Failed to update box name. Please try again.');
    });
}

function cancelBoxEdit(boxId) {
    const row = document.querySelector(`tr[data-box-id="${boxId}"]`);
    if (!row) return;
    
    const nameDisplay = row.querySelector('.box-name-display');
    const nameEdit = row.querySelector('.box-name-edit');
    const editBtn = row.querySelector('.edit-box-btn');
    const saveBtn = row.querySelector('.save-box-btn');
    const cancelBtn = row.querySelector('.cancel-box-btn');
    
    // Reset to original state
    nameDisplay.classList.remove('editing');
    nameEdit.classList.remove('editing');
    nameEdit.value = nameDisplay.textContent.trim(); // Reset to original value
    
    editBtn.style.display = 'inline-block';
    saveBtn.style.display = 'none';
    cancelBtn.style.display = 'none';
}

function deleteBox(boxId, boxName) {
    if (confirm(`Are you sure you want to delete the box "${boxName}"? This cannot be undone.`)) {
        fetch(`/api/boxes/${boxId}`, { method: 'DELETE' })
        .then(response => {
            if (!response.ok) throw new Error('Failed to delete box');
            
            // Remove from the table
            const row = document.querySelector(`tr[data-box-id="${boxId}"]`);
            if (row) row.remove();
            
            // Trigger box list refresh in sidebar
            htmx.trigger(document.body, 'newBoxAdded');
        })
        .catch(error => {
            console.error('Error deleting box:', error);
            alert('Failed to delete box. Please try again.');
        });
    }
}

function createNewBoxFromSettings() {
    const nameInput = document.getElementById('newBoxName');
    const boxName = nameInput.value.trim();
    
    if (!boxName) {
        alert('Please enter a box name.');
        nameInput.focus();
        return;
    }
    
    fetch('/api/boxes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: boxName })
    })
    .then(response => {
        if (!response.ok) throw new Error('Failed to create box');
        return response.json();
    })
    .then(data => {
        // Add to the table
        const tableBody = document.getElementById('boxes-table-body');
        const newRow = document.createElement('tr');
        newRow.setAttribute('data-box-id', data.id);
        newRow.innerHTML = `
            <td>
                <span class="box-name-display">${data.name}</span>
                <input type="text" class="form-control box-name-edit" value="${data.name}" style="display: none;">
            </td>
            <td>0</td>
            <td>
                <button class="btn btn-sm btn-outline-secondary me-1 edit-box-btn" onclick="editBoxName('${data.id}')">
                    <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-outline-secondary me-1 save-box-btn" onclick="saveBoxName('${data.id}')" style="display: none;">
                    <i class="bi bi-check"></i>
                </button>
                <button class="btn btn-sm btn-outline-secondary me-1 cancel-box-btn" onclick="cancelBoxEdit('${data.id}')" style="display: none;">
                    <i class="bi bi-x"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger delete-box-btn" onclick="deleteBox('${data.id}', '${data.name}')">
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        `;
        tableBody.appendChild(newRow);
        
        // Clear the input
        nameInput.value = '';
        
        // Trigger box list refresh in sidebar
        htmx.trigger(document.body, 'newBoxAdded');
    })
    .catch(error => {
        console.error('Error creating box:', error);
        alert('Failed to create box. Please try again.');
    });
}

function handleNewBoxEnter(event) {
    if (event.key === 'Enter') {
        event.preventDefault();
        createNewBoxFromSettings();
    }
}

// Initialize settings when the page loads
document.addEventListener('DOMContentLoaded', function() {
    // Only run this if we're on the settings page
    if (document.querySelector('.settings-container')) {
        loadSettingsForm();
    }
    
    // Always apply preferences when any page loads
    applyPreferences();
    
    // Update HTMX links to use preferred view
    updateHtmxLinksToPreferredView();
});

// Update box links after they're refreshed via HTMX
document.body.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target && evt.detail.target.id === 'box-list') {
        updateBoxLinksToPreferredView();
    }
});

// Update box list links specifically
function updateBoxLinksToPreferredView() {
    const preferences = loadPreferences();
    const preferredView = preferences.defaultView;
    
    try {
        const boxLinks = document.querySelectorAll('#box-list .list-group-item[hx-get]');
        boxLinks.forEach(link => {
            const currentHxGet = link.getAttribute('hx-get');
            if (currentHxGet) {
                const newHxGet = currentHxGet.replace(/\/views\/stamps\/(gallery|list)/, `/views/stamps/${preferredView}`);
                link.setAttribute('hx-get', newHxGet);
            }
        });
        
        console.log('Updated box links to use preferred view:', preferredView);
    } catch (error) {
        console.error('Error updating box links:', error);
    }
}