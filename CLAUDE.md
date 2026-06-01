# Claude Code Configuration

## Project Overview

Pod-reaper is a Go-based Kubernetes controller that automatically reaps pods based on configurable rules. The architecture follows an interface-driven rule system where each rule implements `load()` and `ShouldReap()` methods.

**Key Stats:**
- Test Coverage: 72.6% overall (97% rules, 62% reaper)
- 6 rule implementations
- SOLID principles: Excellent adherence
- 191 total commits, primary maintainer-driven with community contributions

## Code Conventions

### Commit Messages
This project uses a hybrid convention system:

**For External/PR Contributions (Conventional Commits):**
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `feat(scope):` - Scoped feature (e.g., `feat(helm chart):`)

**For Internal Development (Custom Single-Letter):**
- `F` / `f` - Feature (uppercase for major, lowercase for minor)
- `R` / `r` - Refactoring (uppercase for significant)
- `D` / `d` - Documentation
- `t` - Test additions

### Go Standards
- Run `go fmt ./reaper ./rules` before commits
- Run `golint` for linting
- Every rule must have corresponding `_test.go` file
- Use `logrus` for structured logging

### Rule Implementation Pattern
When adding new rules:
1. Create `rules/rulename.go` with struct implementing `Rule` interface
2. Create `rules/rulename_test.go` with comprehensive tests
3. Register rule in `LoadRules()` function in `rules/rules.go`
4. Environment variables for configuration follow `UPPER_SNAKE_CASE`

---

## Specialized Planning Agents

Use these agents before implementing significant changes to ensure robust, high-quality code.

### 1. Edge Case Extraction Agent

**Purpose:** Exhaustively identifies edge cases, boundary conditions, and failure modes before implementation.

**Invocation:**
```
Run the edge-case-extractor agent for [feature/change description]
```

**Agent Behavior:**
- Analyzes existing test files to understand current edge case coverage
- Examines error handling patterns in similar code paths
- Identifies boundary conditions (nil values, empty collections, max/min values)
- Reviews Kubernetes API edge cases (pod phases, container states, timing issues)
- Documents race conditions and concurrency concerns
- Lists potential panic scenarios and recovery strategies
- Cross-references with Go's common pitfalls (nil interfaces, slice mutations)

---

### 2. Git History Convention Analyzer

**Purpose:** Extracts patterns, conventions, and best practices from git history to ensure consistency.

**Invocation:**
```
Run the git-convention-analyzer agent for [area of codebase]
```

---

### 3. Code Deep Dive Agent

**Purpose:** Performs exhaustive analysis of code paths, dependencies, and side effects.

**Invocation:**
```
Run the code-deep-diver agent for [file/package/function]
```

---

### 4. Test Quality Analyzer

**Purpose:** Evaluates test coverage, quality, and identifies gaps before implementation.

**Invocation:**
```
Run the test-quality-analyzer agent for [package/feature area]
```

---

### 5. Architecture Review Agent

**Purpose:** Ensures changes align with existing architecture and best practices.

**Invocation:**
```
Run the architecture-reviewer agent for [proposed change description]
```

---

### 6. Pre-Implementation Checklist Agent

**Purpose:** Generates a comprehensive checklist before any implementation begins.

**Invocation:**
```
Run the pre-implementation-checklist agent for [feature/change]
```

---

## Quick Commands

| Command | Purpose |
|---------|---------|
| `analyze edge cases for X` | Run edge case extractor |
| `analyze git history for X` | Run git convention analyzer |
| `deep dive into X` | Run code deep diver |
| `analyze test quality for X` | Run test quality analyzer |
| `review architecture for X` | Run architecture reviewer |
| `generate checklist for X` | Run pre-implementation checklist |
| `full analysis for X` | Run all agents in sequence |

---

# CODEBASE ANALYSIS RESULTS

*Generated: January 2025*

---

## 1. Architecture Analysis

### SOLID Principles Assessment

| Principle | Rating | Evidence |
|-----------|--------|----------|
| **Single Responsibility** | EXCELLENT | Clear separation: `reaper.go` (K8s), `options.go` (config), `rules/*.go` (decisions) |
| **Open/Closed** | EXCELLENT | New rules added without modifying existing code |
| **Liskov Substitution** | EXCELLENT | All 6 rules properly implement Rule interface |
| **Interface Segregation** | EXCELLENT | Rule interface has only 2 methods (minimal) |
| **Dependency Inversion** | GOOD | Rules use abstraction; reaper depends on K8s client directly (acceptable) |

