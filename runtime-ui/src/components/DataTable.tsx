/**
 * DataTable component.
 *
 * Renders entity records in an HTML table. Columns are driven by the
 * `fields[]` array in the DSL component definition, and rows come from
 * the entity records fetched via the CRUD API.
 *
 * Features:
 * - Fetches records on mount and when the entity name changes
 * - Shows a loading spinner while fetching
 * - Renders action buttons (Edit, Delete) based on DSL actions
 * - Navigates to the form page for editing
 *
 * Usage (via DSL):
 *   { "name": "ProductTable", "type": "table", "entity": "Product",
 *     "fields": ["name", "price"], "actions": ["UpdateProduct", "DeleteProduct"] }
 */

import { useEffect, useCallback, useMemo } from "react";
import { useAppStore } from "../store/appStore";
import type { ComponentProps } from "./registry";

/** Stable empty array to avoid creating new references on every render. */
const EMPTY_RECORDS: Record<string, unknown>[] = [];

export default function DataTable({ component }: ComponentProps) {
  const recordsRaw = useAppStore((state) => state.records[component.entity]);
  const records = recordsRaw ?? EMPTY_RECORDS;
  const isLoading = useAppStore(
    (state) => !!state.entityLoading[component.entity]
  );
  const model = useAppStore((state) => state.model);

  // Stable references to store actions — avoids infinite useEffect loops.
  const loadRecords = useCallback(
    () => useAppStore.getState().loadRecords(component.entity),
    [component.entity]
  );
  const handleDeleteRecord = useCallback(
    (recordId: string) => useAppStore.getState().deleteRecord(component.entity, recordId),
    [component.entity]
  );

  // Fetch records when the component mounts or entity changes.
  useEffect(() => {
    loadRecords();
  }, [loadRecords]);

  /**
   * Check if a specific action type is available for this component.
   * For example, hasAction("delete") returns true if any action in the
   * component's actions[] list has type "delete".
   */
  const hasAction = (actionType: string): boolean => {
    if (!model) return false;
    return component.actions.some((actionName) => {
      const action = model.actions.find((a) => a.name === actionName);
      return action?.type === actionType;
    });
  };

  /**
   * Find the form page path for the entity so we can navigate to it
   * for creating or editing records.
   */
  const getFormPagePath = (): string | null => {
    if (!model) return null;
    // Look for a page that contains a "form" component bound to the same entity.
    for (const page of model.pages) {
      for (const compName of page.components) {
        const comp = model.components.find((c) => c.name === compName);
        if (comp?.type === "form" && comp.entity === component.entity) {
          return page.path;
        }
      }
    }
    return null;
  };

  /** Handle delete button click with confirmation. */
  const handleDelete = async (recordId: string) => {
    if (!window.confirm("Are you sure you want to delete this record?")) return;
    try {
      await handleDeleteRecord(recordId);
    } catch (err) {
      console.error("Failed to delete record:", err);
    }
  };

  // --- Derived state ---
  const canCreate = hasAction("create");
  const canUpdate = hasAction("update");
  const canDelete = hasAction("delete");
  const formPagePath = getFormPagePath();

  // --- Render ---

  if (isLoading) {
    return <div className="loading">Loading {component.entity} records...</div>;
  }

  return (
    <div className="data-table-container">
      {/* Header with entity name and create button */}
      <div className="data-table-header">
        <h2>{component.entity}</h2>
        {canCreate && formPagePath && (
          <a href={formPagePath} className="btn btn-primary">
            + New {component.entity}
          </a>
        )}
      </div>

      {/* Empty state */}
      {records.length === 0 && (
        <p className="empty-state">No {component.entity} records found.</p>
      )}

      {/* Data table */}
      {records.length > 0 && (
        <table className="data-table">
          <thead>
            <tr>
              {/* Render column headers from DSL fields */}
              {component.fields.map((field) => (
                <th key={field}>{field}</th>
              ))}
              {/* Actions column (if update or delete is available) */}
              {(canUpdate || canDelete) && <th>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {records.map((record) => {
              const recordId = String(record.id ?? "");
              return (
                <tr key={recordId}>
                  {/* Render cell values for each field */}
                  {component.fields.map((field) => (
                    <td key={field}>{String(record[field] ?? "")}</td>
                  ))}

                  {/* Action buttons */}
                  {(canUpdate || canDelete) && (
                    <td className="actions-cell">
                      {canUpdate && formPagePath && (
                        <a
                          href={`${formPagePath}?edit=${recordId}`}
                          className="btn btn-sm"
                        >
                          Edit
                        </a>
                      )}
                      {canDelete && (
                        <button
                          onClick={() => handleDelete(recordId)}
                          className="btn btn-sm btn-danger"
                        >
                          Delete
                        </button>
                      )}
                    </td>
                  )}
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}
