# Plan: Multi-Provider Git URL Support

## Context

`helm repos add` and the TUI clone flow use `ParseGitURL` / `ResolveOwnerRepo` in `internal/giturl/github.go`. These functions are GitHub-centric:

- `ParseGitURL` extracts `owner/repo` by stripping the host — works for GitHub, breaks for anything else
- SSH protocol regex `git@([^:]+):(.+?)(?:\.git)?$` is unanchored — matches at position 6 inside `ssh://` URLs, treating the port `7999` as a directory name
- `FetchAvailableRepos` and `CheckGhCli` are GitHub-specific (uses `gh` CLI) — these should stay GitHub-only
- Clone destination path is always `{owner/repo}` — needs to be host-qualified for non-GitHub providers to avoid collisions
- `ListClonedRepos` scans at hardcoded depth 2 — repos at depth 3 won't be discovered
- `scanProjectDirectories` uses fixed-depth walk — same limitation, plus increasing depth scans INTO repo internals

**Example of the bug:**

```
$ helm repos add ssh://git@git.corp.example.com:7999/~alice/my-project.git
→ Cloning 7999/~alice/my-project...
✓ 7999/~alice/my-project cloned to ~/repos/7999/~alice/my-project
```

The SSH regex matches `git@` at position 6 (inside the `ssh://` protocol URL), treats `git.corp.example.com` as the host, `:` as the SCP separator, and captures `7999/~alice/my-project` — the port becomes a directory name.

## Critical Files

| File                             | Change                                                                      |
| -------------------------------- | --------------------------------------------------------------------------- |
| `internal/giturl/url.go`         | **NEW** — `GitURL` struct + `ParseGitURL` with anchored regexes             |
| `internal/giturl/url_test.go`    | **NEW** — tests for all URL formats                                         |
| `internal/giturl/github.go`      | Remove old `ParseGitURL`, update `ResolveOwnerRepo` to use new struct       |
| `internal/config/user_config.go` | Add `GitProviders` map, update `ListClonedRepos` to use recursive walk      |
| `internal/model/directory.go`    | Update `scanProjectDirectories` to use recursive walk with `.git` detection |
| `cmd/helm/repos.go`              | Update `runReposAdd` and `runReposRebuild` to use new resolution            |
| `internal/model/clone.go`        | Update `cloneSelectedRepo` to pass providers config                         |

## Phase 1: GitURL struct + URL parsing (TDD)

### New file: `internal/giturl/url.go`

```go
// GitURL represents a parsed git URL with structured fields.
type GitURL struct {
    Scheme string // "ssh", "https", "git" (SCP-like: git@host:path)
    Host   string // hostname (no port)
    Port   int    // 0 if default
    Path   string // repo path relative to host (without .git)
}

// ParseGitURL parses any supported git URL into a GitURL struct.
// All regexes are ^-anchored to prevent false matches.
func ParseGitURL(url string) (GitURL, error) {
    // SCP-like: git@host:path[.git]
    // SSH:      ssh://git@host[:port]/path[.git]
    // HTTPS:    https://host/path[.git]
}
```

### Tests (TDD — write tests first)

```
TestParseGitURL:
  ✓ SCP: git@github.com:owner/repo.git → host=github.com, path=owner/repo
  ✓ SCP no .git: git@github.com:owner/repo → same
  ✓ SSH: ssh://git@codeberg.org/owner/repo.git → host=codeberg.org, path=owner/repo
  ✓ SSH no .git: ssh://git@codeberg.org/owner/repo → same
  ✓ SSH+port: ssh://git@git.corp.example.com:7999/~alice/my-project.git → host=git.corp.example.com, port=7999, path=~alice/my-project
  ✓ HTTPS: https://github.com/owner/repo.git → host=github.com, path=owner/repo
  ✓ HTTPS no .git: https://github.com/owner/repo → same
  ✓ rejects: not-a-url, empty, git@:
```

### Commits

1. `refactor(giturl): introduce GitURL struct with anchored URL parsing` — `internal/giturl/url.go`, `internal/giturl/url_test.go`

## Phase 2: Config — `git_providers` map

### Config structure

Add to `Config`:

```yaml
# Optional — maps git hosts to directory aliases for clone destination paths.
# Empty string = use path as-is (GitHub style). Omitted hosts use host/ as prefix.
git_providers:
  github.com: "" # owner/repo (no host prefix)
  codeberg.org: "" # owner/repo
  git.corp.example.com: corp # → corp/~alice/my-project (personal) or corp/team/repo (org)
```

```go
// In Config struct:
GitProviders map[string]string `yaml:"git_providers,omitempty"`
```

**Important:** The `~` prefix in paths like `~alice/my-project` is semantically meaningful — it distinguishes **personal repos** (`~user/repo`) from **team/org repos** (`team/repo`). The `~` must be preserved in the directory structure, not stripped.

| URL pattern                                                 | Path                | Resolved directory       |
| ----------------------------------------------------------- | ------------------- | ------------------------ |
| `ssh://git@git.corp.example.com:7999/~alice/my-project.git` | `~alice/my-project` | `corp/~alice/my-project` |
| `ssh://git@git.corp.example.com:7999/websdk/web-ui.git`     | `websdk/web-ui`     | `corp/websdk/web-ui`     |

