/**
 * Global application store using Zustand.
 *
 * This store holds two types of state:
 * 1. **App Model** — the DSL definition fetched once at startup from GET /app-model.
 *    It tells the runtime which entities, pages, components, actions, and navigation
 *    items exist. This rarely changes during a session.
 *
 * 2. **Entity Data** — the actual records for each entity, fetched on demand.
 *    For example, when the user navigates to the Product list page, we fetch
 *    all Product records and store them here.
 *
 * Why Zustand:
 * - Minimal boilerplate compared to Redux
 * - No providers needed — just import and use
 * - Built-in support for async actions
 *
 * @see https://github.com/pmndrs/zustand
 */

import { create } from "zustand";
import type {
  AppModel,
  Entity,
  Component,
  Action,
  Page,
} from "../types/model";
import * as api from "../api/client";

// ---------------------------------------------------------------------------
// Store State & Actions
// ---------------------------------------------------------------------------

interface AppState {
  /** The DSL app model (null until loaded). */
  model: AppModel | null;

  /** Loading state for the initial model fetch. */
  isLoading: boolean;

  /** Error message if model loading fails. */
  error: string | null;

  /** Entity records keyed by entity name → array of records. */
  records: Record<string, Record<string, unknown>[]>;

  /** Loading state per entity (for data fetches). */
  entityLoading: Record<string, boolean>;

  // --- Actions ---

  /** Fetch the app model from the backend. Called once at startup. */
  loadModel: () => Promise<void>;

  /** Fetch all records for an entity from the backend. */
  loadRecords: (entityName: string) => Promise<void>;

  /** Create a new record and refresh the entity's record list. */
  createRecord: (
    entityName: string,
    data: Record<string, unknown>
  ) => Promise<void>;

  /** Update a record and refresh the entity's record list. */
  updateRecord: (
    entityName: string,
    recordId: string,
    data: Record<string, unknown>
  ) => Promise<void>;

  /** Delete a record and refresh the entity's record list. */
  deleteRecord: (entityName: string, recordId: string) => Promise<void>;

  // --- Lookup helpers ---

  /** Find an entity definition by name. */
  getEntity: (name: string) => Entity | undefined;

  /** Find a component definition by name. */
  getComponent: (name: string) => Component | undefined;

  /** Find an action definition by name. */
  getAction: (name: string) => Action | undefined;

  /** Find a page definition by name. */
  getPage: (name: string) => Page | undefined;
}

// ---------------------------------------------------------------------------
// Store Implementation
// ---------------------------------------------------------------------------

export const useAppStore = create<AppState>((set, get) => ({
  // Initial state
  model: null,
  isLoading: false,
  error: null,
  records: {},
  entityLoading: {},

  // --- Async actions ---

  loadModel: async () => {
    set({ isLoading: true, error: null });
    try {
      const model = await api.fetchAppModel();
      set({ model, isLoading: false });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to load app model";
      set({ error: message, isLoading: false });
    }
  },

  loadRecords: async (entityName: string) => {
    set((state) => ({
      entityLoading: { ...state.entityLoading, [entityName]: true },
    }));
    try {
      const response = await api.listRecords(entityName);
      set((state) => ({
        records: { ...state.records, [entityName]: response.items },
        entityLoading: { ...state.entityLoading, [entityName]: false },
      }));
    } catch (err) {
      console.error(`Failed to load records for ${entityName}:`, err);
      set((state) => ({
        entityLoading: { ...state.entityLoading, [entityName]: false },
      }));
    }
  },

  createRecord: async (entityName, data) => {
    await api.createRecord(entityName, data);
    // Refresh the record list after creating.
    await get().loadRecords(entityName);
  },

  updateRecord: async (entityName, recordId, data) => {
    await api.updateRecord(entityName, recordId, data);
    // Refresh the record list after updating.
    await get().loadRecords(entityName);
  },

  deleteRecord: async (entityName, recordId) => {
    await api.deleteRecord(entityName, recordId);
    // Refresh the record list after deleting.
    await get().loadRecords(entityName);
  },

  // --- Lookup helpers ---

  getEntity: (name) => {
    return get().model?.entities.find((e) => e.name === name);
  },

  getComponent: (name) => {
    return get().model?.components.find((c) => c.name === name);
  },

  getAction: (name) => {
    return get().model?.actions.find((a) => a.name === name);
  },

  getPage: (name) => {
    return get().model?.pages.find((p) => p.name === name);
  },
}));
