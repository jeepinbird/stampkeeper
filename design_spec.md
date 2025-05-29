# Design Specification: StampKeeper

## 1. Introduction

This document outlines the design and functionality for a web application aimed at helping stamp collectors manage their collections. The primary target audience is older adults; therefore, the design prioritizes usability, clarity, high readability, and accessibility. The aesthetic will be clean and modern, utilizing a strictly monochrome color scheme with high contrast.

## 2. Overall Design Principles

- **Simplicity & Clarity:** Prioritize clear, straightforward navigation and workflows. Avoid overwhelming users with too many options or complex interfaces. Language used should be simple and direct.
- **Font Sizes:** Default large font sizes (e.g., 16px for body text, 20px+ for headings) with user-adjustable options.
- **Generous Whitespace:** Ample whitespace is crucial to improve readability, reduce visual clutter, and help delineate sections. No borders for sections, though.
- **Monochrome Theme:** The color palette will be strictly limited to shades of black, white, and gray.
	- **Primary Backgrounds:** Off-white (e.g., `#F8F9FA`) for light mode, Dark Gray (e.g., `#212529`) for dark mode.
	- **Primary Text:** Dark Gray (e.g., `#343A40`) for light mode, Light Gray (e.g., `#E9ECEF`) for dark mode.
	- **Accent/Highlight (Subtle):** Medium grays for borders, inactive elements, or subtle highlights (e.g., `#ADB5BD`, `#6C757D`).
	- **Focus Indicators:** A distinct, high-contrast outline (e.g., a darker gray or a slightly thicker border).
- **Intuitive UI:** The interface should be easily understandable, predictable, and require a minimal learning curve. Icons should be accompanied by text labels.
- **Forgiving Design:** Provide clear confirmation messages for destructive actions (e.g., deleting a stamp) and allow for undo where feasible (just mark the record as deleted in the database).

## 3. Core Functionality & Page Layout

The application will be a Single Page Application (SPA) to ensure smooth transitions and a consistent user experience. A persistent header will remain visible which includes a centered search bar. A footer is optional and can be considered if there's a clear need (e.g., copyright, version number).

### 3.1 Header (Consistent across all views)

- **Left Side:**    
    - **Application Title:** Stamp logo icon followed by "StampKeeper" in a large, clear font.
- **Center**:
	- **Search Bar:**
		- **Appearance:** Prominent, full-width (of the sidebar). Clear input field with a visible "Search" button or icon.
		- **Functionality:** Allows searching by Stamp Name, Scott Number, Year, and Tags. Search results update the main content area (Gallery/List View).
- **Right Side:**
	- **User Account/Settings Icon:** A clear, universally recognized icon (e.g., gear `⚙️` or profile `👤`) with a text label "Settings".
		- **Leads to Settings Modal/Panel:**
			- **Theme Switcher:** (See section 6.1)
			- **Font Size Adjustment:** (See section 6.2)

### 3.2 Main Content Area

This area will dynamically display content based on user interaction. It will primarily house the stamp collection browser, stamp details, and dashboard.

### 3.3 Sidebar (Persistent on wider screens, collapsible or off-canvas on smaller screens)

- **Optionally Collapsible:** Persistent sidebar is unless cluttering the interface on screens ~1024px wide. If cluttered, a collapsible or hamburger menu approach for sidebar content should be implemented for narrower views.
- **Quick Filters:**
	- **Appearance:** Clearly labeled buttons or toggle switches.
	    - **Options:**
			- "All Stamps" (Default)
			- "Owned"
			- "Needed"
	- **Behavior:** Filters are mutually exclusive for "Owned" and "Needed".
- **Sort Dropdown:**
	- **Appearance:** Standard dropdown menu.
	- **Label:** "Sort By:"
	- **Options:**
		- Scott Number
		- Stamp Name (A-Z / Z-A)
		- Issue Date (Ascending/Descending)
		- Date Added (Newest/Oldest)
		- Series 
- **Storage Box Navigation:**
	- **Display:** A list of user-created Storage Box names.
	- **Interaction:** Clicking a box name filters the main content area to show only stamps within that box. A "Clear Filter" or "Show All" option should be visible when a box filter is active.
	- **"Add New Box" Button:**
		- **Placement:** Clearly visible at the bottom or top of the storage box list.
		- **Action:** Opens a simple modal prompting for "Box Name/Identifier".

### 3.4 View Toggles (Main Content Area - above Gallery/List)

