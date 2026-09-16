// clientErrors.ts — a small ring buffer of JavaScript-level failures
// (uncaught errors, unhandled rejections, console.error/warn) so a bug
// report can carry the stack trace the reporter never thought to open
// the devtools for.
//
// This is the half of "attach the logs" that GameClient.log can't
// cover: that buffer records protocol traffic, which tells you what
// the server said, but says nothing about the TypeError in a Svelte
// effect that swallowed the update. In practice the second is the
// commoner bug, and the reporter's description of it is "the board
// froze".
//
// Deliberately not a Svelte store: nothing renders this live, and
// installing it must not depend on component lifecycle. main.ts calls
// installErrorCapture() once at boot; everything else just reads.

import { redactSecrets } from "./redact";

// CLIENT_ERROR_LIMIT bounds the buffer. Errors in a broken render loop
// arrive in the thousands, and only the first few are distinct — a
// small window keeps the interesting ones without unbounded growth.
export const CLIENT_ERROR_LIMIT = 50;

export interface ClientErrorEntry {
  // Epoch milliseconds — the server formats these, so no timezone or
  // locale logic lives on this side.
  at: number;
  text: string;
}

let buffer: ClientErrorEntry[] = [];
let installed = false;

// TEXT_MAX clips one entry. A minified stack trace can run to
// kilobytes; the server clips again, this just avoids holding megabytes
// of duplicated frames in memory.
const TEXT_MAX = 400;

// recordClientError appends an entry, dropping the oldest when full.
// Exported so other modules can record a failure they handled but
// still want visible in a report.
//
// Redacted before it is clipped (#721): a console.error that stringifies
// a session object or a fetch URL must not carry the token into a bug
// report, and clipping first could cut the key off a value and leave
// the value behind.
export function recordClientError(text: string): void {
  const entry: ClientErrorEntry = {
    at: Date.now(),
    text: redactSecrets(text).slice(0, TEXT_MAX),
  };
  buffer = buffer.length >= CLIENT_ERROR_LIMIT ? [...buffer.slice(1), entry] : [...buffer, entry];
}

// recentClientErrors returns the buffer oldest-first. The caller merges
// it with GameClient.log by timestamp.
export function recentClientErrors(): ClientErrorEntry[] {
  return buffer;
}

// resetClientErrors empties the buffer. For tests.
export function resetClientErrors(): void {
  buffer = [];
}

// describeReason renders an unhandled-rejection reason. Rejections
// carry anything at all — an Error, a Response, a bare string, or
// undefined — so this has to cope with all of it without throwing
// inside the error handler.
function describeReason(reason: unknown): string {
  if (reason instanceof Error) {
    return `${reason.name}: ${reason.message}`;
  }
  if (typeof reason === "string") return reason;
  try {
    return JSON.stringify(reason) ?? String(reason);
  } catch {
    return String(reason);
  }
}

// formatArgs renders console arguments the way the console would,
// without pulling in a formatter. Errors get their message rather than
// "[object Object]", which is the whole point of capturing them.
function formatArgs(args: unknown[]): string {
  return args
    .map((a) => {
      if (a instanceof Error) return `${a.name}: ${a.message}`;
      if (typeof a === "string") return a;
      try {
        return JSON.stringify(a) ?? String(a);
      } catch {
        return String(a);
      }
    })
    .join(" ");
}

// installErrorCapture wires the listeners and wraps console.error /
// console.warn. Returns a teardown function, and is idempotent — a
// second call is a no-op, so a hot-reload can't stack three layers of
// console wrapper.
//
// The console wrappers always chain to the original: capturing must
// never cost the developer their devtools output.
export function installErrorCapture(
  win: Pick<Window, "addEventListener" | "removeEventListener"> = window,
  con: Pick<Console, "error" | "warn"> = console,
): () => void {
  if (installed) return () => {};
  installed = true;

  const onError = (ev: Event): void => {
    const e = ev as ErrorEvent;
    const where = e.filename ? ` (${e.filename}:${e.lineno}:${e.colno})` : "";
    recordClientError(`uncaught ${e.message}${where}`);
  };
  const onRejection = (ev: Event): void => {
    const e = ev as PromiseRejectionEvent;
    recordClientError(`unhandled rejection: ${describeReason(e.reason)}`);
  };
  win.addEventListener("error", onError);
  win.addEventListener("unhandledrejection", onRejection);

  // Keep the original references, not bound copies: teardown restores
  // these properties by identity, and handing back a wrapper would mean
  // an uninstall that isn't one.
  const origError = con.error;
  const origWarn = con.warn;
  con.error = (...args: unknown[]): void => {
    recordClientError(`console.error: ${formatArgs(args)}`);
    origError.apply(con, args);
  };
  con.warn = (...args: unknown[]): void => {
    recordClientError(`console.warn: ${formatArgs(args)}`);
    origWarn.apply(con, args);
  };

  return () => {
    win.removeEventListener("error", onError);
    win.removeEventListener("unhandledrejection", onRejection);
    con.error = origError;
    con.warn = origWarn;
    installed = false;
  };
}
