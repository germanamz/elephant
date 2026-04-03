# Elephant — Implementation Roadmap

## v0.1 — Foundation

### Initiative: Project Scaffolding

#### Story: Go Module & CI Setup
- [x] **Task:** Initialize Go module
- [x] **Task:** Set up CI pipeline

#### Story: Code Quality Tooling
- [x] **Task:** Configure golangci-lint for static analysis
- [x] **Task:** Configure Lefthook for git hooks (pre-commit, pre-push)
- [x] **Task:** Configure Conform for conventional commit enforcement

#### Story: Build & Release Pipeline
- [x] **Task:** Set up GoReleaser for cross-platform binary builds (macOS, Linux, Windows)

### Initiative: Docker Container Management

#### Story: Container Lifecycle
- [ ] **Task:** Spawn containers with codebase mounted
- [ ] **Task:** Container teardown on task completion
- [ ] **Task:** Ephemeral container lifecycle (spin up per task, tear down after)

### Initiative: Agent Execution

#### Story: Headless Agent Runner
- [ ] **Task:** Run Claude Code in headless mode (`--print --output-format json`)
- [ ] **Task:** Track and parse agent output

#### Story: Session Continuity
- [ ] **Task:** Session persistence (save conversation state)
- [ ] **Task:** Session resumption (restore agent with prior context)

### Initiative: Basic TUI

#### Story: Agent Monitoring
- [ ] **Task:** Agent status monitoring view
- [ ] **Task:** Log output streaming

---

## v0.2 — Orchestration

### Initiative: Git Branching Strategy

#### Story: Inverted Flame Graph Branching
- [ ] **Task:** Auto-create branches per WBS level
- [ ] **Task:** Manage merges from child branches into parent branches
- [ ] **Task:** Trigger parent-level reviews on child completion

### Initiative: Brain Integration

#### Story: WBS & Skills Loading
- [ ] **Task:** Load WBS definitions from Brain
- [ ] **Task:** Apply scoped agent skills/instructions per WBS level
- [ ] **Task:** Support parallel and sequential execution strategies

### Initiative: Tusk Integration

#### Story: MCP Task Management
- [ ] **Task:** Connect to Tusk MCP server from agent containers
- [ ] **Task:** Agents pick up tasks, report progress, and mark completion via MCP

### Initiative: Cascading Review Workflow

#### Story: Parent Session Review
- [ ] **Task:** Resume original parent agent session after children merge
- [ ] **Task:** Parent agent reviews merged result with full decomposition context

#### Story: Review Outcomes
- [ ] **Task:** On review pass: mark done, merge upward
- [ ] **Task:** On review fail: spawn corrective child tasks
- [ ] **Task:** Cascade reviews up the WBS tree to top-level vision

---

## v0.3 — Collaborative TUI

### Initiative: WBS Refinement

#### Story: Collaborative Decomposition
- [ ] **Task:** Back-and-forth dialogue for defining product vision and WBS
- [ ] **Task:** Each WBS level broken down by its own agent session

#### Story: WBS Visualization
- [ ] **Task:** Work breakdown tree visualization

### Initiative: Monitoring

#### Story: Progress & Logs
- [ ] **Task:** Task progress views
- [ ] **Task:** Log streaming per agent

### Initiative: Interactive Agent Control

#### Story: Agent Intervention
- [ ] **Task:** Pause, resume, and intervene on running agents
- [ ] **Task:** Approve or reject reviews from the TUI

---

## v0.4 — Scale

### Initiative: Multi-User & Distribution

#### Story: Multi-User Support
- [ ] **Task:** Multi-user support

#### Story: Remote Execution
- [ ] **Task:** Remote/distributed agent execution

### Initiative: Platform Access

#### Story: API Layer
- [ ] **Task:** API layer for programmatic access

#### Story: Team Workflows
- [ ] **Task:** Team workflows and shared Elephant instances
