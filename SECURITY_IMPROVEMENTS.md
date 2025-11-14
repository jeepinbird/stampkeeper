# Security & Code Quality Improvement Plan

**Branch**: `cleaning-up` → `main`
**Created**: 2025-11-13
**Status**: In Progress

## Overview

This document tracks the implementation of security fixes and code quality improvements identified in the code review for the HTMX refactoring PR. Items are organized by priority with detailed implementation steps.

---

## 🚨 CRITICAL - Must Fix Before Merge to Production

### 1. SQL Injection Risk in Batch Operations
- **Priority**: CRITICAL
- **Files**: `internal/services/stamps.go:472-480, 488-491, 536-539`
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 1-2 hours
- **Actual Effort**: 1 hour
- **Completed**: 2025-11-13

**Issue**: Batch loading operations (`batchLoadStampTags`, `batchLoadStampInstances`) build SQL queries with stampIDs without validating they are UUIDs.

**Implementation Steps**:
1. Add UUID validation helper function to `internal/services/stamps.go`:
   ```go
   func validateStampIDs(stampIDs []string) error {
       for _, id := range stampIDs {
           if _, err := uuid.Parse(id); err != nil {
               return fmt.Errorf("invalid stamp ID: %s", id)
           }
       }
       return nil
   }
   ```

2. Add validation at the start of `batchLoadStampTags()` (before line 476):
   ```go
   if err := validateStampIDs(stampIDs); err != nil {
       return nil, err
   }
   ```

3. Add validation at the start of `batchLoadStampInstances()` (before line 520):
   ```go
   if err := validateStampIDs(stampIDs); err != nil {
       return nil, err
   }
   ```

4. Add tests for malicious input:
   - Test with `'; DROP TABLE stamps; --`
   - Test with `../../../etc/passwd`
   - Test with valid and invalid UUID mix

**Success Criteria**: All stampID inputs validated before query construction

---

### 2. Path Traversal Vulnerability in Image Upload
- **Priority**: CRITICAL
- **Files**: `internal/handlers/htmx.go:554-558`
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 1 hour
- **Actual Effort**: 30 minutes
- **Completed**: 2025-11-13

**Issue**: stampID from URL parameters is used in file paths without validation.

**Implementation Steps**:
1. Add validation in `UploadStampImage()` after getting stampID (after line 549):
   ```go
   if _, err := uuid.Parse(stampID); err != nil {
       http.Error(w, "Invalid stamp ID", http.StatusBadRequest)
       return
   }
   ```

2. Add validation in `DeleteStampImage()` after getting stampID (after line 737):
   ```go
   if _, err := uuid.Parse(stampID); err != nil {
       http.Error(w, "Invalid stamp ID", http.StatusBadRequest)
       return
   }
   ```

3. Add validation anywhere else stampID is used in file operations

4. Add tests:
   - Test with `../../etc/passwd`
   - Test with `../static/hack.jpg`
   - Test with valid UUID

**Success Criteria**: All stampIDs validated before file path operations

---

### 3. Resource Leak on Upload Failure
- **Priority**: CRITICAL
- **Files**: `internal/handlers/htmx.go:672-683`
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 2 hours
- **Actual Effort**: 1 hour
- **Completed**: 2025-11-13

**Issue**: Failed uploads leave partial files on disk. Deferred close doesn't check if file creation succeeded.

**Implementation Steps**:
1. Replace the file creation/copy logic in `UploadStampImage()` (lines 659-677):
   ```go
   dst, err := os.Create(filepath)
   if err != nil {
       http.Error(w, "Error creating file", http.StatusInternalServerError)
       return
   }

   // Cleanup flag - only keep file if everything succeeds
   cleanup := true
   defer func() {
       if err := dst.Close(); err != nil {
           log.Printf("Error closing file: %v", err)
       }
       if cleanup {
           if err := os.Remove(filepath); err != nil {
               log.Printf("Error removing partial upload: %v", err)
           }
       }
   }()

   _, err = io.Copy(dst, file)
   if err != nil {
       http.Error(w, "Error saving file", http.StatusInternalServerError)
       return
   }

   // Success - don't delete the file
   cleanup = false
   ```

