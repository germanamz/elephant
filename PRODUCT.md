# Elephant — Product Design Document

> Internal alignment document for contributors. Defines what Elephant is, why it exists, and how it works.

## Problem Statement

AI coding agents like Claude Code are powerful but require constant user supervision. They ask for permission before risky operations, can't run autonomously for extended periods, and when multiple agents need to work in parallel, there's no way to coordinate them safely. The user becomes the bottleneck.

## Vision

Elephant is an AI orchestration platform that lets you define _what_ you want built at a high level, then steps back and lets agents do the work — safely, in parallel, with structured review points.

It achieves this by:

- Running each agent in an isolated Docker container so they can operate without safety guardrails slowing them down
- Decomposing work from vision down to subtasks through collaborative refinement
- Managing a git branching strategy that mirrors the work breakdown, with review checkpoints at every level
- Presenting everything through a TUI that lets you monitor, interact with, and control the process

_"Define the what. Elephant handles the how."_

## Core Concepts

### Work Breakdown Structure (WBS)

Elephant uses a hierarchical decomposition model, refined collaboratively between the user and agents:

1. **Product Vision** — overarching goals
2. **Roadmap Items** — major milestones
3. **Initiatives** — themed bodies of work
4. **Stories** — user-facing deliverables
5. **Tasks** — concrete implementation units
6. **Subtasks / Spikes** — atomic work items or research

### Inverted Flame Graph Branching

Each WBS level maps to a git branch. Narrower levels branch from wider ones. When children complete, they merge back into the parent branch. The parent-level agent — the same session that originally decomposed the work — then reviews and verifies the merged result before marking its level done. This cascades upward: every merge point is a review checkpoint.

**Example:** Tasks A and B branch from Story H. When A and B finish, they merge into H's branch. The original Story H agent session is resumed to review the combined work. Only after verification does H merge upward into its parent Initiative.

### Session Continuity for Reviews

The agent that decomposes a WBS level into children is the same agent that reviews the merged results. This is critical: the decomposing agent has full context of _why_ it broke the work down that way, what the intent was, and what the expected outcome should look like. A fresh agent would lack this context and produce weaker reviews.

The lifecycle of a parent agent session:

1. Parent agent spawns and decomposes work into children
2. Parent agent session is suspended but preserved
3. Child agents execute in their own sessions and containers
4. When children complete and merge into the parent branch, the original parent session is resumed
5. Parent agent reviews the merged result with full decomposition context
6. If the review passes, the parent marks itself done and merges upward
7. If the review fails, the parent can spawn corrective child tasks

### Two-System Architecture

