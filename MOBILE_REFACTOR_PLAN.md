# StampKeeper Mobile Refactor Plan

**Goal:** Transform StampKeeper into a mobile-first application optimized for elderly users with accessibility zoom enabled on Android phones.

**Key Constraints:**
- Primary user: Elderly dad with Android phone
- Screen is often zoomed in for better readability
- Limited screen real estate
- Must support ALL primary tasks: browsing, adding stamps, managing boxes, searching/filtering

---

## Phase 1: Foundation & Layout (Mobile-First Architecture)

### 1.1 Mobile Navigation System
- [ ] Replace desktop sidebar with mobile-first navigation
  - [ ] Add hamburger menu button in top header
  - [ ] Make sidebar collapsible/drawer-style on mobile
  - [ ] Implement bottom navigation bar for key actions (Browse, Add, Search, Settings)
  - [ ] Ensure 44px minimum touch target size for all nav items

### 1.2 Responsive Layout Updates
- [ ] Update CSS breakpoints for mobile-first approach
  - [ ] Add breakpoints: 320px (small phones), 480px (phones), 768px (tablets)
  - [ ] Make main container full-width on mobile (remove padding)
  - [ ] Stack all multi-column layouts vertically on mobile
  - [ ] Update grid systems for single-column mobile layouts

### 1.3 Typography & Readability
- [ ] Increase base font sizes for better zoom compatibility
  - [ ] Body text: 16px → 18px minimum
  - [ ] Headings: Increase by 1.2x
  - [ ] Buttons/form labels: 16px minimum
  - [ ] Ensure line-height of 1.5+ for all text
- [ ] Improve color contrast ratios (WCAG AA minimum)

---

## Phase 2: Touch-Friendly Interface

### 2.1 Interactive Elements
- [ ] Increase all touch target sizes to 44px minimum
  - [ ] Buttons (all views)
  - [ ] Form inputs
  - [ ] Links in lists
  - [ ] Radio buttons and checkboxes
  - [ ] Filter controls
- [ ] Add spacing between touch targets (minimum 8px gap)
- [ ] Replace hover states with touch-appropriate feedback

### 2.2 Forms Optimization
- [ ] **New Stamp Form** (new-stamp-form.html)
  - [ ] Convert to single-column layout on mobile
  - [ ] Increase input field height (48px minimum)
  - [ ] Use native mobile date picker
  - [ ] Larger "Create Stamp" and "Cancel" buttons (full-width on mobile)
  - [ ] Simplify inline editing (larger editable areas)

- [ ] **Stamp Instance Table** (your-copies-section.html)
  - [ ] Convert table to card-based layout on mobile
  - [ ] Larger quantity controls (+/- buttons)
  - [ ] Easier condition selection (larger dropdown/radio buttons)
  - [ ] Full-width action buttons

### 2.3 Floating Action Button
- [ ] Increase FAB size (60px → 72px on mobile)
- [ ] Ensure proper positioning with bottom navigation
- [ ] Add label for clarity ("Add Stamp")

---

## Phase 3: View-Specific Improvements

### 3.1 Gallery View
- [ ] Optimize card grid for mobile
  - [ ] Single column on smallest screens (320-480px)
  - [ ] Two columns on mid-size phones (480-768px)
  - [ ] Larger card images (better for zoomed viewing)
  - [ ] Increase card padding and spacing
- [ ] Simplify stamp card information (prioritize key details)

### 3.2 List View
- [ ] Replace table with mobile-friendly card list
  - [ ] Each stamp as expandable card
  - [ ] Show thumbnail, name, and Scott# by default
  - [ ] Tap to expand for more details
- [ ] Add horizontal scroll with clear indicators (if table kept)
- [ ] Sticky header row with essential columns only

### 3.3 Stamp Detail View
- [ ] Stack all sections vertically on mobile
- [ ] Larger image viewing area (full-width)
- [ ] Simplify editable fields
  - [ ] Larger click/tap areas for inline editing
  - [ ] Consider modal forms instead of inline editing
- [ ] Full-width action buttons
- [ ] Optimize "Your Copies" section as cards on mobile

### 3.4 Search & Filters
- [ ] Move search to prominent top position (always visible)
- [ ] Larger search input (56px height)
- [ ] Filters in collapsible panel or modal
- [ ] "Jump to Scott#" simplified or combined with search
- [ ] Quick filter chips (Owned, Needed, All) as large buttons

---

## Phase 4: Performance & Polish

### 4.1 Image Handling
- [ ] Implement responsive images (srcset for different sizes)
- [ ] Optimize image loading for mobile bandwidth
- [ ] Larger placeholder icons
- [ ] Consider lazy loading for gallery view

### 4.2 Gestures & Interactions
- [ ] Add swipe gestures for navigation (optional)
- [ ] Pull-to-refresh on main views (optional)
- [ ] Confirm dialogs for destructive actions (larger, clearer)

### 4.3 Settings Page
- [ ] Optimize settings form for mobile
- [ ] Larger toggle switches
- [ ] Group related settings in collapsible sections
- [ ] Full-width buttons

---

## Phase 5: Testing & Refinement

### 5.1 Testing Checklist
- [ ] Test on Android phone with various zoom levels (100%, 150%, 200%)
- [ ] Verify all touch targets are easily tappable
- [ ] Test form submissions on mobile
- [ ] Verify table/list scrolling behavior
- [ ] Test navigation flow (hamburger menu, bottom nav)
- [ ] Check image upload on mobile
- [ ] Test search and filtering

### 5.2 Accessibility
- [ ] Verify semantic HTML structure
- [ ] Add proper ARIA labels for mobile nav
- [ ] Ensure keyboard navigation works
- [ ] Test with Android TalkBack (screen reader)

### 5.3 User Feedback
- [ ] Have dad test the app on his phone
- [ ] Gather feedback on ease of use
- [ ] Identify pain points
- [ ] Iterate based on feedback

---

## Implementation Priority

**High Priority (Must Have):**
1. Mobile navigation system (1.1)
2. Responsive layout updates (1.2)
3. Touch-friendly interactive elements (2.1)
4. Forms optimization (2.2)
5. Gallery & list view improvements (3.1, 3.2)

**Medium Priority (Should Have):**
1. Typography improvements (1.3)
2. Stamp detail view optimization (3.3)
3. Search & filter improvements (3.4)
4. Floating action button updates (2.3)

**Lower Priority (Nice to Have):**
1. Advanced gestures (4.2)
2. Image optimization (4.1)
3. Performance improvements

---

## Notes

- All changes should be incremental and testable
- Maintain backward compatibility with desktop view
- Focus on simplicity and clarity for elderly user
- Prioritize common tasks (browsing and viewing stamps)
- Keep the "HTML over the wire" architecture (HTMX)
- Minimize JavaScript changes

---

## Success Criteria

✅ Dad can easily navigate the app on his zoomed Android phone
✅ All buttons and links are easily tappable (no mis-taps)
✅ Forms are simple to fill out on mobile
✅ Text is readable at high zoom levels
✅ No horizontal scrolling issues
✅ Gallery and list views work smoothly
✅ Adding stamps is intuitive on mobile
✅ Search and filtering are easy to use

---

**Last Updated:** 2025-11-12
**Status:** Planning Complete - Ready for Implementation