2. Add similar cleanup logic for backup file creation (lines 614-635)

3. Add tests:
   - Simulate disk full during copy
   - Simulate permission errors
   - Verify no orphaned files remain

**Success Criteria**: No partial files left on disk after failed uploads

---

### 4. Missing CSRF Protection
- **Priority**: CRITICAL
- **Files**: `internal/router/router.go`, `internal/handlers/views.go`, `templates/index.html`
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 3-4 hours
- **Actual Effort**: 2 hours
- **Completed**: 2025-11-13

**Issue**: All state-changing endpoints are vulnerable to CSRF attacks.

**Implementation Steps**:
1. Add CSRF dependency to `go.mod`:
   ```bash
   go get github.com/gorilla/csrf
   ```

2. Update `internal/router/router.go` to add CSRF middleware:
   ```go
   import "github.com/gorilla/csrf"

   func SetupRouter(db *sql.DB) *mux.Router {
       // ... existing code ...

       // CSRF Protection
       csrfKey := []byte(os.Getenv("CSRF_KEY")) // 32 bytes
       if len(csrfKey) == 0 {
           csrfKey = []byte("32-byte-long-auth-key!!!!!!!!!!")
           log.Println("WARNING: Using default CSRF key. Set CSRF_KEY env var!")
       }

       csrfMiddleware := csrf.Protect(
           csrfKey,
           csrf.Secure(false), // Set to true in production with HTTPS
           csrf.Path("/"),
       )

       r.Use(csrfMiddleware)

       // ... rest of routes ...
   }
   ```

3. Update `templates/index.html` to include CSRF token in forms:
   ```html
   <meta name="csrf-token" content="{{ .csrfToken }}">
   ```

4. Update all forms in templates to include CSRF token:
   ```html
   <input type="hidden" name="gorilla.csrf.Token" value="{{ .csrfToken }}">
   ```

5. Update HTMX requests to include CSRF token in headers:
   ```html
   <script>
   document.body.addEventListener('htmx:configRequest', (event) => {
       event.detail.headers['X-CSRF-Token'] =
           document.querySelector('meta[name="csrf-token"]').content;
   });
   </script>
   ```

6. Add CSRF token to all handler template data

7. Add `.env` documentation for `CSRF_KEY`

**Success Criteria**: All mutation endpoints protected, CSRF attacks blocked

---

### 5. Inconsistent Error Handling
- **Priority**: CRITICAL
- **Files**: `internal/handlers/errors.go` (new), `internal/handlers/htmx.go` (multiple locations)
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 2-3 hours
- **Actual Effort**: 1.5 hours
- **Completed**: 2025-11-13

**Issue**: Some errors expose internal details, others don't. Need standardization.

**Implementation Steps**:
1. Create error handling utilities in new file `internal/handlers/errors.go`:
   ```go
   package handlers

   import (
       "log"
       "net/http"
   )

   // LogAndReturnError logs detailed error internally and returns generic message
   func LogAndReturnError(w http.ResponseWriter, err error, context string, statusCode int) {
       log.Printf("ERROR [%s]: %v", context, err)
       http.Error(w, getGenericMessage(statusCode), statusCode)
   }

   func getGenericMessage(statusCode int) string {
       switch statusCode {
       case http.StatusBadRequest:
           return "Invalid request"
       case http.StatusNotFound:
           return "Resource not found"
       case http.StatusConflict:
           return "Resource conflict"
       case http.StatusInternalServerError:
           return "Internal server error"
       default:
           return "An error occurred"
       }
   }
   ```

2. Replace all error handling in `htmx.go` to use new utility:
   ```go
   // Before:
   http.Error(w, "Failed to create stamp: "+err.Error(), http.StatusInternalServerError)

   // After:
   LogAndReturnError(w, err, "CreateStamp", http.StatusInternalServerError)
   ```

3. Audit all handlers for exposed errors (lines 512, 445, 324, etc.)

4. Add structured logging option for future JSON logs

**Success Criteria**: No internal errors exposed to clients, all errors logged

---

