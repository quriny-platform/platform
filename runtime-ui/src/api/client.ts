/**
 * API client for the Quriny runtime backend.
 *
 * All HTTP calls go through this module so the rest of the app doesn't need
 * to know about URLs, headers, or error handling details. During development,
 * Vite's proxy forwards these requests to the Go backend on :8080.
 *
 * @see vite.config.ts for proxy configuration.
 */

import type { AppModel } from "../types/model";

/** Base URL — empty string because Vite proxies to the backend. */
const BASE = "";

/**
 * Generic fetch wrapper that handles JSON parsing and error responses.
 * Throws an Error with the server's error message if the response is not OK.
 */
async function request<T>(
  url: string,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(`${BASE}${url}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  // DELETE returns 204 No Content — no body to parse.
  if (response.status === 204) {
    return undefined as T;
  }

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || `Request failed: ${response.status}`);
  }

  return data as T;
}

// ---------------------------------------------------------------------------
// App Model
// ---------------------------------------------------------------------------

/** Fetches the full DSL app model from the backend. */
export async function fetchAppModel(): Promise<AppModel> {
  return request<AppModel>("/app-model");
}

// ---------------------------------------------------------------------------
// Entity CRUD — dynamic operations based on entity name
// ---------------------------------------------------------------------------

/** Response shape for listing entity records. */
export interface ListResponse {
  entity: string;
  count: number;
  items: Record<string, unknown>[];
}

/** List all records for an entity. */
export async function listRecords(
  entityName: string
): Promise<ListResponse> {
  return request<ListResponse>(`/entities/${entityName}`);
}

/** Get a single record by entity name and record ID. */
export async function getRecord(
  entityName: string,
  recordId: string
): Promise<Record<string, unknown>> {
  return request<Record<string, unknown>>(
    `/entities/${entityName}/${recordId}`
  );
}

/** Create a new record for an entity. */
export async function createRecord(
  entityName: string,
  data: Record<string, unknown>
): Promise<Record<string, unknown>> {
  return request<Record<string, unknown>>(`/entities/${entityName}`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

/** Update an existing record (full replace). */
export async function updateRecord(
  entityName: string,
  recordId: string,
  data: Record<string, unknown>
): Promise<Record<string, unknown>> {
  return request<Record<string, unknown>>(
    `/entities/${entityName}/${recordId}`,
    {
      method: "PUT",
      body: JSON.stringify(data),
    }
  );
}

/** Delete a record by entity name and record ID. */
export async function deleteRecord(
  entityName: string,
  recordId: string
): Promise<void> {
  return request<void>(`/entities/${entityName}/${recordId}`, {
    method: "DELETE",
  });
}
