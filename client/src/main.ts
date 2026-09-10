import { mount } from "svelte";
import App from "./App.svelte";
import { installErrorCapture } from "./lib/clientErrors";
import { registerServiceWorker } from "./lib/pwa";
import "./app.css";

// Installed before the app mounts so a failure during initial render is
// captured too. Feeds the in-app bug report's activity log (ADR 0017
// §7); costs nothing when no report is ever filed.
installErrorCapture();

// PWA (ADR 0031). Production builds only, and it never reloads the page on
// its own — a waiting worker surfaces as a prompt the player accepts.
registerServiceWorker();

const target = document.getElementById("app");
if (!target) {
  throw new Error("missing #app element");
}

mount(App, { target });