## ⚠️ HIGH PRIORITY - Should Fix Before Merge

### 6. Race Condition in Image Backup
- **Priority**: HIGH
- **Files**: `internal/handlers/htmx.go:628-636`
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 2 hours
- **Actual Effort**: 30 minutes
- **Completed**: 2025-11-13

**Issue**: Check-then-act pattern between file existence check and rename operation.

**Implementation Steps**:
1. Replace check-then-rename logic with atomic operation:
   ```go
   if existingStamp.ImageURL != nil && *existingStamp.ImageURL != "" {
       currentImageURL := *existingStamp.ImageURL
       if strings.HasPrefix(currentImageURL, "/static/images/stamps/") {
           currentFilename := strings.TrimPrefix(currentImageURL, "/static/images/stamps/")
           currentFilepath := filepath.Join(imagesDir, currentFilename)
           backupFilepath := currentFilepath + ".bak"

           // Atomic rename - will fail if source doesn't exist
           // That's fine, we don't want to error out
           if err := os.Rename(currentFilepath, backupFilepath); err != nil {
               // Log but don't fail - file might not exist
               log.Printf("Could not backup existing image: %v", err)
           }
       }
   }
   ```

2. Add file locking for concurrent access protection (optional, more complex)

3. Add tests for concurrent uploads

**Success Criteria**: No lost backups from race conditions

---

### 7. Transaction Rollback Error Handling
- **Priority**: HIGH
- **Files**: `internal/database/transaction.go` (new), `internal/services/stamps.go`, `internal/services/boxes.go`
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 1 hour
- **Actual Effort**: 1 hour
- **Completed**: 2025-11-13

**Issue**: Rollback errors are logged but not properly handled.

**Implementation Steps**:
1. Create rollback helper in `internal/database/transaction.go`:
   ```go
   package database

   import (
       "database/sql"
       "log"
   )

   // Rollback handles transaction rollback with proper error logging
   func Rollback(tx *sql.Tx, context string) {
       if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
           log.Printf("CRITICAL: Transaction rollback failed [%s]: %v", context, err)
           // TODO: Add alerting/monitoring here
       }
   }
   ```

2. Replace all `tx.Rollback()` calls:
   ```go
   // Before:
   if err := tx.Rollback(); err != nil {
       log.Printf("Rollback error: %v", err)
   }

   // After:
   database.Rollback(tx, "CreateStamp-TagCreation")
   ```

3. Update all locations in `stamps.go` (lines 324, 334, 344)

4. Add TODO comments for monitoring/alerting

**Success Criteria**: Rollback failures clearly logged as CRITICAL

---

### 8. Extract Duplicate Box Creation Logic
- **Priority**: HIGH
- **Files**: `internal/handlers/htmx.go:1031-1064` (new helper), replaced duplicates at lines 747-752, 850-857
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 2 hours
- **Actual Effort**: 1.5 hours
- **Completed**: 2025-11-13

**Issue**: Box lookup and creation logic duplicated in two handlers.

**Implementation Steps**:
1. Add method to `HTMXHandler`:
   ```go
   // getOrCreateBox looks up a box by name or creates it if it doesn't exist
   // Returns the box ID or nil if no box name provided
   func (h *HTMXHandler) getOrCreateBox(boxName string) (*string, error) {
       if boxName == "" {
           return nil, nil
       }

       // Look up existing box
       boxes, err := h.boxService.GetBoxes()
       if err != nil {
           return nil, fmt.Errorf("failed to lookup boxes: %w", err)
       }

       for _, box := range boxes {
           if box.Name == boxName {
               return &box.ID, nil
           }
       }

       // Create new box
       newBox := &models.StorageBox{
           ID:          uuid.New().String(),
           Name:        boxName,
           DateCreated: time.Now(),
       }

       createdBox, err := h.boxService.CreateBox(newBox)
       if err != nil {
           return nil, fmt.Errorf("failed to create box: %w", err)
       }

       return &createdBox.ID, nil
   }
   ```

