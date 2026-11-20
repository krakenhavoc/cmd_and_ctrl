// clock.ts — a pausable setTimeout for the fx preview harness (#4868).
//
// A cue's parts are Web Animations, but its later beats start on timers
// and its copy is removed by one. Pausing the animations alone would let
// those timers fire and clear a frozen cue, so the harness swaps in this
// setTimeout in (main.ts, before the page mounts): while the
// clock is paused, every pending timer waits with the time it had left.
// Only the harness page installs it; the app never imports this file.

type Timer = { run: () => void; due: number; left: number; real?: number };

const realSet = window.setTimeout.bind(window);
const realClear = window.clearTimeout.bind(window);
const timers = new Map<number, Timer>();
let nextID = 1;
let paused = false;

export function installClock(): void {
  window.setTimeout = ((handler: TimerHandler, ms = 0, ...args: unknown[]) => {
    if (typeof handler !== "function") return realSet(handler, ms, ...args);
    const id = nextID++;
    const t: Timer = {
      run: () => {
        timers.delete(id);
        (handler as (...a: unknown[]) => void)(...args);
      },
      due: performance.now() + ms,
      left: ms,
    };
    timers.set(id, t);
    if (!paused) t.real = realSet(t.run, ms);
    return id;
  }) as typeof window.setTimeout;
  window.clearTimeout = ((id?: number) => {
    const t = id === undefined ? undefined : timers.get(id);
    if (!t) return realClear(id);
    if (t.real !== undefined) realClear(t.real);
    timers.delete(id!);
  }) as typeof window.clearTimeout;
}

/** Holds every pending timer, and every animation on the page. */
export function pauseClock(): void {
  if (paused) return;
  paused = true;
  const now = performance.now();
  for (const t of timers.values()) {
    if (t.real !== undefined) realClear(t.real);
    t.real = undefined;
    t.left = Math.max(0, t.due - now);
  }
  for (const a of document.getAnimations()) a.pause();
}

/** Lets them all run on from where they stopped. */
export function resumeClock(): void {
  if (!paused) return;
  paused = false;
  const now = performance.now();
  for (const t of timers.values()) {
    t.due = now + t.left;
    t.real = realSet(t.run, t.left);
  }
  for (const a of document.getAnimations()) if (a.playState === "paused") a.play();
}
