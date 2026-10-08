import { defineConfig, devices } from "@playwright/test";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { CLIENT_ORIGIN, CLIENT_PORT, SERVER_ORIGIN, SERVER_PORT } from "./tests/env";

// Resolve repo-root-relative paths from this config file's location.
// The tests-e2e package lives one level below the repo root; the
// webServer commands run the Go server and Vite client from their
// own subdirectories.
const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, "..");

// Admin token matches `server/Makefile`'s `dev` default so the test
// stack reuses whatever already-running `make server-dev` instance is
// on :8080. Tests import this from ./tests/env.ts too.
//
// The ports come from CMDCTRL_DEV_SERVER_PORT / CMDCTRL_DEV_CLIENT_PORT
// (./tests/env.ts), defaulting to 8080 and 5173. The nightly's shards
// each set their own pair (#2661).
export const ADMIN_TOKEN = "dev-admin-token-not-for-production";

// Data dir mirrors `make server-dev` so the Scryfall index at
// <repo>/data/scryfall/default-cards.json is loaded — deck imports
// depend on it. If the server is already running (reuseExistingServer
// below), this block has no effect; when Playwright spawns its own
// server it points at the same shared dir the dev loop uses.
const serverDataDir = path.join(repoRoot, "data");

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  // Lobby state is global per-process on the server; running tests in
  // parallel would have them race on game IDs and invite tokens. Keep
  // a single worker until we add per-test isolation on the server.
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [["list"], ["html", { open: "never" }]] : "list",
  timeout: 30_000,
  // 10s: three browser contexts, two dev servers and the other E2E
  // runs share one runner. (This used to be blamed on the s19 helpers
  // seeding hands by pumping draw_card, which left pages re-rendering
  // 90-card hands; they move the card straight out of the library now,
  // #2253.)
  expect: { timeout: 10_000 },
  use: {
    baseURL: CLIENT_ORIGIN,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: [
    {
      // Run the Go server directly with `go run` so we don't depend on
      // a prebuilt binary. The env block pins a known admin token and
      // points CMDCTRL_DATA_DIR at the shared dev `<repo>/data` dir
      // (same as `make server-dev`) so the Scryfall index is available;
      // e2e leftovers land in data/games and data/replays.
      command: "go run ./cmd/server",
      cwd: path.join(repoRoot, "server"),
      url: `${SERVER_ORIGIN}/healthz`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: "pipe",
      stderr: "pipe",
      env: {
        CMDCTRL_ADMIN_TOKEN: ADMIN_TOKEN,
        CMDCTRL_DATA_DIR: serverDataDir,
        CMDCTRL_ADDR: `:${SERVER_PORT}`,
        // Lift the lobby join/login rate limits so serial suites that
        // mint many sessions (S19 spins up a fresh 2-player game per
        // test) don't have to sleep between tests. Dev/test only.
        CMDCTRL_DEV_RELAX_RATE_LIMITS: "1",
      },
    },
    {
      // Vite dev server; proxies /ws, /admin, /games, /cards, /me,
      // /healthz to the Go server (see client/vite.config.ts, which
      // reads the same two port variables from the environment this
      // process inherits).
      command: `npm run dev -- --strictPort --port ${CLIENT_PORT}`,
      cwd: path.join(repoRoot, "client"),
      url: CLIENT_ORIGIN,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: "pipe",
      stderr: "pipe",
    },
  ],
});