### Package Structure

```
reaper/
├── main.go          # Process entry, logging setup (~50 lines)
├── options.go       # Configuration parsing (100+ lines)
├── reaper.go        # Controller logic: scheduling, filtering, deleting
└── *_test.go        # Tests

rules/
├── rules.go         # Rule interface, loading, composition
├── chaos.go         # Random reaping
├── container_status.go  # Container state matching
├── duration.go      # Age-based reaping
├── pod_status.go    # Pod reason matching
├── pod_status_phase.go  # Pod phase matching
├── unready.go       # Unready duration
└── *_test.go        # Tests for each rule
```

### Core Interface

```go
type Rule interface {
    load() (bool, string, error)      // Configure from env vars
    ShouldReap(pod v1.Pod) (bool, string)  // Decision logic
}
```

### Execution Flow

```
main() → newReaper() → loadOptions() → rules.LoadRules()
                    → harvest() → cron schedule
                              → scytheCycle() [on schedule]
                                  → getPods() → filter → sort
                                  → rules.ShouldReap(pod) [AND logic]
                                  → reapPod()
```

---

## 2. Critical Issues Identified

### HIGH RISK

| Issue | Location | Impact | Recommendation |
|-------|----------|--------|----------------|
| **Race Condition in Chaos Rule** | `rules/chaos.go:21-23` | Global `rand` not thread-safe | Use `sync/rand` or mutex |
| **Error Check After Use** | `reaper.go:62-68` | podList used before error check | Move error check before sorting |
| **Potential Nil Panic** | `rules/unready.go:38` | LastTransitionTime could be nil | Add defensive nil check |
| **Redundant Panic Calls** | `reaper.go:24-25, 29-30` | `logrus.Panic()` then `panic()` | Remove redundant `panic()` |

### MEDIUM RISK

| Issue | Location | Impact |
|-------|----------|--------|
| Chaos values outside [0,1] not validated | `rules/chaos.go` | Unexpected behavior |
| Negative duration accepted | `rules/duration.go` | Pods never reaped |
| Clock skew not handled | Duration rules | Off-by-seconds edge cases |
| Case sensitivity in phase matching | `rules/pod_status_phase.go` | "failed" won't match "Failed" |
| Whitespace in comma-separated values | All rules | Values not trimmed |

---

## 3. Test Coverage Analysis

### Coverage by Package

| Package | Coverage | Status |
|---------|----------|--------|
| `rules` | 97.0% | Excellent |
| `reaper` | 62.0% | Good |
| **Overall** | **72.6%** | Good |

### Critical Untested Paths (0% Coverage)

| Function | File | Why Untested |
|----------|------|--------------|
| `newReaper()` | reaper/reaper.go | Requires K8s cluster/mocking |
| `getPods()` | reaper/reaper.go | Requires K8s API mock |
| `reapPod()` | reaper/reaper.go | Complex mocking needed |
| `scytheCycle()` | reaper/reaper.go | Integration testing |
| `harvest()` | reaper/reaper.go | Timing-dependent |

### Missing Test Scenarios

**Rules Package:**
- Chaos rule: Values outside [0, 1], NaN/Inf, concurrent access
- Container status: Init containers alone, empty string status
- Duration: Negative duration, nil StartTime panic test
- Unready: Status values other than "True"/"False"
- All rules: Whitespace handling, special characters

**Reaper Package:**
- maxPods limit enforcement
- Eviction vs deletion paths
- API failure handling
- Concurrent cycle execution

---

## 4. Edge Cases Reference

### Per-Rule Edge Cases

#### Chaos Rule
| Input | Behavior | Tested? |
|-------|----------|---------|
| `CHAOS_CHANCE=""` | Not loaded | Yes |
| `CHAOS_CHANCE="0"` | Never reaps | Yes |
| `CHAOS_CHANCE="1"` | Always reaps | Yes |
| `CHAOS_CHANCE="-0.5"` | Parses, never reaps | **NO** |
| `CHAOS_CHANCE="2.0"` | Parses, always reaps | **NO** |
| `CHAOS_CHANCE="NaN"` | Parses as NaN | **NO** |

#### Duration Rule
| Scenario | Behavior | Tested? |
|----------|----------|---------|
| Nil StartTime | Returns false | Yes |
| Negative duration | Future cutoff, never reaps | **NO** |
| Zero duration | All pods reaped | **NO** |
| Clock skew | May cause off-by-one | **NO** |

