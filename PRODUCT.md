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

### Three-System Architecture

- **Elephant** — the orchestrator. Manages Docker environments, agent lifecycle, git branching, session persistence, and the TUI
- **Brain** (WIP) — defines the WBS structure and scoped agent skills/instructions. Supports parallel and sequential execution strategies. [github.com/germanamz/brain](https://github.com/germanamz/brain)
- **Tusk** — CLI task management with a built-in MCP server. Agents interact with Tusk via MCP to pick up tasks, report progress, and mark completion. Acts as the shared state layer. [github.com/germanamz/tusk](https://github.com/germanamz/tusk)

## Architecture

### Elephant (Orchestrator)

- Written in Go for performance and cross-platform distribution
- Manages agent lifecycle — spawns, monitors, and terminates agents in Docker containers
- Drives the git branching strategy — creates branches per WBS level, manages merges, triggers parent-level reviews
- Calls agents in headless mode (`--print --output-format json`) and tracks conversation state for session resumption
- Provides the TUI for monitoring, interaction, and control

### Brain (WBS & Skills)

- Defines the work breakdown structure from product vision to subtasks
- Provides scoped instructions and skills so agents know how to operate at each WBS level
- Supports both parallel and sequential execution strategies

### Tusk (Task Management)

- CLI task management tool with built-in MCP server
- Agents interact with Tusk via MCP to pick up tasks, report progress, and mark completion
- Acts as the shared state layer between Elephant and its agents

### Docker Isolation Layer

- Each agent runs in its own container with the relevant codebase mounted
- Agents can operate without permission restrictions — the container _is_ the sandbox
- Containers are ephemeral — spun up per task, torn down after completion

### TUI

- Full collaborative interface — supports brainstorming, WBS refinement, and execution monitoring
- Shows agent statuses, task progress, work breakdown tree, and logs
- Allows the user to interact with agents, approve reviews, and intervene when needed

## Workflow

### Phase 1 — Collaborative Refinement

The user and a top-level agent collaboratively define the product vision and WBS through a back-and-forth dialogue in the TUI. Each WBS level is broken down by its own agent session — the product vision agent spawns roadmap-level agents, which spawn initiative-level agents, and so on. Each agent retains its session for later review.

### Phase 2 — Execution

Leaf-level agents (subtasks/spikes) execute in isolated Docker containers. They interact with Tusk via MCP to manage task state. Elephant monitors progress through the TUI.

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
- Docker container management (spawn, mount codebase, teardown)
- Agent execution via Claude Code headless mode (`--print --output-format json`)
- Session persistence and resumption
- Basic TUI with agent status monitoring

### Phase 2 — Orchestration

- Tusk integration via MCP for task state management
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
