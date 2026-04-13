/**
 * App — root component for the Quriny Runtime UI.
 *
 * Responsibilities:
 * 1. Load the DSL app model from the backend on startup
 * 2. Build React Router routes dynamically from the DSL pages
 * 3. Render loading/error states while the model is being fetched
 *
 * Route generation:
 *   Each DSL page (e.g. { name: "ProductList", path: "/products" }) becomes
 *   a React Router route that renders DynamicPage with the page name as a param.
 *   The first navigation item's path is used as the default redirect.
 */

import { useEffect, useCallback } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { useAppStore } from "./store/appStore";
import AppLayout from "./layouts/AppLayout";
import DynamicPage from "./pages/DynamicPage";

export default function App() {
  const model = useAppStore((state) => state.model);
  const isLoading = useAppStore((state) => state.isLoading);
  const error = useAppStore((state) => state.error);

  // Stable reference to loadModel — avoids infinite useEffect re-triggers.
  const loadModel = useCallback(() => useAppStore.getState().loadModel(), []);

  // Fetch the app model once when the app starts.
  useEffect(() => {
    loadModel();
  }, [loadModel]);

  // --- Loading state ---
  if (isLoading) {
    return (
      <div className="app-loading">
        <div className="spinner" />
        <p>Loading application model...</p>
      </div>
    );
  }

  // --- Error state ---
  if (error) {
    return (
      <div className="app-error">
        <h2>Failed to load application</h2>
        <p>{error}</p>
        <button className="btn btn-primary" onClick={loadModel}>
          Retry
        </button>
      </div>
    );
  }

  // --- Model not loaded yet ---
  if (!model) return null;

  // Determine the default redirect path from the first navigation item.
  const defaultPath =
    model.navigation.length > 0 ? model.navigation[0].path : "/";

  return (
    <BrowserRouter>
      <Routes>
        {/* All pages share the same layout (sidebar + main content). */}
        <Route element={<AppLayout />}>
          {/* Generate a route for each DSL page. */}
          {model.pages.map((page) => (
            <Route
              key={page.name}
              path={page.path}
              element={<DynamicPage />}
            />
          ))}

          {/* Default redirect: go to the first navigation page. */}
          <Route path="/" element={<Navigate to={defaultPath} replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
