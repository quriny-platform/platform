/**
 * DynamicPage renders a DSL-defined page by looking up its components
 * in the registry and rendering them in order.
 *
 * How it works:
 * 1. React Router matches a URL path to a DSL page (e.g. /products → "ProductList")
 * 2. DynamicPage reads the current URL path via useLocation()
 * 3. It finds the page whose DSL path matches the current URL
 * 4. For each component listed on the page, it resolves the React component
 *    from the registry and renders it
 *
 * This is the bridge between the static DSL definition and the live React UI.
 */

import { useLocation } from "react-router-dom";
import { useAppStore } from "../store/appStore";
import { resolveComponent } from "../components/registry";

export default function DynamicPage() {
  const model = useAppStore((state) => state.model);
  const location = useLocation();

  if (!model) {
    return <div className="page-error">App model not loaded.</div>;
  }

  // Find the page definition whose path matches the current URL.
  const page = model.pages.find((p) => p.path === location.pathname);
  if (!page) {
    return (
      <div className="page-error">
        No page defined for path "{location.pathname}".
      </div>
    );
  }

  return (
    <div className="dynamic-page">
      <h1 className="page-title">{page.name}</h1>

      {/* Render each component declared on this page. */}
      {page.components.map((componentName) => {
        // Look up the component definition in the DSL model.
        const componentDef = model.components.find(
          (c) => c.name === componentName
        );

        if (!componentDef) {
          return (
            <div key={componentName} className="component-error">
              Component "{componentName}" is not defined in the app model.
            </div>
          );
        }

        // Resolve the DSL type (e.g. "table") to a React component.
        const ReactComponent = resolveComponent(componentDef.type);

        if (!ReactComponent) {
          return (
            <div key={componentName} className="component-error">
              No renderer found for component type "{componentDef.type}".
            </div>
          );
        }

        return (
          <ReactComponent key={componentName} component={componentDef} />
        );
      })}
    </div>
  );
}
