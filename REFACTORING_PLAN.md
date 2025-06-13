# StampKeeper Refactoring Plan: Go + HTMX First, Minimal JavaScript

**Goal:** Transform codebase to embrace "HTML over the wire" philosophy with minimal JavaScript

## **Phase 1: HTMX-First Architecture** ⚡
**Status:** 🔄 Pending
**Goal:** Replace 90% of JavaScript with declarative HTMX + Go

### 1.1 Convert JavaScript Operations to HTMX + Go Endpoints
**Files to modify:** 
- `static/js/stamp-detail.js` → Remove entirely
- `static/js/settings.js` → Remove entirely  
- `static/js/new-stamp.js` → Minimal Alpine.js only
- `internal/handlers/` → Add new HTMX endpoints

**New Go endpoints to create:**
- [ ] `PUT /api/stamps/{id}/field` - Handle individual field updates
- [ ] `POST /views/stamps/{id}/tag` - Add tag and return updated tag section
- [ ] `DELETE /views/stamps/{id}/tag/{name}` - Remove tag and return updated section
- [ ] `POST /views/boxes` - Create box and return updated sidebar
- [ ] `PUT /views/preferences` - Save user preferences server-side

### 1.2 Server-Side State Management
- [ ] Move localStorage preferences to Go session/cookies
- [ ] Use HTMX `hx-vals` with server-generated values
- [ ] Implement server-side view state persistence
- [ ] Create `internal/middleware/sessions.go`

### 1.3 Replace Complex JavaScript with HTMX Patterns
**Current problematic code in `static/index.html` lines 144-182:**
```javascript
// REMOVE: Complex htmx:configRequest event handler
document.body.addEventListener('htmx:configRequest', function(evt) {
    // 40+ lines of parameter injection logic
});
```

**Replace with:**
```html
<!-- Clean HTMX with server handling complexity -->
<input hx-get="/views/stamps/gallery" 
       hx-include="[name='search'], [name='owned_filter']:checked"
       hx-trigger="keyup changed delay:500ms, search">
```

## **Phase 2: Alpine.js for Minimal Client-Side Needs** 🏔️
**Status:** 🔄 Pending
**Only use Alpine.js for pure UI interactions (No Server State)**

### 2.1 Alpine.js Use Cases (< 50 lines total)
- **Image upload progress bars** 
- **Form validation feedback**
- **Dropdown/modal toggles**
- **Client-side sorting/filtering of already-loaded data**

### 2.2 Create `static/js/alpine-components.js`
```html
<!-- File upload with progress -->
<div x-data="{ uploading: false, progress: 0 }">
  <input type="file" @change="uploadFile($event)">
  <div x-show="uploading" class="progress-bar">
    <div :style="`width: ${progress}%`"></div>
  </div>
</div>

<!-- Form validation -->
<form x-data="{ name: '', isValid: false }" 
      @submit.prevent="isValid && $el.submit()">
  <input x-model="name" @input="isValid = name.length > 0">
</form>
```

## **Phase 3: Go Backend as the Smart Layer** 🧠
**Status:** 🔄 Pending
**Move intelligence from client to server**

### 3.1 Smart HTMX Response Handlers
**Files to create:**
- [ ] `internal/handlers/htmx/stamps.go`
- [ ] `internal/handlers/htmx/preferences.go`
- [ ] `internal/handlers/htmx/boxes.go`

```go
// Handle complex filter combinations server-side
func (h *ViewHandler) GetStampsView(w http.ResponseWriter, r *http.Request) {
    // Parse ALL filter parameters
    filters := parseFiltersFromRequest(r)
    userPrefs := getUserPreferences(r) // From cookies/session
    
    // Apply intelligent defaults
    view := determineOptimalView(filters, userPrefs)
    
    // Return appropriate template with all state
    h.renderStampsView(w, view, filters, userPrefs)
}
```

### 3.2 Server-Side Form Processing
- [ ] Handle all CRUD operations through Go
- [ ] Return HTML fragments with updated state
- [ ] Use HTMX `hx-trigger` to chain server actions

### 3.3 Session-Based Preferences
```go
// Replace localStorage with server session
func (h *ViewHandler) SaveUserPreferences(w http.ResponseWriter, r *http.Request) {
    prefs := parsePreferences(r)
    saveToSession(r, prefs)
    
    // Return updated UI state
    h.templates.ExecuteTemplate(w, "preferences-saved.html", prefs)
}
```

