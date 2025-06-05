/**
 * Fetches and adds a new draft row to the instances table.
 */
async function addDraftRow() {
    // Prevent adding multiple draft rows
    if (document.getElementById('draft-row')) {
        document.getElementById('draft-row').style.backgroundColor = '#fff3cd';
        setTimeout(() => { document.getElementById('draft-row').style.backgroundColor = ''; }, 1000);
        return;
    }

    // Reliably get the stampId from the data-attribute
    const container = document.querySelector('.your-copies-section');
    const stampId = container ? container.dataset.stampId : null;
    if (!stampId) {
        alert('Could not determine stamp ID.');
        return;
    }

    try {
        const response = await fetch(`/views/stamps/${stampId}/new-instance-row`);
        if (!response.ok) throw new Error('Could not load new copy form.');
        
        const rowHTML = await response.text();
        const tableBody = document.getElementById('copies-table-body');
        if (tableBody) {
            tableBody.insertAdjacentHTML('beforeend', rowHTML);
            tableBody.querySelector('#draft-row select')?.focus();
        }
    } catch (error) {
        alert(error.message);
    }
}

/**
 * Saves the new instance from the draft row.
 * @param {HTMLElement} button - The save button element.
 */
async function saveNewInstance(button) {
    const row = button.closest('tr');
    if (!row) return;

    const container = document.querySelector('.your-copies-section');
    const stampId = container ? container.dataset.stampId : null;
    if (!stampId) {
        alert('Could not determine stamp ID.');
        return;
    }


    const condition = row.querySelector('[name="condition"]').value.trim();
    const boxName = row.querySelector('[name="box_name"]').value.trim();
    const quantity = parseInt(row.querySelector('[name="quantity"]').value);

    if (!condition && !boxName && quantity === 0) {
        alert('Please choose a condition, box, and set the quantity.');
        return;
    }

    button.disabled = true;
    button.innerHTML = '<span class="spinner-border spinner-border-sm" role="status"></span>';

    try {
        const datalist = row.querySelector('datalist');
        const existingOption = Array.from(datalist.options).find(opt => opt.value === boxName);
        let boxId = null;

        if (existingOption) {
            boxId = existingOption.dataset.id;
        } else if (boxName !== '') {
            const newBox = await createNewBox(boxName);
            if (newBox && newBox.id) {
                boxId = newBox.id;
                document.querySelectorAll('datalist').forEach(dl => {
                    const newOption = document.createElement('option');
                    newOption.value = newBox.name;
                    newOption.dataset.id = newBox.id;
                    dl.appendChild(newOption);
                });
                htmx.trigger(document.body, 'newBoxAdded');
            }
        }

        const newInstanceData = { condition: condition || null, box_id: boxId, quantity: quantity };
        const response = await fetch(`/api/instances/${stampId}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newInstanceData)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText.includes('UNIQUE constraint') ? 'An instance with this condition and box already exists.' : errorText);
        }

        const savedInstance = await response.json();
        const realRowHTML = createRealRowHTML(savedInstance);
        row.outerHTML = realRowHTML;
        
        updateInstanceCount();
        updateTotalCopiesCount();
        htmx.trigger(document.body, 'newBoxAdded');

    } catch (error) {
        alert(`Failed to save: ${error.message}`);
        const saveIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24" fill="currentColor"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm-2-9.59L8.59 12 10 13.41 14 9.41l-1.41-1.41L10 10.59z"/></svg>`;
        button.disabled = false;
        button.innerHTML = saveIconSVG;
    }
}

/**
 * Creates the HTML for a standard, editable instance row from saved data.
 * @param {object} instance - The instance data from the API.
 * @returns {string} The HTML string for the new row.
 */
function createRealRowHTML(instance) {
    let allBoxes = [];
    const allBoxesDataEl = document.getElementById('all-boxes-data');
    
    // Safely parse the box data from the DOM
    if (allBoxesDataEl && allBoxesDataEl.textContent) {
        try {
            const parsedData = JSON.parse(allBoxesDataEl.textContent);
            if (Array.isArray(parsedData)) {
                allBoxes = parsedData;
            } else {
                console.error("Parsed 'all-boxes-data' is not an array:", parsedData);
            }
        } catch (e) {
            console.error("Failed to parse 'all-boxes-data':", e);
        }
    } else {
        console.error("'all-boxes-data' element not found or is empty.");
    }
    
    let boxOptionsHTML = allBoxes.map(box => `<option value="${box.name}" data-id="${box.id}"></option>`).join('');
    const boxName = instance.box_name || '';

    return `
        <tr data-instance-id="${instance.id}">
            <td>
                <select class="form-select instance-field" data-field="condition" data-instance-id="${instance.id}" onchange="saveInstanceField(this)">
                    <option value="" ${!instance.condition ? 'selected' : ''}>No condition</option>
                    <option value="Mint" ${instance.condition === 'Mint' ? 'selected' : ''}>Mint</option>
                    <option value="Used" ${instance.condition === 'Used' ? 'selected' : ''}>Used</option>
                    <option value="Damaged" ${instance.condition === 'Damaged' ? 'selected' : ''}>Damaged</option>
                    <option value="Fine" ${instance.condition === 'Fine' ? 'selected' : ''}>Fine</option>
                    <option value="Very Fine" ${instance.condition === 'Very Fine' ? 'selected' : ''}>Very Fine</option>
                    <option value="Excellent" ${instance.condition === 'Excellent' ? 'selected' : ''}>Excellent</option>
                </select>
            </td>
            <td>
                <input class="info-value-input instance-field" list="box-options-${instance.id}" value="${boxName}" placeholder="Type or select a box" data-instance-id="${instance.id}" onchange="handleBoxChange(this)" onkeydown="handleBoxInput(event, this)" autocomplete="off">
                <datalist id="box-options-${instance.id}">${boxOptionsHTML}</datalist>
            </td>
            <td>
                <div class="quantity-controls">
                     <button class="quantity-btn" onclick="adjustInstanceQuantity('${instance.id}', -1)"><i class="bi bi-dash"></i></button>
                     <input type="number" class="quantity-input instance-field" value="${instance.quantity}" min="0" data-field="quantity" data-instance-id="${instance.id}" onchange="saveInstanceField(this)" onblur="saveInstanceField(this)">
                     <button class="quantity-btn" onclick="adjustInstanceQuantity('${instance.id}', 1)"><i class="bi bi-plus"></i></button>
                </div>
            </td>
            <td>
                 <button class="btn btn-sm btn-outline-danger delete-instance-btn" onclick="deleteInstance('${instance.id}')" title="Delete this group"><i class="bi bi-trash"></i></button>
            </td>
        </tr>
    `;
}

// --- UI Update and State Management ---

function updateInstanceCount() {
    const countSpan = document.getElementById('instance-group-count');
    if (!countSpan) return;
    const rows = document.querySelectorAll('#copies-table-body tr[data-instance-id]').length;
    countSpan.textContent = `(${rows} ${rows === 1 ? 'group' : 'groups'})`;
}

function updateTotalCopiesCount() {
    const totalCopiesSpan = document.querySelector('.copies-count');
    if (!totalCopiesSpan) return;
    let total = 0;
    document.querySelectorAll('#copies-table-body tr[data-instance-id] .quantity-input').forEach(input => {
        total += parseInt(input.value) || 0;
    });
    totalCopiesSpan.textContent = `${total} ${total === 1 ? 'copy' : 'copies'}`;
}

/**
 * Removes an instance row from the table and updates the UI state.
 * @param {string} instanceId The ID of the instance to remove.
 */
function removeInstanceRow(instanceId) {
    const row = document.querySelector(`tr[data-instance-id="${instanceId}"]`);
    if (row) row.remove();
    
    const tableBody = document.getElementById('copies-table-body');
    if (tableBody && tableBody.children.length === 0) {
        addDraftRow();
    }
    
    updateInstanceCount();
    updateTotalCopiesCount();
    htmx.trigger(document.body, 'newBoxAdded');
}

// --- Core API Functions ---

function saveInstanceField(element) {
    const instanceId = element.dataset.instanceId;
    let value = (element.type === 'number') ? parseInt(element.value) || 0 : element.value;
    element.classList.add('saving');
    
    fetch(`/api/instances/${instanceId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ [element.dataset.field]: value })
    })
    .then(response => {
        if (response.status === 204) {
            removeInstanceRow(instanceId);
            return null;
        }
        if (!response.ok) return response.text().then(text => { throw new Error(text); });
        return response.json();
    })
    .then(data => {
        if (data) {
            element.classList.remove('saving');
            element.classList.add('saved');
            setTimeout(() => element.classList.remove('saved'), 1000);
            updateTotalCopiesCount();
            htmx.trigger(document.body, 'newBoxAdded');
        }
    })
    .catch(error => {
        element.classList.remove('saving');
        element.classList.add('error');
        setTimeout(() => element.classList.remove('error'), 2000);
        alert(`Failed to save changes: ${error.message}`);
    });
}

// Delete an entire instance group
function deleteInstance(instanceId) {
    if (confirm('Are you sure you want to delete this group of copies?')) {
        fetch(`/api/instances/${instanceId}`, { method: 'DELETE' })
        .then(response => {
            if (response.ok) {
                removeInstanceRow(instanceId);
            } else {
                return response.text().then(text => { throw new Error(`Failed to delete: ${text}`); });
            }
        })
        .catch(error => alert(error.message));
    }
}

async function createNewBox(boxName) {
    try {
        const response = await fetch('/api/boxes', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: boxName })
        });
        if (!response.ok) throw new Error(`Server returned ${response.status}: ${await response.text()}`);
        return await response.json();
    } catch (error) {
        console.error('Error creating new box:', error);
        alert(`Failed to create new box: ${error.message}`);
        return null;
    }
}

function handleBoxChange(element) {
    const instanceId = element.dataset.instanceId;
    const boxName = element.value.trim();
    const datalist = document.getElementById(element.getAttribute('list'));
    
    // Case 1: Input is cleared, so unassign the box.
    if (boxName === '') {
        updateInstanceBox(instanceId, null, element);
        return;
    }

    // Find if the entered box name exists in our datalist
    const existingOption = Array.from(datalist.options).find(opt => opt.value === boxName);

    // Case 2: User selected an existing box.
    if (existingOption) {
        updateInstanceBox(instanceId, existingOption.dataset.id, element);
    } else {
        createNewBox(boxName).then(newBox => {
            if (newBox && newBox.id) {
                // Add the new box to the datalist so it's available next time
                const newOption = document.createElement('option');
                newOption.value = newBox.name;
                newOption.dataset.id = newBox.id;
                datalist.appendChild(newOption);
                
                // Now, update the stamp to use the newly created box
                updateInstanceBox(instanceId, newBox.id, element);

                // Trigger an event to refresh the box list in the sidebar
                htmx.trigger(document.body, 'newBoxAdded');
            }
        });
    }
}

function updateInstanceBox(instanceId, boxId, element) {
    fetch(`/api/instances/${instanceId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ box_id: boxId })
    })
    .then(response => {
        if (!response.ok) return response.text().then(text => { throw new Error(text) });
        return response.json();
    })
    .then(data => {
        console.log('Stamp box updated successfully:', data);
        // Visual feedback
        element.style.transition = 'background-color 0.3s ease';
        element.style.backgroundColor = '#d4edda';  // light green
        setTimeout(() => { element.style.backgroundColor = ''; }, 1200);
        htmx.trigger(document.body, 'newBoxAdded');
    })
    .catch(error => {
        console.error('Error updating stamp box:', error);
        alert(`Failed to save changes: ${error.message}`);
        // Visual error feedback
        element.style.backgroundColor = '#f8d7da'; // light red
        setTimeout(() => { element.style.backgroundColor = ''; }, 2000);
    });
}

// Adjust instance quantity using +/- buttons
function adjustInstanceQuantity(instanceId, delta) {
    const input = document.querySelector(`input[data-instance-id="${instanceId}"][data-field="quantity"]`);
    if (!input) return;
    const currentValue = parseInt(input.value) || 0;
    const newValue = Math.max(0, currentValue + delta);
    input.value = newValue;
    saveInstanceField(input);
}

// Storage box selection handling
function handleBoxInput(event, element) {
    if (event.key === 'Enter') {
        event.preventDefault(); // Prevent form submission
        element.blur();         // Trigger the onchange event
    }
}