### Commits

1. `feat(config): add git_providers config for host-to-alias mapping` — `internal/config/user_config.go`

## Phase 3: Resolution logic

### New function: `ResolveRepoDir`

```go
// ResolveRepoDir returns the directory name for a parsed git URL.
// The providers map is from config.GitProviders (host → alias).
//
// Rules:
//   1. Look up GitURL.Host in providers map
//   2. If found with alias: "alias/{cleaned_path}"
//   3. If found with empty string: "{cleaned_path}"
//   4. If not found: "{host}/{cleaned_path}"
//
// cleaned_path = strip leading /, keep ~ prefix (semantically meaningful)
func ResolveRepoDir(gitURL GitURL, providers map[string]string) string {
    cleaned := cleanPath(gitURL.Path)
    if providers == nil {
        return gitURL.Host + "/" + cleaned
    }
    alias, ok := providers[gitURL.Host]
    if !ok {
        return gitURL.Host + "/" + cleaned
    }
    if alias == "" {
        return cleaned
    }
    return alias + "/" + cleaned
}
```

`cleanPath` strips leading `/` (URL syntax) but preserves `~` prefix — `~` distinguishes personal repos from team/org repos.

### Update `ResolveOwnerRepo`

Keep backward compat by delegating to `ParseGitURL` + `ResolveRepoDir`:

```go
func ResolveOwnerRepo(input string, providers map[string]string) (string, error) {
    if strings.Contains(input, "@") || strings.Contains(input, "://") {
        parsed, err := ParseGitURL(input)
        if err != nil {
            return "", err
        }
        return ResolveRepoDir(parsed, providers), nil
    }
    // Plain owner/repo format — return as-is
    if !strings.Contains(input, "/") {
        return "", fmt.Errorf("invalid repo format: %s", input)
    }
    return input, nil
}
```

### Tests

```
TestResolveRepoDir:
  ✓ GitHub: host=github.com, path=owner/repo, config=["github.com": ""] → owner/repo
  ✓ Codeberg: host=codeberg.org, path=ziglings/exercises, config=["codeberg.org": ""] → ziglings/exercises
  ✓ Personal repo (tilde preserved): host=git.corp.example.com, path=~alice/my-project, config=["git.corp.example.com": "corp"] → corp/~alice/my-project
  ✓ Team repo (no tilde): host=git.corp.example.com, path=websdk/web-ui, config=["git.corp.example.com": "corp"] → corp/websdk/web-ui
  ✓ Unknown host (not in config): host=gitlab.com, path=mygroup/project → gitlab.com/mygroup/project
  ✓ Path with leading slash: path=/owner/repo → owner/repo
  ✓ nil providers: → host/path
```

### Commits

1. `feat(giturl): add ResolveRepoDir for provider-aware directory naming` — `internal/giturl/url.go`, `internal/giturl/url_test.go`
2. `refactor(giturl): update ResolveOwnerRepo to accept providers config` — `internal/giturl/github.go`, `internal/giturl/github_test.go`

## Phase 4: Recursive directory scanning

### Problem

Both `ListClonedRepos` and `scanProjectDirectories` use fixed-depth scanning:

- `ListClonedRepos` scans at depth 2 — repos at depth 3 (like `imfusion/brunner/agents`) aren't found
- `scanProjectDirectories` uses `walkAtDepth` with `ProjectDepth` (default 2) — same issue
- Increasing depth would scan INTO repo internals (`.agents`, `.claude`, `node_modules`, etc.)

### Solution

Replace both with a recursive walk that **stops at `.git` directories**:

```go
// scanForGitRepos walks baseDir recursively, collecting directories that contain .git.
// Stops recursing into directories that are themselves repos.
func scanForGitRepos(baseDir string) []string {
    var repos []string
    var walk func(dir string, relPath string)
    walk = func(dir string, relPath string) {
        entries, err := os.ReadDir(dir)
        if err != nil {
            return
        }
        for _, entry := range entries {
            if !entry.IsDir() {
                continue
            }
            if isInternalHiddenDir(entry.Name()) {
                continue
            }
            entryRel := entry.Name()
            if relPath != "" {
                entryRel = relPath + "/" + entry.Name()
            }
            entryPath := filepath.Join(dir, entry.Name())
            gitPath := filepath.Join(entryPath, ".git")
            if _, err := os.Stat(gitPath); err == nil {
                // Found a repo — add it, don't recurse into it
                repos = append(repos, entryRel)
                continue
            }
            // Not a repo — recurse deeper
            walk(entryPath, entryRel)
        }
    }
    walk(baseDir, "")
    sort.Strings(repos)
    return repos
}
```

### `ListClonedRepos` — update to use recursive walk

```go
func ListClonedRepos(basePath string) ([]string, error) {
    if _, err := os.Stat(basePath); os.IsNotExist(err) {
        return []string{}, nil
    }
    return scanForGitRepos(basePath), nil
}
```

