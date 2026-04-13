/**
 * Entry point for the Quriny Runtime UI.
 *
 * This file mounts the React app into the DOM. It uses StrictMode to
 * surface potential issues during development (double-renders, deprecated
 * API usage, etc.).
 */

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./index.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
