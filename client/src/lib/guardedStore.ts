// guardedStore.ts — `writable` and `derived` whose callbacks cannot
// escape into svelte/store's shared queue (#266, #720).
//
// THE AMPLIFIER
//
// svelte/store keeps ONE module-global `subscriber_queue`, shared by
// every store in the application. `set()` drains it with a bare loop
// and clears it only after the loop finishes:
//
//     for (let i = 0; i < subscriber_queue.length; i += 2) {
//       subscriber_queue[i][0](subscriber_queue[i + 1]);
//     }
//     subscriber_queue.length = 0;
//
// A callback that throws skips both the remaining callbacks AND the
// reset, so the queue stays permanently non-empty. From then on every
// `set()` on EVERY store in the app — snapshot, status, log, settings,
// targeting, the router — sees a non-empty queue, enqueues its value
// and returns without flushing. The socket stays green, handleMessage
// keeps running, seq keeps advancing, and nothing ever reaches the DOM
// again. That is #266's reported "state freeze", and it lasts for the
// life of the page.
//
// #284 guarded GameClient's own stores, which is why this module
// exists rather than that one: the queue is GLOBAL, so guarding one
// module's stores protects nothing. A throw from a subscriber on
// `settings`, `targeting`, `manualStops` or a per-card meta store
// stalls GameClient's guarded stores just as thoroughly. Every
// module-level store has to be guarded, or none of them are.
//
// WHAT CAN ACTUALLY THROW (the skeptic review on #720)
//
// Not renders. In Svelte 5 a component's `$store` subscriber only
// calls `set(source)` and schedules a microtask batch, so a throw in
// render, `$derived` or `$effect` unwinds on the batch, not inside the
// queue drain. What poisons the queue is a throw inside a SUBSCRIBE
// CALLBACK or a store's own logic:
//
//   - an explicit `.subscribe(...)` — `settings.ts` persists to
//     localStorage in one, `HoverZoomOverlay.svelte` reads card meta in
//     another
//   - a `derived()` callback, which runs inside its parent's subscriber
//   - unguarded side effects in either: a `localStorage` write throws
//     on a full quota or in Safari private browsing, and `JSON.stringify`
//     throws on a cycle
//
// So the guard is aimed there, and `<svelte:boundary>` in Game.svelte
// covers the render side.
//
// WHAT THIS DOES NOT DO
//
// It does not make the throwing code correct. One component's update is
// still lost. The recorded `clientErrors` entry is the point: the next
// bug report names the thing that threw instead of arriving empty.

import { derived, writable, type Readable, type Writable } from "svelte/store";

import { recordClientError } from "./clientErrors";

// describeThrown renders whatever a callback threw. Callbacks can throw
// anything at all, and this runs inside the failure path, so it must
// not be able to throw itself.
export function describeThrown(err: unknown): string {
  if (err instanceof Error) return `${err.name}: ${err.message}`;
  if (typeof err === "string") return err;
  try {
    return JSON.stringify(err) ?? String(err);
  } catch {
    return "unknown error";
  }
}

/**
 * guardSubscribe wraps a store's `subscribe` so a throwing subscriber
 * is contained and recorded instead of reaching svelte/store's drain
 * loop. Exported for a store this module does not build — a
 * hand-rolled `{ subscribe }` object, or one from a library.
 */
export function guardSubscribe<T>(
  subscribe: Readable<T>["subscribe"],
  label: string,
): Readable<T>["subscribe"] {
  return (run, invalidate) =>
    subscribe((value) => {
      try {
        run(value);
      } catch (err) {
        recordClientError(`subscriber threw on ${label}: ${describeThrown(err)}`);
      }
    }, invalidate);
}

/**
 * guardedWritable is a `writable` whose subscribers cannot escape.
 * `label` names the store in the recorded error, so use the exported
 * name.
 */
export function guardedWritable<T>(initial: T, label: string): Writable<T> {
  const inner = writable(initial);
  return {
    set: inner.set,
    update: inner.update,
    subscribe: guardSubscribe(inner.subscribe, label),
  };
}

/**
 * guardedDerived is `derived` with both halves guarded: the mapping
 * callback, which runs inside the PARENT store's subscriber and is the
 * likelier of the two to throw, and the derived store's own
 * subscribers.
 *
 * `fallback` is the value seen while the mapping is failing — the
 * derived store holds its last good value, or `fallback` if it has
 * never produced one. There is deliberately no way to express "no
 * value": a derived store that cannot compute still has to answer, and
 * a caller has to be handed something. Pick the answer that fails
 * safe (`false` for a feature flag, `0` for a depth), not the one that
 * is convenient.
 */
export function guardedDerived<S, T>(
  store: Readable<S>,
  fn: (value: S) => T,
  label: string,
  fallback: T,
): Readable<T> {
  let last = fallback;
  // svelte's own `derived` is kept underneath for its laziness: it
  // subscribes to the parent on the first subscriber and drops the
  // subscription on the last. Only the callback is wrapped.
  const inner = derived(store, (value) => {
    try {
      last = fn(value);
    } catch (err) {
      recordClientError(`derived ${label} threw: ${describeThrown(err)}`);
    }
    return last;
  });
  return { subscribe: guardSubscribe(inner.subscribe, label) };
}
