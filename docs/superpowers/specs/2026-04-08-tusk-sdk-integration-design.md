# Tusk SDK Integration — PRODUCT.md & ROADMAP.md Update

> Design spec for updating Elephant's product and roadmap documents to reflect Tusk's availability as an embedded Go SDK.

## Context

Tusk (v0.8+) now exposes its complete internals as importable Go packages via a high-level `Client` type. This eliminates the need for MCP-based communication between Elephant and Tusk. Tusk becomes an embedded library rather than a peer system.

Key Tusk capabilities now available as a library:
- Task hierarchy (parent-child nesting to arbitrary depth)
- Workflow enforcement (configurable status transitions, optimistic locking)
- Player management (agent registration, task claiming, atomic `Pop`)
- Relations (blocks, relates_to, duplicates) with cycle detection
- Urgency scoring (multi-factor ranking)
- Rich content (markdown descriptions, UDAs)
- Filters (boolean composition, negation)

## Decisions

1. **Elephant owns the MCP server** — Elephant exposes a network-based MCP server to agent containers, backed by Tusk's SDK. Agents only know about Elephant's MCP.
2. **Thin wrapper over Tusk** — Elephant's MCP tools map closely to Tusk concepts (claim, complete, pop, annotate). Agents are Tusk "players." Plus Elephant-specific tools like `create_pr`.
3. **Two-system + library framing** — Elephant and Brain are the two systems. Tusk is described as an embedded library, not a peer system.
4. **Tusk folds into v0.1** — Embedding Tusk is foundational, not a separate integration phase. The v0.2 "Tusk Integration" initiative is removed.
5. **Token-based agent auth** — Elephant generates a unique token per agent, injected as an env var. Every MCP call includes the token. Elephant validates it before forwarding to Tusk.
6. **Elephant-managed secrets** — Secrets (API keys, credentials) are injected as env vars at container spawn, scoped per task/project. Least-privilege by default.
7. **Standby mode** — After completing work and creating a PR, containers stay alive for review feedback. Torn down only after PR merge or task resolution.

## PRODUCT.md Changes

### Section: Architecture (rewrite)

Replace "Three-System Architecture" with "Two-System Architecture":

#### Two-System Architecture

- **Elephant** — the orchestrator. Manages Docker environments, agent lifecycle, git branching, session persistence, secret management, and the TUI
- **Brain** (WIP) — defines the WBS structure and scoped agent skills/instructions. Supports parallel and sequential execution strategies

#### Tusk (Embedded Task Engine)

Elephant embeds Tusk as a Go library for task management, workflow enforcement, and agent coordination. Rather than running as a separate MCP server, Tusk is imported directly via its SDK (`github.com/germanamz/tusk`).

Tusk provides:
- **Task hierarchy** — parent-child nesting to arbitrary depth, matching the WBS structure
- **Workflow enforcement** — configurable status transitions with optimistic locking
- **Player management** — agents register as Tusk "players", claim tasks to prevent overlapping work
- **Task queue** — atomic `Pop` operation assigns the highest-urgency unclaimed task to an agent
- **Relations** — typed edges (blocks, relates_to, duplicates) with cycle detection
- **Urgency scoring** — multi-factor ranking for task prioritization

#### Elephant MCP Server

Elephant exposes its own MCP server (network-based) to agent containers. This is a thin wrapper over Tusk's SDK plus Elephant-specific operations:

- **Task tools** — create, claim, complete, annotate, pop (delegated to Tusk)
- **Git tools** — `create_pr` and other git operations that are Elephant's domain
- **Progress reporting** — agents report status via Tusk task annotations

**Agent authentication:** Each agent receives a unique token (injected as an env var at container spawn). Every MCP call must include this token. Elephant validates the token maps to the correct player before executing any operation, preventing agents from impersonating each other.

**Secret management:** Elephant manages secrets (API keys, credentials) and injects them as env vars at container spawn time, scoped per task/project. Agents receive only the secrets relevant to their work — least-privilege by default.

### Section: Elephant Orchestrator (rewrite)

#### Elephant (Orchestrator)

- Written in Go for performance and cross-platform distribution
- Embeds Tusk as a library for task management, workflow enforcement, and agent coordination
- Manages agent lifecycle — spawns, monitors, suspends, and tears down agents in Docker containers
- Drives the git branching strategy — creates branches per WBS level, manages merges, triggers parent-level reviews
- Exposes an MCP server to agent containers — thin wrapper over Tusk's SDK plus git operations
- Handles agent authentication (per-agent tokens) and secret injection (least-privilege)
- Calls agents in headless mode (`--print --output-format json`) and tracks conversation state for session resumption
- Provides the TUI for monitoring, interaction, and control