#### Container Status Rule
| Scenario | Behavior | Tested? |
|----------|----------|---------|
| No containers | Returns false | Implicit |
| Init container only | Checks init containers | **NO** |
| Running state | Not matched | **NO** |
| Empty status string | May match empty reason | **NO** |

#### Unready Rule
| Scenario | Behavior | Tested? |
|----------|----------|---------|
| No Ready condition | Returns false | Yes |
| Ready=True | Returns false | Implicit |
| Nil LastTransitionTime | **POTENTIAL PANIC** | **NO** |

### Kubernetes API Edge Cases

| Scenario | Current Handling |
|----------|------------------|
| API rate limiting (429) | Panics |
| Network timeout | Blocks indefinitely |
| Pod deleted between list and delete | Error logged, continues |
| PDB violation (eviction) | Error logged, continues |
| Stale cache | Not handled |

---

## 5. Git Convention Summary

### Commit Message Format

**External PRs:** `type(scope): message`
```
feat: reaping by POD_STATUS_PHASES
fix(readme): typo fix
docs: update changelog
```

**Internal Development:** `[Letter] - description`
```
F - have reaper use sorting strategies
r - extract variables
t - added test for method
```

### Branch Naming
```
feat/pod-status-phase
feature/lookup-initcontainer
doc-readme
```

### Files Likely to Conflict
- `rules/rules.go` - Rule registry
- `CHANGELOG.md` - Version tracking
- `go.mod` / `go.sum` - Dependencies

### PR Merge Strategy
- GitHub merge commits (preserves branch history)
- No squash or rebase

---

## 6. Dependency Graph

### External Packages

| Package | Purpose |
|---------|---------|
| `k8s.io/api/core/v1` | Pod types |
| `k8s.io/apimachinery` | Metadata, labels |
| `k8s.io/client-go` | K8s client |
| `github.com/robfig/cron/v3` | Scheduling |
| `github.com/sirupsen/logrus` | Logging |
| `github.com/stretchr/testify` | Testing |

### Internal Dependencies

```
main.go
  └── rules.LoadRules()
  └── reaper.harvest()

reaper.go
  └── options (struct)
  └── rules.Rules (interface)

options.go
  └── rules.LoadRules()

rules/rules.go
  └── All rule implementations
```

---

## 7. Environment Variables Reference

### Core Configuration

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `NAMESPACE` | string | "" (all) | Target namespace |
| `SCHEDULE` | cron | "@every 1m" | Reap cycle schedule |
| `RUN_DURATION` | duration | "0s" (forever) | Process lifetime |
| `GRACE_PERIOD` | duration | nil (K8s default) | Pod termination grace |
| `DRY_RUN` | bool | false | Skip actual deletion |
| `MAX_PODS` | int | 0 (unlimited) | Max deletions per cycle |
| `EVICT` | bool | false | Use eviction API |

### Rule Configuration

| Variable | Rule | Format |
|----------|------|--------|
| `CHAOS_CHANCE` | chaos | Float 0.0-1.0 |
| `MAX_DURATION` | duration | Go duration (e.g., "24h") |
| `MAX_UNREADY` | unready | Go duration |
| `CONTAINER_STATUSES` | container_status | Comma-separated |
| `POD_STATUSES` | pod_status | Comma-separated |
| `POD_STATUS_PHASES` | pod_status_phase | Comma-separated |

### Filtering Configuration

| Variable | Purpose |
|----------|---------|
| `EXCLUDE_LABEL_KEY` | Label key to exclude |
| `EXCLUDE_LABEL_VALUES` | Comma-separated values |
| `REQUIRE_LABEL_KEY` | Required label key |
| `REQUIRE_LABEL_VALUES` | Comma-separated values |
| `REQUIRE_ANNOTATION_KEY` | Required annotation key |
| `REQUIRE_ANNOTATION_VALUES` | Comma-separated values |
| `POD_SORTING_STRATEGY` | random/oldest-first/youngest-first/pod-deletion-cost |

---

## 8. Adding a New Rule - Complete Guide

### Step 1: Create Rule File

```go
// rules/memory.go
package rules

import (
    "os"
    "strconv"
    v1 "k8s.io/api/core/v1"
)

const envMemoryLimit = "MEMORY_LIMIT_BYTES"

type memoryLimit struct {
    limitBytes int64
}

func (m *memoryLimit) load() (bool, string, error) {
    value, exists := os.LookupEnv(envMemoryLimit)
    if !exists {
        return false, "", nil
    }

    limit, err := strconv.ParseInt(value, 10, 64)
    if err != nil {
        return false, "", err
    }

    m.limitBytes = limit
    return true, "memory limit " + value, nil
}

func (m *memoryLimit) ShouldReap(pod v1.Pod) (bool, string) {
    // Implementation here
    return false, ""
}
```

