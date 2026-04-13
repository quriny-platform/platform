/**
 * Component Registry.
 *
 * Maps DSL component `type` values (e.g. "table", "form") to the React
 * component that should render them. This is the core pattern that makes
 * Quriny's runtime dynamic — the DSL author declares a component type,
 * and the registry resolves it to real UI.
 *
 * To add a new component type:
 * 1. Create the React component in `src/components/`
 * 2. Register it here with its DSL type name
 *
 * Example:
 *   registry["chart"] = ChartComponent;
 */

import type { Component as DSLComponent } from "../types/model";
import DataTable from "./DataTable";
import DataForm from "./DataForm";

/**
 * Props that every registered component receives from the DynamicPage renderer.
 * This is the contract between the page renderer and the individual components.
 */
export interface ComponentProps {
  /** The full DSL component definition (name, type, entity, fields, actions). */
  component: DSLComponent;
}

/** Registry mapping DSL type string → React component. */
const registry: Record<string, React.FC<ComponentProps>> = {
  table: DataTable,
  form: DataForm,
};

/**
 * Resolves a DSL component type to its React component.
 * Returns undefined if the type is not registered.
 */
export function resolveComponent(
  type: string
): React.FC<ComponentProps> | undefined {
  return registry[type];
}
