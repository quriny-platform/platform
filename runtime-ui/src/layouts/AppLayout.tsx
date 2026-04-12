/**
 * AppLayout provides the top-level page structure: a sidebar for navigation
 * and a main content area where pages are rendered.
 *
 * The <Outlet /> from React Router renders the matched child route inside
 * the main content area, so navigation stays persistent across page changes.
 */

import { Outlet } from "react-router-dom";
import Sidebar from "./Sidebar";

export default function AppLayout() {
  return (
    <div className="app-layout">
      <Sidebar />
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