## **Phase 4: Eliminate JavaScript Files** 🗑️
**Status:** 🔄 Pending

### Files to DELETE:
- [ ] `static/js/settings.js` (400+ lines → 0 lines)
- [ ] `static/js/new-stamp.js` → minimal Alpine.js
- [ ] `static/js/stamp-detail.js` → pure HTMX
- [ ] `static/js/stamp-instance.js` → server-side handling

### Remaining Alpine.js (< 50 lines total):
```html
<!-- Only for pure UI interactions -->
<div x-data="{ showModal: false }">
  <button @click="showModal = true">Upload Image</button>
  <div x-show="showModal" x-transition>
    <!-- Upload form with HTMX -->
  </div>
</div>
```

## **Phase 5: HTMX Advanced Patterns** 🚀
**Status:** 🔄 Pending

### 5.1 Server-Sent Events (Replace WebSocket needs)
```go
// Real-time updates without JavaScript
func (h *ViewHandler) StampUpdatesSSE(w http.ResponseWriter, r *http.Request) {
    // Send HTML fragments via SSE
}
```

### 5.2 HTMX Extensions
- [ ] `hx-preserve` for maintaining scroll position
- [ ] `hx-push-url` for proper navigation
- [ ] `hx-boost` for progressive enhancement

### 5.3 Smart Server Responses
```go
// Return different templates based on request context
func (h *ViewHandler) SmartResponse(w http.ResponseWriter, r *http.Request) {
    if isHTMXRequest(r) {
        // Return fragment
        h.templates.ExecuteTemplate(w, "stamp-card.html", data)
    } else {
        // Return full page
        h.templates.ExecuteTemplate(w, "full-page.html", data)
    }
}
```

## **Target File Structure:**
```
static/js/
├── alpine-components.js  (~30 lines - pure UI only)
└── [DELETE everything else]

internal/
├── handlers/
│   ├── htmx/           # HTMX-specific handlers
│   └── preferences/    # Server-side state management
├── middleware/
│   └── sessions.go     # Replace localStorage
└── templates/
    └── fragments/      # Atomic HTMX response templates
```

## **The "JavaScript Escape Hatch" Rule:**
Only use Alpine.js when you need:
1. **Immediate UI feedback** (before server round-trip)
2. **Client-side calculations** (progress bars, validation)
3. **Pure presentation logic** (animations, transitions)

Everything else stays in Go + HTMX.

## **Implementation Checklist:**

### Phase 1 Tasks:
- [x] Remove complex `htmx:configRequest` handler from `static/index.html`
- [x] Update HTMX elements to use `hx-include` for parameter passing
- [x] Update Go service layer to handle new parameter format
- [ ] Create new HTMX endpoints in Go
- [ ] Add session middleware
- [ ] Convert stamp field updates to HTMX forms
- [ ] Convert tag management to server-side operations
- [ ] Move preferences from localStorage to server sessions

### Phase 2 Tasks:
- [ ] Install Alpine.js
- [ ] Create minimal `alpine-components.js`
- [ ] Convert image upload to Alpine.js + HTMX
- [ ] Add form validation with Alpine.js

### Phase 3 Tasks:
- [ ] Create smart HTMX handlers
- [ ] Implement server-side state management
- [ ] Add session-based preferences

### Phase 4 Tasks:
- [ ] Delete `settings.js`
- [ ] Delete `stamp-detail.js`
- [ ] Delete `new-stamp.js`
- [ ] Delete `stamp-instance.js`
- [ ] Update templates to remove inline JavaScript

### Phase 5 Tasks:
- [ ] Add HTMX extensions
- [ ] Implement SSE for real-time updates
- [ ] Add smart response handling

## **Benefits Expected:**
✅ **Dramatically Less JavaScript** (400+ lines → ~30 lines)  
✅ **Server-Side Intelligence** - Go handles complexity  
✅ **Better SEO & Accessibility** - Works without JavaScript  
✅ **Simpler Architecture** - One source of truth (server)  
✅ **Faster Development** - No client/server state sync issues  
✅ **More Maintainable** - Logic in one place (Go)  
✅ **Progressive Enhancement** - Graceful degradation built-in  

---

**Last Updated:** January 2025  
**Current Status:** Ready to begin Phase 1  
**Estimated Completion:** 5 phases, implementing incrementally to maintain functionality