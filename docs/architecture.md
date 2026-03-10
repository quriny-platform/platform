# Quriny Architecture

Quriny is a low-code platform for building web applications using a **DSL (application model)** plus a **runtime engine** that interprets that model to serve APIs and render UI.

This document explains the system at a high level: the main components, how data flows, and how the monorepo is organized.

## Goals

- **Model-driven**: a single source of truth (DSL) defines entities, pages, and behavior.
- **Modular**: components can evolve independently within a monorepo.
- **API-first**: platform capabilities are exposed via HTTP APIs.
- **Runtime execution**: apps run without code generation as the default path (generation can come later).

## System Overview

```text
Builder UI (authoring)
   │
   │  (create/update app model)
   ▼
Platform API (projects, auth, model storage)
   │
   │  (serve app model to runtime)
   ▼
Runtime Engine (execute model: APIs + UI schema)
   │
   ▼
Database (app + platform data)
```

## Core Components

### Builder UI

The web UI used to author applications.

Responsibilities:

- Entity editor (tables/fields/relationships)
- Page builder and layout
- Component configuration (props, bindings)
- Preview/publish flows (through the Platform API)

Tech (current):

- React + TypeScript
- Zustand (state management)

### Platform API

The backend service that manages projects and stores application models.

Responsibilities:

- Project/workspace management
- Authentication/authorization
- DSL storage and versioning (source of truth)
- Publishing/deployment orchestration (as applicable)

Tech (current):

- Go
- REST APIs

### Runtime Engine

The service that executes applications defined by the DSL.

Responsibilities:

- Load and validate the app model
- Provide runtime APIs (e.g., CRUD endpoints derived from entities)
- Apply business rules/workflows (when present)
- Emit UI schema/metadata consumed by the Runtime UI

Tech (current):

- Go

### Runtime UI

The client that renders a running Quriny application.

Responsibilities:

- Render components dynamically based on runtime schema
- Bind UI to runtime APIs (data fetching, mutations)
- Manage UI state and navigation

Tech (current):

- React

## Data & Model Flow

At a high level:

1. Builder UI writes the app model (DSL) via Platform API.
2. Platform API persists the model (and metadata such as versions).
3. Runtime Engine reads the model and exposes runtime APIs + UI schema.
4. Runtime UI renders pages and calls runtime APIs for data.

Notes:

- The DSL is the **contract** between authoring (Builder) and execution (Runtime).
- Keep the DSL backward-compatible where possible; when breaking changes are needed, version the schema.

## Monorepo Layout

```text
platform/
├── backend/      # Platform API + runtime services
├── builder-ui/   # Authoring UI
├── runtime-ui/   # Runtime UI renderer
├── packages/     # Shared libs (types, UI components, tooling)
├── examples/     # Example apps and templates
├── docs/         # Documentation
├── docker/       # Container files and compose manifests
└── scripts/      # Dev tooling and automation
```

## Cross-Cutting Concerns (Guidelines)

- **Security**: authenticate every request; authorize by project/workspace; avoid storing secrets in models.
- **Observability**: structured logs + correlation IDs across API/runtime; basic metrics for request latency/errors.
- **Validation**: validate DSL at write time (Platform API) and at read/execute time (Runtime Engine).
- **Compatibility**: introduce schema changes with migrations and/or version gates.

## Non-Goals (for now)

- Supporting every database/hosting combination out of the gate (start with a supported “happy path”).
- Allowing direct runtime mutations of the DSL (runtime should treat models as immutable inputs).

## Glossary

- **DSL / App Model**: the structured definition of an application (entities, pages, components, actions).
- **Runtime Engine**: server-side interpreter of the model.
- **Runtime UI**: client-side renderer of the model.