2. Replace duplicate logic in `CreateStampInstance()` (lines 735-767):
   ```go
   boxID, err := h.getOrCreateBox(boxName)
   if err != nil {
       LogAndReturnError(w, err, "CreateStampInstance-BoxLookup", http.StatusInternalServerError)
       return
   }
   ```

3. Replace duplicate logic in `UpdateInstanceField()` (lines 871-901)

4. Add tests for box creation scenarios

**Success Criteria**: Single source of truth for box creation logic

---

### 9. Unbounded Query Results
- **Priority**: HIGH
- **Files**: `internal/services/boxes.go:21-71` (updated GetBoxes with optional limit)
- **Status**: ✅ COMPLETED
- **Estimated Effort**: 3 hours
- **Actual Effort**: 1 hour
- **Completed**: 2025-11-13

**Issue**: `GetBoxes()` called without pagination, could cause memory issues.

**Implementation Steps**:
1. Add limit parameter to `GetBoxes()` in `internal/services/boxes.go`:
   ```go
   func (s *BoxService) GetBoxes(limit int) ([]models.StorageBox, error) {
       query := `SELECT id, name, date_created FROM storage_boxes
                 WHERE date_deleted IS NULL
                 ORDER BY name ASC`

       if limit > 0 {
           query += fmt.Sprintf(" LIMIT %d", limit)
       }

       // ... rest of implementation
   }
   ```

2. Update all `GetBoxes()` calls to specify limits:
   - Settings page: no limit (showing all boxes)
   - Dropdowns: limit 1000 (reasonable cap)

3. Add warning log if box count exceeds threshold:
   ```go
   if len(boxes) >= 1000 {
       log.Printf("WARNING: Box count approaching limit: %d", len(boxes))
   }
   ```

4. Consider adding autocomplete for box selection if count is high

**Success Criteria**: Box queries capped at reasonable limits

---

## 📋 MEDIUM PRIORITY - Fix After Merge

### 10. File Extension Whitelist
- **Priority**: MEDIUM
- **Files**: `internal/handlers/htmx.go:638-653`
- **Status**: ⬜ Not Started
- **Estimated Effort**: 1 hour

**Issue**: No whitelist for allowed file extensions, could accept SVG with embedded JS.

**Implementation Steps**:
1. Add constant at top of `htmx.go`:
   ```go
   var allowedImageExtensions = map[string]bool{
       ".jpg":  true,
       ".jpeg": true,
       ".png":  true,
       ".gif":  true,
       ".webp": true,
   }
   ```

2. Add validation after extension detection (after line 653):
   ```go
   ext = strings.ToLower(ext)
   if !allowedImageExtensions[ext] {
       http.Error(w, "Invalid file type. Allowed: JPG, PNG, GIF, WebP", http.StatusBadRequest)
       return
   }
   ```

3. Update error messages to show allowed types

**Success Criteria**: Only safe image formats accepted

---

### 11. Image Cleanup on Stamp Deletion
- **Priority**: MEDIUM
- **Files**: `internal/services/stamps.go:312-348`
- **Status**: ⬜ Not Started
- **Estimated Effort**: 2 hours

**Issue**: Deleted stamps leave orphaned image files.

**Implementation Steps**:
1. Update `DeleteStamp()` to remove image before soft delete:
   ```go
   func (s *StampService) DeleteStamp(id string) error {
       // Get stamp to find image path
       stamp, err := s.GetStampByID(id)
       if err != nil {
           return err
       }

       // Remove image if exists
       if stamp.ImageURL != nil && *stamp.ImageURL != "" {
           imagePath := filepath.Join("./static", *stamp.ImageURL)
           if err := os.Remove(imagePath); err != nil {
               log.Printf("Warning: Could not delete image %s: %v", imagePath, err)
               // Don't fail deletion if image removal fails
           }

           // Also remove backup if exists
           backupPath := imagePath + ".bak"
           os.Remove(backupPath) // Ignore error
       }

       // Proceed with soft delete
       _, err = s.db.Exec(`UPDATE stamps SET date_deleted = $1 WHERE id = $2`,
           time.Now(), id)
       return err
   }
   ```