### `scanProjectDirectories` — update to use recursive walk

```go
func (m *Model) scanProjectDirectories() []string {
    var dirs []string
    for _, baseDir := range m.config.ProjectDirs {
        repos := scanForGitRepos(baseDir)
        for _, repo := range repos {
            dirs = append(dirs, filepath.Join(baseDir, repo))
        }
    }
    return dirs
}
```

Remove `walkAtDepth` and the `ProjectDepth` config field (no longer needed — depth is determined dynamically by `.git` detection).

### How it handles mixed depths

```
~/repos/imfusion/
  brunner/          ← no .git → recurse
    agents/.git     ← found → add "brunner/agents", stop
  web-sdk/.git      ← found → add "web-sdk", stop (no scan into .teamcity, .vscode)
  web-ui/.git       ← found → add "web-ui", stop (no scan into .agents, .claude)
  web-viewer/.git   ← found → add "web-viewer", stop
  web-viewer-next/.git ← found → add "web-viewer-next", stop

~/repos/black-atom-industries/
  helm/.git         ← found → add "helm", stop
  cockpit/.git      ← found → add "cockpit", stop
```

Result: `["brunner/agents", "web-sdk", "web-ui", "web-viewer", "web-viewer-next"]`

### Commits

1. `refactor(config): replace fixed-depth scan with recursive .git detection` — `internal/config/user_config.go`, `internal/config/user_config_test.go`
2. `refactor(model): update scanProjectDirectories to use recursive walk` — `internal/model/directory.go`

## Phase 5: Wire up CLI + TUI

### `cmd/helm/repos.go` — `runReposAdd`

```go
func runReposAdd(args []string) error {
    cfg, err := config.Load()
    // ...
    ownerRepo, err := giturl.ResolveOwnerRepo(args[0], cfg.GitProviders)
    // ... rest unchanged
}
```

### `cmd/helm/repos.go` — `runReposRebuild`

```go
// postCloneMap keyed by the resolved directory name
if parsed, err := giturl.ParseGitURL(entry.URL); err == nil {
    postCloneMap[giturl.ResolveRepoDir(parsed, cfg.GitProviders)] = entry.PostClone
}
```

### `internal/model/clone.go` — `cloneSelectedRepo`

The TUI clone flow has two entry points:

1. "My repos" (gh CLI) — stays GitHub-only, passes `owner/repo` as selected
2. "Enter URL" — calls `ResolveOwnerRepo` with providers config

For (2), update `handleCloneURLMode` to pass `m.config.GitProviders`:

```go
ownerRepo, err := giturl.ResolveOwnerRepo(value, m.config.GitProviders)
```

### Commits

1. `refactor(repos): use ResolveOwnerRepo with providers for clone destination` — `cmd/helm/repos.go`
2. `refactor(clone): pass GitProviders through TUI for URL resolution` — `internal/model/clone.go`

## Phase 6: Cleanup

### Remove old `ParseGitURL` (string version)

The old `ParseGitURL(string) string` is replaced by `ParseGitURL(string) (GitURL, error)`. Remove the old function and update all callers.

### Commits

1. `refactor(giturl): remove old string-based ParseGitURL` — `internal/giturl/github.go`, `internal/giturl/github_test.go`

## Usage Examples

### Basic GitHub (no config change needed)

```bash
helm repos add black-atom-industries/helm
# → Cloning black-atom-industries/helm...
# ✓ black-atom-industries/helm cloned to ~/repos/black-atom-industries/helm
```

### Self-hosted (with config)

```yaml
# ~/.config/black-atom/helm-tmux/config.yml
git_providers:
  git.corp.example.com: corp
```

Personal repo:

```bash
$ helm repos add ssh://git@git.corp.example.com:7999/~alice/my-project.git
→ Cloning corp/~alice/my-project...
✓ corp/~alice/my-project cloned to ~/repos/corp/~alice/my-project
```

Team repo:

```bash
$ helm repos add ssh://git@git.corp.example.com:7999/websdk/web-ui.git
→ Cloning corp/websdk/web-ui...
✓ corp/websdk/web-ui cloned to ~/repos/corp/websdk/web-ui
```

### Unknown host (no config needed, uses host prefix)

```bash
$ helm repos add git@gitlab.com:mygroup/myproject.git
→ Cloning gitlab.com/mygroup/myproject...
✓ gitlab.com/mygroup/myproject cloned to ~/repos/gitlab.com/mygroup/myproject
```

### TUI clone flow

```
[Clone] Enter URL: ssh://git@git.corp.example.com:7999/~alice/my-project.git

Cloning corp/~alice/my-project...
Session: corp-~alice-my-project
Switch to the new session? [y/n]
```

### Project picker

The project picker now dynamically discovers repos at any depth by detecting `.git` directories:

```
[Projects] filter:
  ▸ brunner/agents
    helm
    web-sdk
    web-ui
    web-viewer
    web-viewer-next
```

Repos at depth 3 (`brunner/agents`) appear alongside depth-2 repos. Subdirectories inside repos (`.agents`, `.claude`, `node_modules`) are never shown.