- **Appearance:** Two clearly labeled buttons or segmented controls: "Gallery View" and "List View". The active view should be visually distinct.
- **Gallery View:**
	- **Layout:** Displays stamps as a responsive grid of image cards.
	- **Image Size:** Sufficiently large for easy recognition, with consideration for older eyes.
	- **Information on Hover/Focus (Optional):** Scott Number, Stamp Name, Issue Year.
	- **Action:** Clicking a stamp image/card navigates to the Stamp Details View.
- **List View:**
	- **Layout:** Displays stamps as a table or a vertical list.
	- **Columns/Information per item:**
		- Thumbnail Image (Small, clear)
		- Scott Number (Clickable, navigates to Stamp Details View)
		- Stamp Name (Clickable, navigates to Stamp Details View)
		- Issue Date
		- Quantity (If multiple identical copies are owned)
		- Storage Box Name (If assigned)
	- **Action:** Clicking the Scott number or Stamp Name navigates to the Stamp Details View.

## 4. Detailed Features

### 4.1 Dashboard/Stats (Initial Landing Page or accessible via a clear "Dashboard" link in Header/Sidebar if space allows)

- **Layout:** Clean, spacious, with large, easy-to-read text for statistics.
- **Key Statistics:**
	- **Total Owned Stamps:** Large, prominent display.
	- **Number of Unique Stamps:** Displays the count of distinct stamp issues (based on Scott Number).
	- **Number of Stamps Needed:** Displays count of stamps marked as "Needed".
	- **Number of Storage Boxes:** Displays count of created boxes.
- **(Future Consideration):** Missing stamps from defined sets.

### 4.2 Storage Box Management

- **Access:** Via Sidebar or a dedicated "Manage Boxes" page accessible from Settings or Dashboard.
- **List of Boxes:**
	- **Display:** Box Name.
	- **Summary:** Below each box name, display the number of stamps it contains (e.g., "Box A: 25 stamps").
- **Add New Box:**
	- **Trigger:** Button in Sidebar or on "Manage Boxes" page.
	- **Process:**
		1. User clicks "Add New Box".
		2. Modal appears prompting for "Box Name" (text input).
		3. User enters name and clicks "Save" or "Create".
		4. New box is added to the list and database.
- **Edit Box Name:**
	- **Trigger:** An "Edit" icon next to each box name in the list (on "Manage Boxes" page or right-click/context menu in sidebar).
	- **Process:** Modal pre-filled with current name, user edits and saves.
- **Delete Box:**
	- **Trigger:** A "Delete" icon next to each box name.
	- **Process:** Confirmation modal ("Are you sure you want to delete '[Box Name]'? Stamps within this box will NOT be deleted but will become unassigned."). If confirmed, box is deleted. Stamps previously in this box should have their `storage_box_id` field cleared.
- **Print Summary Sheet (Per Box):**
	- **Trigger:** A "Print Summary" button visible when viewing stamps filtered by a specific box, or on a "Box Details" page (if implemented).
	- **Output:** Generates a simple, printer-friendly HTML page listing stamps in that box (Thumbnail, Name, Scott No., Issue Date, Quantity). Should include Box Name and Date Printed.

### 4.3 Stamp Details View

- **Access:** Clicking a stamp in Gallery or List view.
- **Layout:** Two-column layout on wider screens (image on one side, details on the other) or stacked on smaller screens.
- **Displayed Information (Read-Only Mode):**
	- Large Stamp Image (If available, otherwise a placeholder)
	- Stamp Name
	- Scott Number
	- Issue Date
	- Series (If applicable)
	- Condition (e.g., Mint, Used, Damaged)
	- Quantity (If multiple identical copies)
	- Storage Box (Displays name; clickable to change if in edit mode or via a dedicated "Move" button)
	- Tags (Displayed as pills/badges)
	- Notes (Multi-line text display)
	- Owned/Needed Status (Clearly indicated; simply a boolean checkbox. If checked, that it's owned, if not, it's needed)
- **"Edit Stamp" Button:**
	- **Action:** Switches the view to an editable form, pre-filled with existing details.
- **Editable Fields (When in "Edit Mode" or "Add New Stamp" form):**
	- Stamp Name (Text input)
	- Scott Number (Text input)
	- Issue Date (Date picker, manual entry friendly)
	- Series (Text input)
	- Condition (Dropdown: Mint, Used, Damaged, Other - user can define "Other")
	- Quantity (Number input, default 1)
	- Storage Box (Dropdown list of existing boxes, plus "None" or "Unassigned")
	- Tags (Text input with typeahead for existing tags; allow creating new tags by typing and hitting enter/comma. Displayed as removable pills.)
	- Notes (Textarea)
	- Image URL (Text input for external image link - actual upload is future)
	- Favorite (Checkbox/Toggle)
	- Owned/Want List (Radio buttons or Toggle: "I Own This Stamp" / "I Want This Stamp")