2. Add cleanup job for orphaned images (optional):
   - Scan `static/images/stamps/` directory
   - Check if each image has corresponding stamp
   - Remove orphans older than 30 days

**Success Criteria**: No disk space accumulation from deleted stamps

---

### 12. Template Injection Risk
- **Priority**: MEDIUM
- **Files**: `templates/instance-row.html:47`, other templates with `hx-on`
- **Status**: ⬜ Not Started
- **Estimated Effort**: 2 hours

**Issue**: Inline JavaScript in `hx-on` attributes not escaped.

**Implementation Steps**:
1. Replace `hx-on::after-request` with HTMX event listeners:
   ```html
   <!-- Before: -->
   <button hx-on::after-request="if(event.detail.xhr.status === 204) this.closest('tr').remove();">

   <!-- After: -->
   <button data-delete-row="true">
   ```

2. Add global HTMX event handler in `templates/index.html`:
   ```html
   <script>
   document.body.addEventListener('htmx:afterRequest', function(event) {
       if (event.detail.xhr.status === 204 &&
           event.target.hasAttribute('data-delete-row')) {
           event.target.closest('tr').remove();
       }
   });
   </script>
   ```

3. Audit all templates for `hx-on` usage

4. Create safer alternatives for common patterns

**Success Criteria**: No inline JavaScript in templates

---

### 13. Cache Busting Improvement
- **Priority**: MEDIUM
- **Files**: `templates/stamp-image-section.html:5`, `internal/handlers/htmx.go`
- **Status**: ⬜ Not Started
- **Estimated Effort**: 1 hour

**Issue**: DateModified might not change if update happens in same second.

**Implementation Steps**:
1. Update stamp's DateModified in microseconds on image upload:
   ```go
   // In UploadStampImage after successful save
   _, err = h.db.Exec(`UPDATE stamps SET
       image_url = $1,
       date_modified = NOW()
       WHERE id = $2`, imageURL, stampID)
   ```

2. Alternative: Use random cache buster:
   ```html
   <img src="{{deref .Stamp.ImageURL}}?v={{randomString}}">
   ```

3. Or use file modification time:
   ```go
   fileInfo, _ := os.Stat(imagePath)
   cacheTime := fileInfo.ModTime().Unix()
   ```

**Success Criteria**: Images always refresh after upload

---

## 🔧 LOW PRIORITY - Technical Debt

### 14. Remove Redundant HTTP Method Checks
- **Priority**: LOW
- **Files**: `internal/handlers/htmx.go` (multiple locations)
- **Status**: ⬜ Not Started
- **Estimated Effort**: 30 minutes

**Implementation Steps**:
1. Remove manual `r.Method` checks where router already restricts via `.Methods()`
2. Or document why manual checks are needed
3. Be consistent across all handlers

**Success Criteria**: No redundant method checking

---

### 15. Extract Magic Numbers
- **Priority**: LOW
- **Files**: `internal/handlers/htmx.go:555, 573`
- **Status**: ⬜ Not Started
- **Estimated Effort**: 15 minutes

**Implementation Steps**:
1. Add constants at top of `htmx.go`:
   ```go
   const (
       MaxUploadSize = 5 << 20 // 5MB
       ImagesDir     = "./static/images/stamps"
   )
   ```

2. Replace all hardcoded values

**Success Criteria**: All magic numbers extracted to constants

---

### 16. Reduce Logging Noise
- **Priority**: LOW
- **Files**: Multiple handlers
- **Status**: ⬜ Not Started
- **Estimated Effort**: 2 hours

**Implementation Steps**:
1. Implement log levels (use `log/slog` or similar):
   ```go
   log.Debug("handlers.htmx.UpdateBoxName: %+v", box)  // Only in dev
   log.Info("Stamp created: %s", stampID)
   log.Error("Failed to create stamp: %v", err)
   ```

2. Add environment variable for log level

3. Remove debug logs that expose sensitive data

**Success Criteria**: Clean production logs, detailed debug logs

---

### 17. Implement Prepared Statements
- **Priority**: LOW
- **Files**: `internal/services/*.go`
- **Status**: ⬜ Not Started
- **Estimated Effort**: 4-6 hours

