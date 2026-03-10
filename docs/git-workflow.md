# Quriny Git Workflow

This document describes how we use Git in the Quriny platform repo: how branches are named, how changes land, and how we cut releases.

## Goals

- Keep `main` always deployable.
- Keep changes small, reviewed, and easy to revert.
- Make releases repeatable and auditable (tags + release notes).

## Branch Strategy (Simplified GitFlow)

We use two long-lived branches and short-lived supporting branches:

- `main` — production-ready, protected
- `develop` — integration branch for upcoming work
- `docs` — documentation purposes
- `feature/*` — new features (short-lived)
- `fix/*` — non-urgent bug fixes (short-lived)
- `release/*` — release stabilization (short-lived)
- `hotfix/*` — urgent production fixes (short-lived)

### Protected branch rules (recommended)

- Disallow direct pushes to `main` (and usually `develop`).
- Require PR review approval.
- Require CI checks to pass.
- Prefer **squash merge** to keep history clean (see “Merging”).

## Branch Naming

Use lowercase, hyphenated names:

- `feature/<short-description>` (e.g. `feature/dsl-v1`)
- `fix/<short-description>` (e.g. `fix/project-create-500`)
- `hotfix/<short-description>` (e.g. `hotfix/auth-token-expiry`)
- `release/vX.Y.Z` (e.g. `release/v0.1.0`)

If you track work items, optionally prefix with an ID:

- `feature/QRY-123-dsl-v1`

## Typical Branch Shape

```text
main
└─ develop
   ├─ feature/dsl-v1
   ├─ feature/builder-ui
   └─ fix/project-create-500
```

## Feature / Fix Development

Create a branch from `develop`, push it, then open a PR back to `develop`.

```bash
git checkout develop
git pull --ff-only

git checkout -b feature/dsl-v1
git push -u origin feature/dsl-v1
```

### Keep your branch up to date

Prefer rebasing onto `develop` to reduce merge noise (before opening a PR or when requested in review):

```bash
git fetch origin
git rebase origin/develop
```

If rebasing is unfamiliar, ask in the PR and we can pair on it.

## Pull Requests

PRs are the only way changes land in `develop` / `main`.

PR checklist (recommended):

- Clear title and description (“what” + “why”).
- Link the issue/task (if applicable).
- Small, focused diff (split large work into multiple PRs).
- Tests added/updated where appropriate.
- No secrets in commits (keys, tokens, `.env`, credentials).

## Merging

Default merge strategy: **Squash merge** into the target branch.

Why:

- Keeps `develop` and `main` linear and readable.
- Makes reverts simple (one commit).

When squashing, ensure the resulting commit message follows Conventional Commits.

## Commit Convention (Conventional Commits)

We use **Conventional Commits** to keep history readable and to support changelog generation:

- Spec: https://www.conventionalcommits.org/

Format:

```text
type(scope): description
```

Examples:

```text
feat(runtime): add dynamic CRUD generator
feat(builder): implement entity editor
fix(api): correct project creation endpoint
docs: update architecture documentation
refactor(runtime): simplify model loader
```

Common types:

| Type     | Meaning                 |
| -------- | ----------------------- |
| feat     | new feature             |
| fix      | bug fix                 |
| docs     | documentation           |
| refactor | internal improvement    |
| test     | tests                   |
| chore    | tooling/maintenance     |
| perf     | performance improvement |
| ci       | CI/build changes        |

## Release Process

Releases are cut from `develop` using a `release/*` branch:

1. Create a release branch:

```bash
git checkout develop
git pull --ff-only
git checkout -b release/v0.1.0
git push -u origin release/v0.1.0
```

2. Stabilize on the release branch:

   - Only allow bug fixes, docs, and release-related changes (version bump, changelog).
   - Avoid merging new features into a release branch.

3. Merge the release to `main` and tag:
   - PR `release/v0.1.0` → `main`
   - After merging to `main`, create an annotated tag:

```bash
git checkout main
git pull --ff-only
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

4. Back-merge to `develop`:
   - PR `release/v0.1.0` → `develop` (or `main` → `develop`, depending on your merge strategy)
   - This ensures `develop` contains the exact released state.

## Hotfix Process (Production)

Hotfixes branch from `main` and merge back into both `main` and `develop`.

```bash
git checkout main
git pull --ff-only
git checkout -b hotfix/auth-token-expiry
git push -u origin hotfix/auth-token-expiry
```

After approval:

- PR `hotfix/*` → `main` (deploy)
- Tag a patch release (e.g. `v0.1.1`) after merge
- PR `main` (or the hotfix branch) → `develop` to keep branches in sync

## Versioning

We follow **Semantic Versioning**: `MAJOR.MINOR.PATCH` (e.g. `v0.1.0`).

- **MAJOR** — breaking changes
- **MINOR** — new features (backwards compatible)
- **PATCH** — bug fixes (backwards compatible)