### Step 2: Create Tests

```go
// rules/memory_test.go
package rules

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMemoryLimitLoad(t *testing.T) {
    os.Clearenv()

    t.Run("no load", func(t *testing.T) {
        m := memoryLimit{}
        loaded, _, err := m.load()
        assert.NoError(t, err)
        assert.False(t, loaded)
    })

    t.Run("load", func(t *testing.T) {
        os.Setenv(envMemoryLimit, "1073741824")
        m := memoryLimit{}
        loaded, message, err := m.load()
        assert.NoError(t, err)
        assert.True(t, loaded)
        assert.Contains(t, message, "1073741824")
    })
}
```

### Step 3: Register in LoadRules

```go
// rules/rules.go
func LoadRules() (Rules, error) {
    rules := []Rule{
        &chaos{},
        &containerStatus{},
        &duration{},
        &unready{},
        &podStatus{},
        &podStatusPhase{},
        &memoryLimit{},  // Add here
    }
    // ...
}
```

### Step 4: Update Documentation
- README.md: Add env var documentation
- CHANGELOG.md: Add entry for new feature

### Step 5: Run Checks
```bash
go fmt ./rules
golint ./rules
go test ./rules -cover
```

---

## 9. Pre-Implementation Checklist Template

```markdown
## Pre-Implementation Checklist: [Feature Name]

### Prerequisites
- [ ] Read existing similar rules for patterns
- [ ] Understand Rule interface contract
- [ ] Identify all edge cases (use edge case analysis above)

### Files to Create
- [ ] rules/[name].go - Rule implementation
- [ ] rules/[name]_test.go - Test file

### Files to Modify
- [ ] rules/rules.go - Register in LoadRules()
- [ ] README.md - Document new env var
- [ ] CHANGELOG.md - Add feature entry

### Edge Cases to Handle
- [ ] Empty/missing env var
- [ ] Invalid env var value
- [ ] Nil pod fields
- [ ] Empty collections

### Tests Required
- [ ] Load with no env var (not loaded)
- [ ] Load with valid env var
- [ ] Load with invalid env var (error)
- [ ] ShouldReap positive case
- [ ] ShouldReap negative case
- [ ] Edge cases from analysis

### Review Checklist
- [ ] Follows commit conventions
- [ ] Passes `go fmt ./rules`
- [ ] Passes `golint ./rules`
- [ ] All tests pass: `go test ./rules`
- [ ] Coverage maintained: `go test ./rules -cover`
- [ ] No redundant panic() calls
- [ ] Error messages include context
```

---

## 10. Known Improvement Opportunities

### Priority 1: Bug Fixes
1. Fix error check order in `reaper.go:62-68`
2. Remove redundant `panic()` calls after `logrus.Panic()`
3. Add nil check for LastTransitionTime in unready rule

### Priority 2: Test Coverage
1. Add Kubernetes client mocking for reaper tests
2. Add init container test coverage
3. Fix time-based test flakiness (use fixed times)
4. Add benchmark tests

### Priority 3: Enhancements
1. Validate chaos chance in [0, 1] range
2. Trim whitespace from comma-separated values
3. Add graceful shutdown handling
4. Document pod sorting nil StartTime behavior

---

## Usage Workflow

For any significant change, run agents in this order:

1. **Architecture Review** - Validate the approach fits the codebase
2. **Git Convention Analyzer** - Understand how to structure the work
3. **Code Deep Diver** - Understand all affected code paths
4. **Edge Case Extractor** - Identify all edge cases upfront
5. **Test Quality Analyzer** - Plan test coverage
6. **Pre-Implementation Checklist** - Generate actionable plan

### Example

```
User: I want to add a new rule that reaps pods based on memory usage

Claude: Let me run the planning agents to ensure a robust implementation.

[Runs architecture-reviewer for "memory-based pod reaping rule"]
[Runs git-convention-analyzer for "rules package"]
[Runs code-deep-diver for "rules/rules.go and existing rules"]
[Runs edge-case-extractor for "memory threshold rule"]
[Runs test-quality-analyzer for "rules package"]
[Runs pre-implementation-checklist for "memory usage rule"]

Based on the analysis, here's the comprehensive implementation plan...
```