**Issue**: Frequently-executed queries would benefit from prepared statements.

**Implementation Steps**:
1. Identify hot paths (use profiling)
2. Create prepared statements in service constructors
3. Use prepared statements for frequent queries
4. Benchmark performance improvement

**Success Criteria**: Measurable performance improvement on hot paths

---

### 18. Add Database Indexes
- **Priority**: LOW (depends on production usage patterns)
- **Files**: `internal/database/migrations.go`
- **Status**: ⬜ Not Started
- **Estimated Effort**: 2 hours

**Implementation Steps**:
1. Add indexes to migration:
   ```sql
   CREATE INDEX IF NOT EXISTS idx_stamps_date_deleted ON stamps(date_deleted);
   CREATE INDEX IF NOT EXISTS idx_stamps_scott_number ON stamps(scott_number);
   CREATE INDEX IF NOT EXISTS idx_stamp_instances_stamp_id ON stamp_instances(stamp_id);
   CREATE INDEX IF NOT EXISTS idx_stamp_instances_box_id ON stamp_instances(box_id);
   CREATE INDEX IF NOT EXISTS idx_stamp_tags_stamp_id ON stamp_tags(stamp_id);
   CREATE INDEX IF NOT EXISTS idx_stamp_tags_tag_id ON stamp_tags(tag_id);
   ```

2. Test query performance before/after
3. Monitor index usage in production

**Success Criteria**: Faster queries on filtered/joined data

---

## 🚀 LONG-TERM IMPROVEMENTS

### 19. Comprehensive Test Suite
- **Priority**: LONG-TERM
- **Status**: ⬜ Not Started
- **Estimated Effort**: 20+ hours

**Areas to Test**:
- [ ] Unit tests for all services
- [ ] Handler tests for all endpoints
- [ ] Integration tests for HTMX flows
- [ ] Security tests (SQL injection, XSS, CSRF, file upload)
- [ ] Performance tests (N+1 queries, large datasets)
- [ ] Concurrent operation tests

---

### 20. Content Security Policy
- **Priority**: LONG-TERM
- **Status**: ⬜ Not Started
- **Estimated Effort**: 4 hours

**Implementation Steps**:
1. Add CSP middleware
2. Configure policy headers
3. Test with browser DevTools
4. Add nonces for inline scripts

---

### 21. Database Migration Framework
- **Priority**: LONG-TERM
- **Status**: ⬜ Not Started
- **Estimated Effort**: 8 hours

**Options to Evaluate**:
- golang-migrate
- goose
- sql-migrate

---

### 22. Cloud Image Storage
- **Priority**: LONG-TERM
- **Status**: ⬜ Not Started
- **Estimated Effort**: 12+ hours

**Consider**:
- S3 integration
- CDN for image serving
- Image optimization pipeline

---

## Progress Tracking

### Phase 1: Critical Security (Before Merge)
- **Items**: 1-5
- **Estimated Total**: 9-12 hours
- **Actual Total**: 6 hours
- **Completion Date**: 2025-11-13
- **Completed**: 5/5 ✅

### Phase 2: High Priority (Before/Shortly After Merge)
- **Items**: 6-9
- **Estimated Total**: 8-9 hours
- **Actual Total**: 4 hours
- **Completion Date**: 2025-11-13
- **Completed**: 4/4 ✅

### Phase 3: Medium Priority (Post-Merge)
- **Items**: 10-13
- **Estimated Total**: 6-8 hours
- **Target Date**: [To be set]
- **Completed**: 0/4

### Phase 4: Low Priority (Technical Debt)
- **Items**: 14-18
- **Estimated Total**: 8-12 hours
- **Target Date**: [To be set]
- **Completed**: 0/5

### Phase 5: Long-Term (Ongoing)
- **Items**: 19-22
- **Estimated Total**: 40+ hours
- **Target Date**: [To be set]
- **Completed**: 0/4

---

## Notes

- Keep this file updated as work progresses
- Mark items as complete by changing ⬜ to ✅
- Add actual completion dates
- Note any blockers or dependencies
- Update estimates based on actual time spent