- **Elephant** — the orchestrator. Manages Docker environments, agent lifecycle, git branching, session persistence, secret management, and the TUI
- **Brain** (WIP) — defines the WBS structure and scoped agent skills/instructions. Supports parallel and sequential execution strategies. [github.com/germanamz/brain](https://github.com/germanamz/brain)

## Architecture

### Elephant (Orchestrator)

- Written in Go for performance and cross-platform distribution
- Embeds Tusk as a library for task management, workflow enforcement, and agent coordination
- Manages agent lifecycle — spawns, monitors, suspends, and tears down agents in Docker containers
- Drives the git branching strategy — creates branches per WBS level, manages merges, triggers parent-level reviews
- Exposes an MCP server to agent containers — thin wrapper over Tusk's SDK plus git operations
- Handles agent authentication (per-agent tokens) and secret injection (least-privilege)
- Calls agents in headless mode (`--print --output-format json`) and tracks conversation state for session resumption
- Provides the TUI for monitoring, interaction, and control

### Tusk (Embedded Task Engine)

Elephant embeds [Tusk](https://github.com/germanamz/tusk) as a Go library for task management, workflow enforcement, and agent coordination. Rather than running as a separate MCP server, Tusk is imported directly via its SDK (`github.com/germanamz/tusk`).

Tusk provides:
- **Task hierarchy** — parent-child nesting to arbitrary depth, matching the WBS structure
- **Workflow enforcement** — configurable status transitions with optimistic locking
- **Player management** — agents register as Tusk "players", claim tasks to prevent overlapping work
- **Task queue** — atomic `Pop` operation assigns the highest-urgency unclaimed task to an agent
- **Relations** — typed edges (blocks, relates_to, duplicates) with cycle detection
- **Urgency scoring** — multi-factor ranking for task prioritization

### Brain (WBS & Skills)

- Defines the work breakdown structure from product vision to subtasks
- Provides scoped instructions and skills so agents know how to operate at each WBS level
- Supports both parallel and sequential execution strategies

### Elephant MCP Server

Elephant exposes its own MCP server (network-based) to agent containers. This is a thin wrapper over Tusk's SDK plus Elephant-specific operations:

- **Task tools** — create, claim, complete, annotate, pop (delegated to Tusk)
- **Git tools** — `create_pr` and other git operations that are Elephant's domain
- **Progress reporting** — agents report status via Tusk task annotations

**Agent authentication:** Each agent receives a unique token (injected as an env var at container spawn). Every MCP call must include this token. Elephant validates the token maps to the correct player before executing any operation, preventing agents from impersonating each other.

**Secret management:** Elephant manages secrets (API keys, credentials) and injects them as env vars at container spawn time, scoped per task/project. Agents receive only the secrets relevant to their work — least-privilege by default.

### Docker Isolation Layer

- Each agent runs in its own container with the relevant codebase mounted
- Agents can operate without permission restrictions — the container _is_ the sandbox
- At spawn time, Elephant injects:
  - An **auth token** unique to the agent, used for all MCP communication
  - **Scoped secrets** (API keys, credentials) required for the agent's task
  - **MCP endpoint** for communicating back to Elephant
- After completing work, agents create a PR and enter **standby mode** — the container stays alive so the agent can address review feedback on its PR/MR
- Containers are torn down only after the PR is merged (or the task is otherwise resolved)

### TUI

- Full collaborative interface — supports brainstorming, WBS refinement, and execution monitoring
- Shows agent statuses, task progress, work breakdown tree, and logs
- Allows the user to interact with agents, approve reviews, and intervene when needed

## Workflow

### Phase 1 — Collaborative Refinement

The user and a top-level agent collaboratively define the product vision and WBS through a back-and-forth dialogue in the TUI. Each WBS level is broken down by its own agent session — the product vision agent spawns roadmap-level agents, which spawn initiative-level agents, and so on. Each agent retains its session for later review.

### Phase 2 — Execution

Leaf-level agents (subtasks/spikes) execute in isolated Docker containers. They interact with Elephant's MCP server to manage task state (backed by Tusk's SDK). Elephant monitors progress through the TUI.

### Phase 3 — Cascading Review

When all children of a WBS node complete:

1. Child branches merge into the parent branch
2. Elephant resumes the original parent agent session
3. The parent agent reviews the merged result with full context from the decomposition phase
4. If the review passes, the parent marks itself done and merges upward
5. If the review fails, the parent can spawn corrective child tasks
6. This cascades up the tree until the top-level vision is fulfilled

## Tooling & Quality

### Build & Release

- **Go** — fast compilation, single binary distribution, cross-platform
- **GoReleaser** — automated releases with multi-platform binaries

### Code Quality

- **Lefthook** — git hooks for pre-commit and pre-push checks
- **golangci-lint** — static analysis and linting
- **Conform** — conventional commit enforcement

### Testing

- **E2E tests** — regression testing for agent orchestration behaviors
- **Unit tests** — core logic (branching, session management, container lifecycle)

### Distribution

- Single binary CLI
- Cross-platform: macOS, Linux, Windows

## Phasing

### Phase 1 — Foundation

- Project scaffolding (Go module, CI, linting, commit enforcement)
- Docker container management (spawn, mount codebase, teardown, standby mode, secret/token injection)
- Tusk SDK integration (embedded task engine, player management, task hierarchy)
- Elephant MCP server (thin wrapper over Tusk, agent auth, git tools, secret management)
- Agent execution via Claude Code headless mode (`--print --output-format json`)
- Session persistence and resumption
- Basic TUI with agent status monitoring

### Phase 2 — Orchestration

- Brain integration for WBS definition and agent skills
- Git branching strategy (auto-create branches per WBS level, manage merges)
- Cascading review workflow (resume parent sessions, review merged results)

### Phase 3 — Collaborative TUI

- Back-and-forth WBS refinement through the TUI
- Interactive agent control (pause, resume, intervene)
- Work breakdown tree visualization
- Log streaming and task progress views

### Phase 4 — Scale

- Multi-user support
- Remote/distributed agent execution
- API layer for programmatic access
- Team workflows and shared Elephant instances

## Target User

Starting with solo developers who want to leverage multiple AI agents on their own projects. Designed to grow toward small teams and eventually platform/infra engineers running Elephant as a service for their organization.