- **Buttons in Edit Mode:** "Save Changes", "Cancel", "Delete Stamp" (with confirmation).

### 4.4 Tags/Labels Management

- **Creation:** Tags are created organically when adding/editing a stamp. Typing a new tag and saving the stamp adds it to the global list of tags.
- **Application:** Users can apply multiple tags to a single stamp.
- **Tag Filtering:**
	- **Mechanism:** Clicking a tag pill/badge (in Stamp Details view or the "Filter by Tag" section in the sidebar) filters the main stamp view.
	- Multiple tags can be selected for filtering (e.g., stamps with "Rare" AND "USA").
- **Tag Management Page (Accessible from Settings):**
	- **List all unique tags.**
	- **Functionality:**
		- Rename Tag (Updates tag on all associated stamps).
		- Delete Tag (Removes tag from all associated stamps and from the global list. Confirmation needed).
		- See count of stamps using each tag.

## 5. Utility Buttons/Features

- **"Add New Stamp" Button:**
	- **Placement:** Prominent, fixed/floating button (e.g., bottom-right with a `+` icon and "Add Stamp" label on hover/focus). Consistent location is key.
	- **Action:** Opens the "Add New Stamp" form (identical in fields to the Stamp Details edit mode, but blank).

## 6. User Customization (Accessed via Settings Icon in Header)

- **Theme Switcher:**
	- **Options:**
		- **Light Mode:** Off-white/light gray background, dark gray/black text. (Default)
		- **Dark Mode:** Dark gray/near-black background, light gray/white text.
	- **Mechanism:** Simple toggle switch or radio buttons.
	- **Persistence:** User's choice should be saved (e.g., in `localStorage` or user profile if accounts are added later).
- **Font Size Adjustment:**
	- **Options:** Dropdown with predefined choices (e.g., "Standard", "Large", "Extra Large") OR a slider.
	- **Impact:** Adjusts the base font size for the entire application.
	- **Accessibility Focus:** Ensure this significantly impacts readability for those needing larger text.
	- **Persistence:** User's choice should be saved.

## 7. Technical Considerations

- **Frontend:** Uses HTMX for dynamic interactions, Alpine.js for lightweight JavaScript (if needed), and Bootstrap CSS for styling
- **Backend**: Go, handling server-side rendering and API logic, with DuckDB for data storage
- **Responsive Design:** Must be fully responsive and usable on desktop, tablet (landscape/portrait), and mobile (portrait primarily). Test thoroughly on various screen sizes.
- **Data Storage:** DuckDB server-side database.
	- **Persistence:** The DuckDB database will be stored in the server file system for persistence across sessions.
	- **Schema (`stamps` table):**
		- `id`: TEXT PRIMARY KEY (UUID)
		- `name`: TEXT NOT NULL
		- `scott_number`: TEXT
		- `issue_date`: TEXT (ISO 8601 Format: YYYY-MM-DD)
		- `series`: TEXT
		- `condition`: TEXT
		- `quantity`: INTEGER DEFAULT 1
		- `box_id`: TEXT (FOREIGN KEY to `storage_boxes.id`, can be NULL)
		- `notes`: TEXT
		- `image_url`: TEXT
		- `is_owned`: BOOLEAN
		- `date_added`: TEXT (ISO 8601 Timestamp)
		- `date_modified`: TEXT (ISO 8601 Timestamp)
	- **Schema (`stamp_tags` table - for many-to-many relationship):**
		- `stamp_id`: TEXT (FOREIGN KEY to `stamps.id`)
		- `tag_id`: TEXT (FOREIGN KEY to `tags.id`)
		- PRIMARY KEY (`stamp_id`, `tag_id`)
	- **Schema (`tags` table):**
		- `id`: TEXT PRIMARY KEY (UUID or normalized tag name)
		- `name`: TEXT UNIQUE NOT NULL
	- **Schema (`boxes` table):**
		- `id`: TEXT PRIMARY KEY (UUID)
		- `name`: TEXT UNIQUE NOT NULL
		- `date_created`: TEXT (ISO 8601 Timestamp)
	- **Data Export/Import:** simple CSV export/import feature for user data backup/migration

## 8. Non-Functional Requirements

- **Usability:** Application must be highly usable for older adults with varying levels of technical proficiency.
- **Performance:** Pages and interactions should feel responsive. List/gallery views with many stamps should load efficiently. Infinite scroll, load as you go.
- **Maintainability:** Code should be well-organized, commented, and easy to understand for future development.
- **Browser Compatibility:** Support latest versions of major browsers (Chrome, Firefox, Edge, Safari).