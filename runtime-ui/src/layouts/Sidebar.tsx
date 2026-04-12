/**
 * Sidebar navigation component.
 *
 * Renders the app's navigation menu dynamically from the DSL `navigation[]`
 * section. Each menu item is a link to the page's path. The active item is
 * highlighted based on the current URL.
 *
 * This component reads from the Zustand store — no props needed.
 */

import { NavLink } from "react-router-dom";
import { useAppStore } from "../store/appStore";

export default function Sidebar() {
  const model = useAppStore((state) => state.model);

  // Don't render until the model is loaded.
  if (!model) return null;

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <h2 className="sidebar-title">Quriny</h2>
      </div>

      <nav className="sidebar-nav">
        {model.navigation.map((nav) => (
          <NavLink
            key={nav.label}
            to={nav.path}
            className={({ isActive }) =>
              `sidebar-link ${isActive ? "sidebar-link--active" : ""}`
            }
          >
            {nav.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
