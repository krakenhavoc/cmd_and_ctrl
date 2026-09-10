import { writable, type Readable } from "svelte/store";

// Service-worker registration and the update handshake.
//
// UPDATE POLICY (ADR 0031): the page is NEVER reloaded on our own
// initiative. A new worker installs in the background and then waits; the
// player is told a new version exists and reloads when they are not in the
// middle of a combat step. That is why nothing here calls skipWaiting()
// except applyUpdate(), and why the controllerchange reload is gated behind
// a flag that only applyUpdate() sets.
//
// The cost of that policy is a stale client running against a newer server.
// It is bounded: the shell cache is build-scoped (see src/sw/service-worker.js),
// the update check re-runs on focus and on a timer, and the prompt comes back
// on the next check after being dismissed.

const UPDATE_POLL_MS = 30 * 60 * 1000;
// Re-checking on every tab focus would hammer the origin during normal
// play, where focus changes constantly.
const FOCUS_CHECK_MIN_INTERVAL_MS = 5 * 60 * 1000;

const updateReadyStore = writable(false);

/** True when a new build has installed and is waiting for the player. */
export const updateReady: Readable<boolean> = { subscribe: updateReadyStore.subscribe };

let waiting: ServiceWorker | null = null;
let accepted = false;
let lastCheck = 0;

/**
 * applyUpdate hands over to the waiting worker and reloads once it has taken
 * control. Only ever called from the player's own click.
 */
export function applyUpdate(): void {
  if (!waiting) return;
  accepted = true;
  updateReadyStore.set(false);
  waiting.postMessage({ type: "SKIP_WAITING" });
  // The reload happens on controllerchange, once the new worker is actually
  // in control -- reloading before that would just re-run the old build.
}

/** dismissUpdate hides the prompt until the next update check finds it again. */
export function dismissUpdate(): void {
  updateReadyStore.set(false);
}

function promote(registration: ServiceWorkerRegistration): void {
  if (!registration.waiting) return;
  // navigator.serviceWorker.controller is null on the very first install.
  // There is no previous version to replace, so there is nothing to prompt
  // about -- the worker simply starts serving.
  if (!navigator.serviceWorker.controller) return;
  waiting = registration.waiting;
  updateReadyStore.set(true);
}

function checkForUpdate(registration: ServiceWorkerRegistration): void {
  const now = Date.now();
  if (now - lastCheck < FOCUS_CHECK_MIN_INTERVAL_MS) return;
  lastCheck = now;
  void registration.update().catch(() => {
    // Offline, or the server is down. The next check will pick it up.
  });
}

/**
 * registerServiceWorker installs /sw.js. No-op in dev: `vite dev` does not
 * emit a worker, and a caching layer in front of HMR is nothing but a
 * debugging trap.
 */
export function registerServiceWorker(): void {
  if (!import.meta.env.PROD) return;
  if (typeof navigator === "undefined" || !("serviceWorker" in navigator)) return;

  navigator.serviceWorker.addEventListener("controllerchange", () => {
    if (!accepted) return;
    window.location.reload();
  });

  window.addEventListener("load", () => {
    void register();
  });
}

async function register(): Promise<void> {
  let registration: ServiceWorkerRegistration;
  try {
    // updateViaCache: "none" keeps the browser's HTTP cache out of the update
    // check. Without it a cached sw.js can pin a client to an old build for
    // up to 24 hours.
    const options: RegistrationOptions = { updateViaCache: "none" };
    registration = await navigator.serviceWorker.register("/sw.js", options);
  } catch {
    // A failed registration is not worth bothering the player about: the app
    // works exactly as it did before there was a service worker.
    return;
  }

  lastCheck = Date.now();
  // A worker may already be waiting from a previous visit.
  promote(registration);

  registration.addEventListener("updatefound", () => {
    const installing = registration.installing;
    if (!installing) return;
    installing.addEventListener("statechange", () => {
      if (installing.state === "installed") promote(registration);
    });
  });

  setInterval(() => checkForUpdate(registration), UPDATE_POLL_MS);
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") checkForUpdate(registration);
  });
}
