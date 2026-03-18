# Internal

## Real Flow (Production System)

User edits app (DSL JSON / UI)
↓
Load DSL
↓
Validate DSL
↓
Build IR (Graph)
↓
[Now DSL is no longer used]
↓
Everything runs from IR

## 🧠 What OutSystems Actually Does (Simplified)

DSL (visual model)
↓
IR (internal graph)
↓
Code Generation (C#/Java)
↓
Compiled app (fast)

- Runtime engine (dynamic parts)

## 🎯 So What Are YOU Building?

For Quriny:

Recommended Architecture (don’t skip this)
Phase 1 (Now)

👉 DSL → IR (Graph)

Phase 2

👉 IR → Code Generation

API

DB schema

basic UI config

Phase 3

👉 Runtime Engine

actions

workflows

dynamic logic
