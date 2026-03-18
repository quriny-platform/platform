# Quriny

Quriny is a low-code platform for building applications from a structured application model instead of hand-writing every layer from scratch.

The long-term inspiration is tools like OutSystems, but with a different development experience: faster iteration, more transparent architecture, and a workflow that can run locally during development instead of depending on direct deployment to a remote server for every change.

## Why Quriny Exists

Traditional low-code platforms can be powerful, but they often make development slower than expected because:

- development is tightly coupled to platform deployment
- local iteration is limited or awkward
- internal platform behavior is hard to inspect or extend
- teams can become locked into a single vendor workflow

Quriny explores a more developer-friendly alternative:

- **local-first development** during the app-building phase
- **model-driven architecture** with a clear, explicit DSL
- **runtime execution** for fast iteration
- **extensible platform design** that can later support code generation, desktop tooling, and mobile targets

## Vision

Quriny is intended to become a platform where users can define an application through a visual builder and/or structured DSL, then have the platform provide:

- data models and CRUD behavior
- pages and navigation
- reusable UI components
- actions and workflows
- runtime APIs and UI rendering

The first focus is **web-based application development**. Later, the platform may expand to:

- mobile application targets
- richer automation and workflow capabilities
- desktop tooling / IDE support
- optional code generation for standalone deployments

## Development Philosophy

The core product idea is simple:

1. During development, app creators should be able to work primarily on `localhost`.
2. The platform should validate and execute an app model quickly.
3. Deployment should be a later step, not the center of the development loop.

This is one of the main ways Quriny differs from traditional server-centered low-code workflows. The goal is to shorten the feedback loop.

## Architecture Overview

Quriny is organized around a model-driven architecture.

```text
Builder UI (authoring)
   |
   | create/update app model
   v
Platform API (projects, auth, model storage)
   |
   | serve model to runtime
   v
Runtime Engine (execute model: APIs + UI schema)
   |
   v
Database
```

### Core Flow

At a high level, the system works like this:

1. A user defines an app using a visual builder or structured DSL.
2. The platform stores that app definition as the source of truth.
3. The runtime engine validates and interprets the model.
4. The runtime exposes APIs and metadata for the UI.
5. The runtime UI renders the application in the browser.

### Internal Runtime Direction

The backend is already moving toward a compiler-style pipeline:

```text
DSL -> Validation -> IR / Graph -> Runtime Execution
```

This gives Quriny a clean internal split:

- the **DSL** is for authoring
- the **IR (intermediate representation)** is for machine-friendly execution
- the **runtime** uses the IR to serve application behavior

In the future, this same IR could also support:

- code generation
- deployment packaging
- mobile targets
- richer workflow engines

## Main Components

### Builder UI

The authoring interface for creating applications.

Planned responsibilities:

- entity editor
- page builder
- component configuration
- preview and publishing flows

Planned stack:

- React
- TypeScript

### Platform API

The backend service that manages projects, authentication, and model storage.

Responsibilities:

- project/workspace management
- model persistence
- validation entry points
- runtime coordination

Current and planned stack:

- Go
- REST APIs

### Runtime Engine

The service that interprets application models and turns them into executable behavior.

Responsibilities:

- load and validate the model
- build intermediate representation (IR)
- expose runtime APIs
- drive page/component behavior

Current and planned stack:

- Go

### Runtime UI

The browser-based renderer for running Quriny applications.

Responsibilities:

- render pages and components dynamically
- connect UI elements to runtime APIs
- manage client-side app state

Planned stack:

- React
- TypeScript

### Future Desktop IDE

A richer IDE experience may later be added for advanced authoring workflows.

Current direction from earlier planning:

- C# for a future desktop IDE

This is not the immediate focus. The current priority is the web platform.

## Technology Direction

The current technology recommendation for Quriny is:

- **Backend:** Go
- **Frontend:** React + TypeScript
- **Future desktop IDE:** C#

Why this combination:

- **Go** keeps backend services fast, simple, and easy to deploy
- **React + TypeScript** provides a strong foundation for dynamic builders and runtime UIs
- **C#** can support a more powerful desktop IDE later if that becomes valuable for advanced authoring workflows

## Current Scope

Quriny is still in an early platform-building stage.

Current repo direction indicates work around:

- DSL/app model definition
- JSON loading and validation
- IR/graph building
- architecture and workflow documentation

The repo is not yet a complete end-user platform. Right now, it is establishing the platform foundation.

## Near-Term Roadmap

The current roadmap points toward these milestones:

1. Foundation and repo workflow
2. DSL v1 definition
3. Runtime Engine MVP
4. Web Runtime MVP
5. Builder UI MVP

Later phases may include:

6. release/distribution improvements
7. code generation
8. workflow automation
9. mobile-oriented expansion

## Repository Structure

```text
.
├── backend
├── builder-ui
├── docker
├── docs
├── examples
├── packages
├── runtime-ui
└── scripts
```

## Documentation

Additional project context lives in:

- [docs/architecture.md](docs/architecture.md)
- [docs/roadmap.md](docs/roadmap.md)
- [docs/git-workflow.md](docs/git-workflow.md)

## Project Status

This project is currently exploratory and foundational. The goal right now is to establish the architecture, model, runtime direction, and developer workflow before expanding into a fuller product experience.

## Summary

Quriny is a local-first, model-driven low-code platform aimed at making application development faster and more flexible than traditional server-centered low-code workflows. The immediate target is web application development, with a long-term path toward richer tooling, broader deployment options, and eventually mobile support.
