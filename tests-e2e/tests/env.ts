// Shared test constants. Uses the same default admin token as
// `server/Makefile`'s `dev` target ("dev-admin-token-not-for-production")
// so the test suite runs against a `make server-dev` stack without
// extra config.

export const ADMIN_TOKEN = "dev-admin-token-not-for-production";

// Ports of the Go server and the Vite client the suite drives. They
// default to `make server-dev` / `make client-dev`'s 8080 and 5173, so a
// local run is unchanged. The nightly's Playwright shards set them to a
// distinct pair each (#2661): the self-hosted runners share one host, so
// two shards on the default ports would talk to each other's servers.
// client/vite.config.ts reads the same two variables, so its proxy
// follows the server wherever it is.
function portFromEnv(name: string, fallback: number): number {
  const raw = process.env[name];
  if (raw === undefined || raw === "") return fallback;
  const port = Number(raw);
  if (!Number.isInteger(port) || port <= 0 || port > 65535) {
    throw new Error(`${name}=${raw} is not a port`);
  }
  return port;
}

export const SERVER_PORT = portFromEnv("CMDCTRL_DEV_SERVER_PORT", 8080);
export const CLIENT_PORT = portFromEnv("CMDCTRL_DEV_CLIENT_PORT", 5173);
export const SERVER_ORIGIN = `http://localhost:${SERVER_PORT}`;
export const SERVER_WS_ORIGIN = `ws://localhost:${SERVER_PORT}`;
export const CLIENT_ORIGIN = `http://localhost:${CLIENT_PORT}`;
