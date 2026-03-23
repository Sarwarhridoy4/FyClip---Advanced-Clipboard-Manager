# FyClip - Scopes of Improvements

This document outlines the specific issues and planned fixes for confusing menus, crashes, and memory optimization.

---

## 1. Confusing Menu Issues

### Issue 1.1: Inconsistent Toolbar Button Labels
**Location**: [`internal/ui/toolbar.go:51-64`](internal/ui/toolbar.go:51)

**Problem**: 
- "Pin/Unpin" button label is confusing - doesn't indicate current state
- "Favorites" icon without clear purpose
- Snippets button shows but is non-functional (`_ = snippetsBtn`)
- Pause button shows "Pause 5m" but doesn't indicate toggle state clearly

**Fix**:
- Change "Pin/Unpin" to "Pin" (enable) / "Unpin" (disable) based on selection
- Rename favorites button to "Show Pinned" / "Show All"
- Enable snippets functionality or remove the button
- Make pause button text more explicit: "Pause" / "Resume"

### Issue 1.2: Tray Menu Lacks Clear Organization
**Location**: [`internal/tray/tray.go:73-97`](internal/tray/tray.go:73)

**Problem**:
- Recent items shown without clear categorization
- "Clear History" is dangerous but easily accessible
- AutoStart toggle label doesn't reflect current state clearly

**Fix**:
- Add separator between recent items and actions
- Add confirmation dialog for Clear History
- Show checkmark next to enabled AutoStart

### Issue 1.3: Context Menu Missing on List Items
**Location**: [`internal/ui/list.go`](internal/ui/list.go)

**Problem**:
- Right-click context menu not implemented
- Users expect quick actions on right-click

**Fix**:
- Implement right-click context menu with: Copy, Pin/Unpin, Delete, Export

---

## 2. Crash Issues

### Issue 2.1: Race Condition in Index Map Updates
**Location**: [`internal/clipboard/manager.go:278-308`](internal/clipboard/manager.go:278)

**Problem**:
- `rebuildIndexMapsFrom` is called after each deletion but may not handle all edge cases
- Potential race condition when multiple goroutines access history simultaneously

**Fix**:
```go
// Add atomic operations and proper locking
func (m *Manager) rebuildIndexMapsFrom(startIdx int) {
    m.mu.Lock()  // Ensure exclusive access
    defer m.mu.Unlock()
    // ... existing code
}
```

### Issue 2.2: Monitor Loop Panic Recovery Without Proper Cleanup
**Location**: [`internal/clipboard/monitor.go:136-164`](internal/clipboard/monitor.go:136)

**Problem**:
- Panic recovery attempts to restart loop but may leave resources in bad state
- No limit on restart attempts could cause infinite restart loop

**Fix**:
```go
const maxPanicRestarts = 3

func (m *Monitor) monitorLoop() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic in monitor loop: %v", r)
            m.mu.Lock()
            restartCount := m.restartCount
            m.restartCount++
            running := m.running
            m.mu.Unlock()
            
            if running && restartCount < maxPanicRestarts {
                time.Sleep(1 * time.Second)
                go m.monitorLoop()
            } else {
                log.Printf("Monitor loop stopped after %d panic restarts", restartCount)
            }
        }
    }()
    // ... existing code
}
```

### Issue 2.3: Storage Save Failure Silently Ignored
**Location**: [`internal/clipboard/manager.go:186-203`](internal/clipboard/manager.go:186)

**Problem**:
- `saveHistoryNow` logs errors but doesn't notify user
- Failed saves could lead to data loss

**Fix**:
- Add retry mechanism for failed saves
- Notify user when save fails after multiple attempts

### Issue 2.4: Nil Pointer Dereference in Tray
**Location**: [`internal/tray/tray.go:100-134`](internal/tray/tray.go:100)

**Problem**:
- If manager is nil, `GetFiltered()` could panic
- Recent items could contain nil items

**Fix**:
```go
func (st *SystemTray) buildRecentMenuItems() []*fyne.MenuItem {
    items := []*fyne.MenuItem{}
    
    if st.manager == nil {
        return append(items, fyne.NewMenuItem("No items", nil))
    }
    
    history := st.manager.GetFiltered()
    if history == nil || len(history) == 0 {
        return append(items, fyne.NewMenuItem("No items", nil))
    }
    // ... rest of the code
}
```

---

## 3. Memory Optimization Issues

### Issue 3.1: Base64 Image Storage Overhead
**Location**: [`internal/clipboard/monitor.go:275-281`](internal/clipboard/monitor.go:275)

**Problem**:
- Images stored as base64 strings take ~33% more memory than raw bytes
- No lazy loading - all images loaded into memory

**Fix**:
- Store raw image bytes instead of base64
- Implement lazy loading for image preview
- Add LRU cache for frequently accessed images

### Issue 3.2: Unnecessary Copy in Filtered List
**Location**: [`internal/clipboard/manager.go:345-366`](internal/clipboard/manager.go:345)

**Problem**:
- `updateFiltered` creates copies of items unnecessarily
- Double memory usage during filtering

**Fix**:
- Use pointers instead of copies in filtered list
```go
func (m *Manager) updateFiltered() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Use pointers to avoid copying
    m.filtered = m.filtered[:0]
    
    // ... rest of code using &m.history[i] instead of copy
}
```

### Issue 3.3: No Memory Pressure Handling
**Location**: [`internal/clipboard/manager.go`](internal/clipboard/manager.go)

**Problem**:
- No automatic memory cleanup when system is under pressure
- History grows unbounded until max items reached

**Fix**:
```go
// Add memory monitoring
func (m *Manager) checkMemoryPressure() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // If heap usage > 100MB, reduce history size
    if m.Alloc > 100*1024*1024 {
        m.trimHistory()
    }
}
```

### Issue 3.4: Sync Pool Not Utilized
**Location**: [`internal/clipboard/pool.go`](internal/clipboard/pool.go)

**Problem**:
- Object pool exists but is never used in the codebase
- Creates overhead without benefit

**Fix**:
- Integrate pool usage in item creation/deletion
- Or remove unused pool code

---

## 4. Summary of Fixes

| Issue | Severity | Status |
|-------|----------|--------|
| Toolbar button labels confusing | Medium | Pending |
| Tray menu organization | Low | Pending |
| Race condition in index maps | High | Pending |
| Monitor panic restart loop | High | Pending |
| Storage save failure handling | Medium | Pending |
| Nil pointer in tray | Medium | Pending |
| Image base64 overhead | Medium | Pending |
| Unnecessary copies in filtered | Medium | Pending |
| Memory pressure handling | Medium | Pending |
| Unused object pool | Low | Pending |

---

## Implementation Plan

1. **Phase 1 - Crash Fixes** (High Priority)
   - Fix race conditions
   - Add panic restart limits
   - Add nil checks
   
2. **Phase 2 - Menu Improvements** (Medium Priority)
   - Improve toolbar labels
   - Add context menus
   - Improve tray organization

3. **Phase 3 - Memory Optimization** (Low Priority)
   - Optimize image storage
   - Add memory pressure handling
   - Remove unused code
