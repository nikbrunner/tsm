# Plan: tsm-r2t - Session List Refactoring

## Investigation Summary

### Libraries Evaluated

| Library | Verdict | Reason |
|---------|---------|--------|
| **bubbles/table** | Not suitable | Flat `[]string` rows, no hierarchy, no expand/collapse |
| **Evertras/bubble-table** | Not suitable | More features but still flat row model |
| **tree-bubble** | Not suitable | Too simple: single `Value` string, no multi-column, no filtering |
| **bubbles/list** | Overkill | Would need custom ItemDelegate, adds abstraction without benefit |

### Why the current approach is correct

tsm's session list has requirements that don't fit standard components:
1. **Hierarchical**: Sessions contain windows (parent→child)
2. **Collapsible**: Sessions expand/collapse independently
3. **Multi-column**: 7 columns for sessions (num, last-icon, expand, name, time, git, claude)
4. **Different row types**: Sessions render differently than windows
5. **Fuzzy filtering**: Real-time filter as user types

### Real pain points in current code

The complexity isn't in rendering - it's in **duplication**:
- Three list modes (sessions, projects, clones) with near-identical scroll/cursor/filter logic
- `updateScrollOffset()`, `rebuildItems()`, filtering logic repeated per mode
- 93-line Model struct with scattered state for each mode

---

## Recommended Approach: Internal Refactoring

Instead of adopting an external library, clean up the existing implementation:

### Phase 1: Extract Scrollable List Helper

Create `internal/ui/scrolllist.go` with generic scroll/cursor/filter logic:

```go
type ScrollList[T any] struct {
    items        []T
    filtered     []T
    cursor       int
    scrollOffset int
    filter       string
    filterFn     func(T, string) bool
    height       int
}

func (s *ScrollList[T]) Filter(text string)
func (s *ScrollList[T]) MoveCursor(delta int)
func (s *ScrollList[T]) VisibleItems() []T
func (s *ScrollList[T]) SelectedItem() (T, bool)
func (s *ScrollList[T]) UpdateScrollOffset()
```

### Phase 2: Simplify Model State

Consolidate the three list modes into using the same helper:
- `sessionList ScrollList[Item]` replaces `items`, `cursor`, `scrollOffset`, `filter`
- `projectList ScrollList[string]` replaces `projectDirs`, `projectFiltered`, `projectCursor`, `projectScrollOffset`, `projectFilter`
- `cloneList ScrollList[string]` replaces `cloneRepos`, `cloneReposAll`, `cloneCursor`, `cloneScrollOffset`, `cloneFilter`

### Phase 3: Component-Style Row Rendering

Adopt a React-like pattern where each column is a "component" (function) and Row combines them:

**New file: `internal/ui/columns.go`**
```go
// RowLayout holds calculated widths for alignment
type RowLayout struct {
    NameWidth      int
    GitStatusWidth int
    ShowGitStatus  bool
}

// Column "components" - each returns a styled string
func RenderIndex(num int, selected bool) string
func RenderLastIcon(isFirst, selected bool) string
func RenderExpandIcon(expanded, selected bool) string
func RenderSessionName(name string, width int, selected bool) string
func RenderTimeAgo(t time.Time, selected bool) string
func RenderGitStatus(status git.Status, width int) string
func RenderClaudeStatus(status claude.Status, frame int) string

// RenderSessionRow combines columns with layout
func RenderSessionRow(s Session, layout RowLayout, opts RowOpts) string {
    cols := []string{
        RenderIndex(opts.Num, opts.Selected),
        RenderLastIcon(opts.IsFirst, opts.Selected),
        RenderExpandIcon(s.Expanded, opts.Selected),
        RenderSessionName(s.Name, layout.NameWidth, opts.Selected),
        RenderTimeAgo(s.LastActivity, opts.Selected),
    }
    if layout.ShowGitStatus {
        cols = append(cols, RenderGitStatus(opts.GitStatus, layout.GitStatusWidth))
    }
    if opts.ClaudeStatus != nil {
        cols = append(cols, RenderClaudeStatus(*opts.ClaudeStatus, opts.AnimFrame))
    }
    return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
```

This mirrors React's component composition:
- Each column is a small, testable function
- Row handles layout/joining
- RowLayout calculates widths once, passed to all rows

---

## Implementation Steps

### Step 1: Create column components (`internal/ui/columns.go`)
- `RenderIndex()`, `RenderLastIcon()`, `RenderExpandIcon()`
- `RenderSessionName()`, `RenderTimeAgo()`
- `RenderGitStatus()`, `RenderClaudeStatus()`
- `RenderSessionRow()` that composes them
- `RenderWindowRow()` for window items
- `RowLayout` struct for width calculations

### Step 2: Create ScrollList helper (`internal/ui/scrolllist.go`)
- Generic `ScrollList[T]` type
- Methods: `SetItems`, `Filter`, `MoveCursor`, `VisibleRange`, `SelectedItem`
- `UpdateScrollOffset()` - keeps cursor visible

### Step 3: Migrate session list
- Replace `items`, `cursor`, `scrollOffset`, `filter` with `ScrollList[Item]`
- Update `handleNormalMode()` to use ScrollList methods
- Update `viewSessionList()` to use new column components

### Step 4: Migrate project picker
- Replace project-related fields with `ScrollList[string]`
- Update `viewPickDirectory()`

### Step 5: Migrate clone picker
- Replace clone-related fields with `ScrollList[string]`
- Update `viewCloneRepo()`

### Step 6: Clean up Model struct
- Remove replaced fields
- Consolidate remaining state

---

## Files to Modify

| File | Changes |
|------|---------|
| `internal/ui/columns.go` | **New** - Column "components" and RenderSessionRow |
| `internal/ui/scrolllist.go` | **New** - Generic ScrollList[T] helper |
| `internal/model/model.go` | Refactor: use ScrollList, call column components |
| `internal/ui/styles.go` | Minor: may move some styles to columns.go |

---

## Expected Outcomes

- **~200-300 lines deleted** from model.go (duplicate scroll logic)
- **Cleaner state** - Model struct shrinks from 93 fields
- **Testable scroll logic** - Unit tests for cursor/scroll behavior
- **Maintainable** - Adding new list modes becomes trivial

---

## Verification

1. Run `make test` - all existing tests pass
2. Run `make build && make install`
3. Test in tmux popup: `tmux display-popup -w50% -h35% -B -E "~/.local/bin/tsm"`
4. Verify:
   - Session navigation (Ctrl+j/k)
   - Session expand/collapse (Ctrl+h/l)
   - Fuzzy filtering (type characters)
   - Session switching (Enter)
   - Project picker (Ctrl+n → pick directory)
   - Clone picker (Ctrl+g)
