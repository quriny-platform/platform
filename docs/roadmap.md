# Quriny Roadmap

This roadmap communicates direction and priorities for the Quriny platform. It is **not a promise**: scope and ordering can change as we learn.

How to change this roadmap:

- Propose changes via PR to `platform/docs/roadmap.md`.
- Keep milestones outcome-based (deliverables), not just “things to build”.

## Planning Principles

- Ship thin vertical slices (end-to-end) over large subsystems.
- Prefer runtime-first (interpret the model) before code generation.
- Keep the DSL stable: version and validate schema changes.
- Invest early in developer experience: tests, CI, and docs.

## Roadmap Format

We organize work as milestones. Each milestone has:

- **Goal**: what users can do after it ships
- **Deliverables**: concrete outcomes
- **Notes**: scope boundaries / dependencies

## Milestones

### M0 — Foundation

Goal: a workable repo with repeatable development and contributor flow.

Deliverables:

- Monorepo structure and baseline services/apps
- Git workflow + contribution guidelines
- Core docs: architecture + roadmap
- Basic CI (lint/test/build) to protect `main`

### M1 — DSL v1 (Application Model)

Goal: a versioned DSL that can represent a basic app.

Deliverables:

- DSL specification (entities, pages, components, navigation)
- JSON schema + validation rules
- Migration strategy for future schema changes (even if minimal)

Notes:

- Keep v1 intentionally small; avoid “everything” in the first schema.

### M2 — Runtime Engine MVP

Goal: run a DSL-defined app with a minimal backend.

Deliverables:

- Model loader + validator
- CRUD API generation for entities
- Database integration for CRUD
- Runtime HTTP server with basic auth integration (as needed)

### M3 — Web Runtime MVP

Goal: render a DSL-defined app in the browser.

Deliverables:

- Dynamic component rendering from runtime schema
- Core components: table, form, input controls
- Data binding to runtime APIs (read/write)

### M4 — Builder UI MVP

Goal: create and preview a basic app without editing raw JSON.

Deliverables:

- Entity editor (fields + types)
- Page builder (layout + component placement)
- Component inspector (props + bindings)
- Preview flow that runs against the runtime

### M5 — Release Process + Distribution

Goal: make releases repeatable and easy to consume.

Deliverables:

- Versioning + release tagging process in practice
- Release notes template (what shipped, breaking changes)
- Example apps + “getting started” path

### M6 — Code Generation (Optional Track)

Goal: export an app as a standalone project when runtime hosting is not desired.

Deliverables:

- Backend generation (Go) and frontend generation
- Deployment templates

Notes:

- This is optional and should not block runtime milestones.

### M7 — Workflow / Automation Engine (Optional Track)

Goal: automation primitives for actions, workflows, and triggers.

Deliverables:

- Action system (invoke server-side logic)
- Workflow execution and state model
- Event triggers (time/data-driven)

## What “Done” Means

A milestone is considered done when:

- The core user journey works end-to-end (happy path).
- Docs exist for the feature and how to run it locally.
- Basic automated tests cover critical paths.

## Status Tracking (Recommended)

Track milestone work in the issue tracker using labels like:

- `milestone:M1` / `milestone:M2`
- `area:builder` / `area:runtime` / `area:api`
- `type:feat` / `type:fix` / `type:docs`
