# Defensive Dashboard - Refactored Component Structure

This dashboard has been refactored from a single 1198-line file into smaller, maintainable components for easier development and debugging.

## 📂 File Structure

```
defensive-dashboard/
├── +page.svelte              # Main page (now ~260 lines)
├── types.ts                  # TypeScript type definitions
├── StatsHeader.svelte        # Header with statistics
├── CameraSelector.svelte     # Camera selection sidebar
├── DetectionCard.svelte      # Individual detection card
└── README.md                 # This file
```

## 🧩 Component Breakdown

### 1. **types.ts** (~50 lines)
**Purpose:** Centralized TypeScript type definitions

**Exports:**
- `Camera` - Camera entity with id, name, location, institute
- `Detection` - Detection event with camera_id, timestamp, objects, image
- `DetectedObject` - Detected object with class_name and confidence
- `Pagination` - Pagination state (page, limit, total, totalPages)
- `CameraListResponse` - API response structure for camera list

**Why separate?** 
- Type safety across all components
- Single source of truth for data structures
- Easy to update when API changes

---

### 2. **StatsHeader.svelte** (~200 lines)
**Purpose:** Dashboard header with live statistics and connection controls

**Props:**
- `selectedCamerasCount: number` - Number of selected cameras
- `detectionsCount: number` - Total detections received
- `activeConnectionsCount: number` - Active WebSocket connections
- `onDisconnectAll: () => void` - Callback to disconnect all cameras

**Features:**
- Animated logo with floating effect
- Real-time stat displays (cameras/detections)
- Conditional "Disconnect All" button (only shows when cameras selected)
- Connection status indicator with pulse animation
- Gradient background matching theme

**Key Styles:**
- Floating logo animation
- Pulse animation for connection indicator
- Responsive stat cards
- Smooth transitions

---

### 3. **CameraSelector.svelte** (~400 lines)
**Purpose:** Sidebar component for camera search, selection, and pagination

**Props:**
- `cameras: Camera[]` - Array of available cameras
- `selectedCameraIds: Set<string>` - Set of selected camera IDs
- `searchName: string` (bindable) - Search query value
- `isLoading: boolean` - Loading state
- `pagination: Pagination` - Pagination state
- `wsConnections: Map<string, WebSocket>` - Active WebSocket connections
- `onToggleCamera: (cameraId: string) => void` - Toggle camera selection
- `onSearch: () => void` - Trigger search
- `onSearchChange: (value: string) => void` - Update search value
- `onNextPage: () => void` - Go to next page
- `onPrevPage: () => void` - Go to previous page

**Features:**
- Search box with Enter key support
- Multi-select camera cards with checkboxes
- Visual indicators for selected/connected cameras
- Loading spinner during fetch
- Empty state when no cameras found
- Pagination controls with page info
- Scrollable camera list

**Key Styles:**
- Selected camera: purple gradient with border
- Connected camera: green status dot with pulse
- Hover effects: translateX(2px)
- Custom scrollbar styling

---

### 4. **DetectionCard.svelte** (~180 lines)
**Purpose:** Reusable card for displaying individual detections

**Props:**
- `detection: Detection` - Detection object to display
- `isSelected: boolean` - Whether this detection is currently selected
- `cameraName: string` - Display name of the camera
- `onClick: () => void` - Click handler for selection

**Features:**
- Thumbnail image with fallback placeholder
- Object badge showing detection count
- Camera name with icon
- Formatted timestamp
- Object tags (shows first 3, "+X more" for rest)
- Visual selection state

**Key Styles:**
- Selected card: purple gradient with elevated shadow
- Hover effect: translateY(-2px) with shadow
- Object badge: red badge with absolute positioning
- Thumbnail: 80x80px with border-radius

---

### 5. **+page.svelte** (Main Page, ~260 lines)
**Purpose:** Application logic and layout orchestration

**Responsibilities:**
- State management (cameras, detections, selections, WebSocket connections)
- API calls (`fetchCameras` with pagination)
- WebSocket connection management (connect/disconnect per camera)
- Auto-reconnect logic with 3-second timeout
- Event handlers (search, pagination, selection)
- Component composition

**Key Functions:**
- `fetchCameras(page, search)` - GET /api/v1/camera
- `toggleCameraSelection(cameraId)` - Multi-select with auto-connect/disconnect
- `connectCamera(cameraId)` - Create WebSocket connection
- `disconnectCamera(cameraId)` - Close WebSocket and cleanup
- `disconnectAllCameras()` - Cleanup all connections
- `getCameraName(cameraId)` - Helper for display

**Layout:**
- `StatsHeader` - Top header
- `CameraSelector` (30%) + Map (70%) - Main content
- Horizontal detection list - Footer

**State Management:**
```typescript
cameras: Camera[]
selectedCameraIds: Set<string>
detections: Detection[]
wsConnections: Map<string, WebSocket>
reconnectTimeouts: Map<string, any>
```

---

## 🔄 WebSocket Flow

1. User selects camera in `CameraSelector`
2. `onToggleCamera` calls `toggleCameraSelection(cameraId)`
3. `connectCamera(cameraId)` creates new WebSocket
4. WebSocket onopen → sends `{camera_id: cameraId}`
5. WebSocket onmessage → parses detection and adds to `detections[]`
6. `DetectionCard` components automatically re-render
7. WebSocket onclose → auto-reconnect after 3s if still selected

## 🎨 Design System

**Colors:**
- Primary: `#667eea` (purple)
- Danger: `#ef4444` (red)
- Success: `#10b981` (green)
- Gray scale: `#f9fafb`, `#f3f4f6`, `#e5e7eb`, `#d1d5db`, `#9ca3af`, `#6b7280`

**Animations:**
- Float: Logo floating effect (3s)
- Pulse: Connection indicator (2s)
- Spin: Loading spinner (1s)

**Transitions:**
- All interactive elements: `transition: all 0.2s`
- Hover: `transform: translateY(-2px)` or `translateX(2px)`

## 🛠️ Development Tips

### Adding a new feature to camera selector:
Edit `CameraSelector.svelte` - all camera selection logic is isolated there.

### Changing detection display:
Edit `DetectionCard.svelte` - affects all detection cards uniformly.

### Modifying API responses:
Update `types.ts` first, then TypeScript will guide you to fix all usages.

### Debugging WebSocket:
Check `+page.svelte` → `connectCamera()` function. All WebSocket logic is centralized there.

## ✅ Benefits of Refactoring

1. **Maintainability**: Each component has single responsibility
2. **Reusability**: `DetectionCard` can be used anywhere
3. **Type Safety**: Centralized types prevent bugs
4. **Testability**: Components can be tested in isolation
5. **Readability**: ~260 lines vs 1198 lines in main file
6. **Debugging**: Easy to locate bugs (camera issues → CameraSelector.svelte)

## 📦 Component Sizes

- `+page.svelte`: 1198 lines → **260 lines** (78% reduction)
- `types.ts`: **50 lines**
- `StatsHeader.svelte`: **200 lines**
- `CameraSelector.svelte`: **400 lines**
- `DetectionCard.svelte`: **180 lines**

**Total:** 1090 lines across 5 files (vs 1198 in 1 file)

## 🚀 Future Enhancements

Potential component extractions:
- `MapView.svelte` - Encapsulate map logic
- `PaginationControls.svelte` - Reusable pagination
- `SearchBox.svelte` - Reusable search input
- `EmptyState.svelte` - Reusable empty state

---

**Last Updated:** December 2024
**UI Status:** ✅ Fully functional, maintains all original features
**Code Quality:** ✅ Refactored for maintainability
