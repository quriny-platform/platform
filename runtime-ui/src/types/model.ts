/**
 * Type definitions for the Quriny DSL model.
 *
 * These types mirror the Go structs in `backend/internal/dsl/model.go`.
 * The runtime UI fetches the app model as JSON from GET /app-model
 * and deserialises it into these types.
 */

/** Field defines a single property on an entity (e.g. { name: "price", type: "number" }). */
export interface Field {
  id: string;
  name: string;
  type: "uuid" | "string" | "text" | "number" | "boolean" | "date";
}

/** Entity describes a business object (e.g. Product, Order). */
export interface Entity {
  id: string;
  name: string;
  fields: Field[];
}

/** Page describes a route and the components rendered on it. */
export interface Page {
  id: string;
  name: string;
  path: string;
  components: string[];
}

/** Component describes a UI building block bound to data and actions. */
export interface Component {
  id: string;
  name: string;
  type: "table" | "form" | string;
  entity: string;
  fields: string[];
  actions: string[];
}

/** Action describes a CRUD operation that can be triggered from the UI. */
export interface Action {
  id: string;
  name: string;
  type: "create" | "read" | "update" | "delete";
  entity: string;
}

/** Navigation describes a menu item linking to a page. */
export interface Navigation {
  id: string;
  label: string;
  path: string;
  page: string;
}

/**
 * AppModel is the top-level DSL document for a Quriny application.
 * This is the JSON structure returned by GET /app-model.
 */
export interface AppModel {
  entities: Entity[];
  pages: Page[];
  components: Component[];
  actions: Action[];
  navigation: Navigation[];
}