### Section: Docker Isolation Layer (rewrite)

#### Docker Isolation Layer

- Each agent runs in its own container with the relevant codebase mounted
- Agents can operate without permission restrictions — the container _is_ the sandbox
- At spawn time, Elephant injects:
  - An **auth token** unique to the agent, used for all MCP communication
  - **Scoped secrets** (API keys, credentials) required for the agent's task
  - **MCP endpoint** for communicating back to Elephant
- After completing work, agents create a PR and enter **standby mode** — the container stays alive so the agent can address review feedback on its PR/MR
- Containers are torn down only after the PR is merged (or the task is otherwise resolved)

### Section: Tusk subsection (rewrite)

Remove the standalone "Tusk" subsection from the architecture list. Its content is now covered by the "Tusk (Embedded Task Engine)" section above.

### Unchanged sections

- Problem Statement
- Vision
- Core Concepts (WBS, Inverted Flame Graph Branching, Session Continuity)
- Brain (WBS & Skills)
- TUI
- Workflow (Phases 1-3)
- Tooling & Quality
- Phasing (high-level)
- Target User

## ROADMAP.md Changes

### v0.1 — Foundation (rewrite)

#### Initiative: Project Scaffolding
*(unchanged — all complete)*

#### Initiative: Docker Container Management

##### Story: Container Lifecycle
- [x] **Task:** Spawn containers with codebase mounted
- [x] **Task:** Container teardown on task completion
- [x] **Task:** Ephemeral container lifecycle (spin up per task, tear down after)

##### Story: Container Provisioning
- [ ] **Task:** Inject auth token as env var at container spawn
- [ ] **Task:** Inject scoped secrets as env vars at container spawn
- [ ] **Task:** Inject MCP endpoint for agent-to-Elephant communication
- [ ] **Task:** Standby mode — keep container alive after PR creation for review feedback
- [ ] **Task:** Tear down container after PR merge or task resolution

#### Initiative: Tusk Integration

##### Story: Embedded Task Engine
- [ ] **Task:** Add Tusk SDK dependency (`github.com/germanamz/tusk`)
- [ ] **Task:** Initialize Tusk Client with Elephant-managed config (DB path, workflows, projects)
- [ ] **Task:** Configure WBS-aligned task hierarchy (map WBS levels to Tusk parent-child nesting)
- [ ] **Task:** Register agents as Tusk players on container spawn

#### Initiative: Elephant MCP Server

##### Story: MCP Server Core
- [ ] **Task:** Implement network-based MCP server accessible from agent containers
- [ ] **Task:** Token-based authentication middleware (validate agent token on every call)

##### Story: Task Tools
- [ ] **Task:** Expose Tusk task operations as MCP tools (create, claim, complete, annotate, pop)
- [ ] **Task:** Scope tool access — agents can only operate on tasks they are assigned to
- [ ] **Task:** Progress reporting via Tusk task annotations

##### Story: Git Tools
- [ ] **Task:** `create_pr` — agent requests PR creation for its branch
- [ ] **Task:** Additional git operations as needed (TBD based on agent workflow needs)

##### Story: Secret Management
- [ ] **Task:** Secret store configuration (config file defining available secrets per project/task type)
- [ ] **Task:** Secret injection at container spawn (env vars, least-privilege)

##### Story: Safety
- [ ] **Task:** Timeout safety net for unresponsive agents

#### Initiative: Agent Execution
*(unchanged)*

#### Initiative: Basic TUI
*(unchanged)*

### v0.2 — Orchestration (rewrite)

Remove the "Tusk Integration" initiative entirely. Keep everything else:

#### Initiative: Git Branching Strategy

##### Story: Inverted Flame Graph Branching
- [ ] **Task:** Auto-create branches per WBS level
- [ ] **Task:** Manage merges from child branches into parent branches
- [ ] **Task:** Trigger parent-level reviews on child completion

#### Initiative: Brain Integration

##### Story: WBS & Skills Loading
- [ ] **Task:** Load WBS definitions from Brain
- [ ] **Task:** Apply scoped agent skills/instructions per WBS level
- [ ] **Task:** Support parallel and sequential execution strategies

#### Initiative: Cascading Review Workflow

##### Story: Parent Session Review
- [ ] **Task:** Resume original parent agent session after children merge
- [ ] **Task:** Parent agent reviews merged result with full decomposition context

##### Story: Review Outcomes
- [ ] **Task:** On review pass: mark done, merge upward
- [ ] **Task:** On review fail: spawn corrective child tasks
- [ ] **Task:** Cascade reviews up the WBS tree to top-level vision

### v0.3, v0.4
*(unchanged)*
