# Quriny Runtime UI

The browser-based renderer for Quriny applications. It fetches a DSL app model from the backend and dynamically renders pages, tables, forms, and navigation — all without hand-written page code.

## How It Works

```
Backend (Go :8080)                 Runtime UI (React :5173)
┌────────────────┐                 ┌──────────────────────┐
│ GET /app-model  │ ──── fetch ──→ │ Zustand store         │
│                 │                 │   ├─ model (DSL)      │
│ GET /entities/* │ ←── CRUD ────→ │   └─ records (data)   │
│ POST/PUT/DELETE │                 │                       │
└────────────────┘                 │ Dynamic Renderer      │
                                   │   ├─ Router (pages)   │
                                   │   ├─ Sidebar (nav)    │
                                   │   ├─ DataTable         │
                                   │   └─ DataForm          │
                                   └──────────────────────┘
```

1. On startup, the UI calls `GET /app-model` to load the DSL definition
2. React Router routes are built dynamically from the `pages[]` section
3. The sidebar is built from `navigation[]`
4. Each page renders its `components[]` using the **component registry**
5. Components (table, form) fetch and display entity data via the CRUD API

## Project Structure

```
src/
├── api/
│   └── client.ts         # HTTP client for backend API calls
├── components/
│   ├── registry.ts       # Maps DSL component types → React components
│   ├── DataTable.tsx      # Table component (list, edit, delete)
│   └── DataForm.tsx       # Form component (create, edit)
├── layouts/
│   ├── AppLayout.tsx      # Sidebar + main content layout
│   └── Sidebar.tsx        # Dynamic navigation from DSL
├── pages/
│   └── DynamicPage.tsx    # Renders components for the current page
├── store/
│   └── appStore.ts        # Zustand global state (model + records)
├── types/
│   └── model.ts           # TypeScript types mirroring Go DSL structs
├── App.tsx                # Root component with dynamic routing
├── main.tsx               # Entry point
└── index.css              # Global styles
```

## Running Locally

### Prerequisites
- Node.js 20+
- Go backend running on :8080

### Steps

```bash
# 1. Start the Go backend (from platform/backend/)
cd ../backend
go run ./cmd/server -model examples/product_app.json

# 2. Start the frontend dev server (from platform/runtime-ui/)
cd ../runtime-ui
npm install
npm run dev
```

Open http://localhost:5173 — you'll see the Product app rendered from the DSL.

## Adding a New Component Type

1. Create `src/components/MyComponent.tsx` implementing `ComponentProps`
2. Register it in `src/components/registry.ts`:
   ```ts
   import MyComponent from "./MyComponent";
   registry["my-type"] = MyComponent;
   ```
3. Use it in a DSL component definition:
   ```json
   { "name": "MyWidget", "type": "my-type", "entity": "Product", ... }
   ```

## Tech Stack

- **React 19** + TypeScript
- **Vite 5** — fast dev server with HMR
- **React Router 7** — dynamic routing from DSL pages
- **Zustand** — lightweight global state management
