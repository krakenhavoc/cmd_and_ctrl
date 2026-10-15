// serverPing — is the C2 server up, and how far away is it from this
// browser? The site shell's live signal: one GET /healthz (public, no
// session, a few bytes) timed round trip, again every 30 seconds while the
// page is visible, and never while it is hidden.
//
// It says nothing about the server's load or who is on it: table and
// player counts stay off public pages (ADR 0123 keeps /metrics loopback-only
// for exactly that reason). The number is this browser's own round trip.

import type { Readable } from "svelte/store";
import { guardedWritable } from "./guardedStore";

export type Ping = { state: "unknown" } | { state: "online"; ms: number } | { state: "offline" };

export const PING_INTERVAL_MS = 30_000;
const TIMEOUT_MS = 5_000;

// measure times one GET /healthz. Any answer that isn't 2xx, a network
// error or a timeout is "offline".
export async function measure(
  fetchFn: typeof fetch = fetch,
  now: () => number = () => performance.now(),
): Promise<Ping> {
  const ctl = new AbortController();
  const timer = setTimeout(() => ctl.abort(), TIMEOUT_MS);
  const start = now();
  try {
    const res = await fetchFn("/healthz", { cache: "no-store", signal: ctl.signal });
    if (!res.ok) return { state: "offline" };
    return { state: "online", ms: Math.max(1, Math.round(now() - start)) };
  } catch {
    return { state: "offline" };
  } finally {
    clearTimeout(timer);
  }
}

// pingLabel is the chip's text: "Online · 38 ms", "Offline", or "" before
// the first answer (the chip hides until then).
export function pingLabel(p: Ping): string {
  if (p.state === "online") return `Online · ${p.ms} ms`;
  if (p.state === "offline") return "Offline";
  return "";
}

// serverPing measures while anything is subscribed: the first subscriber
// starts it, the last one leaving stops it, so a page without the chip
// sends nothing.
// The header and the front page's table both show it; they share one
// poller. Each start bumps `run`, so a measurement still in flight when the
// last subscriber left can't write or reschedule after a later start.
const store = guardedWritable<Ping>({ state: "unknown" }, "serverPing");
let subscribers = 0;
let run = 0;
let timer: ReturnType<typeof setTimeout> | undefined;

function onVisible(): void {
  if (document.hidden || subscribers === 0) return;
  clearTimeout(timer);
  void tick(run);
}

async function tick(mine: number): Promise<void> {
  if (mine !== run) return;
  if (!document.hidden) {
    const p = await measure();
    if (mine !== run) return;
    store.set(p);
  }
  timer = setTimeout(() => void tick(mine), PING_INTERVAL_MS);
}

function start(): void {
  run += 1;
  document.addEventListener("visibilitychange", onVisible);
  void tick(run);
}

function stop(): void {
  run += 1;
  clearTimeout(timer);
  document.removeEventListener("visibilitychange", onVisible);
}

export const serverPing: Readable<Ping> = {
  subscribe(fn) {
    const unsubscribe = store.subscribe(fn);
    const live = typeof document !== "undefined";
    if (live && ++subscribers === 1) start();
    return () => {
      unsubscribe();
      if (live && --subscribers === 0) stop();
    };
  },
};
