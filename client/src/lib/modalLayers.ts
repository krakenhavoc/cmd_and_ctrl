// modalLayers — "is something on screen that owns the keyboard?"
//
// The global shortcut layer must not fire behind a modal. Asking that
// question by checking a list of flags (`bugReportOpen || autoTapCardID
// || choicePromptOpen || …`) would be correct for exactly as long as
// nobody adds a modal, and this client has more than twenty of them
// spread over three directories. So the modals answer instead: each
// one mounts a <ModalLayer /> inside its own `{#if}`, which registers
// a layer for as long as it is on screen and drops it on destroy.
//
// A new modal that forgets the line is the one failure mode, and it
// is a visible one — shortcuts keep firing behind it — rather than the
// invisible one a missed flag produces (a stale `modalOpen` that
// disables every shortcut forever).
//
// Registration is a set of ids rather than a counter so a double
// unregister (Svelte can run a cleanup twice across a keyed re-mount)
// cannot drive the count negative and strand the keymap off.
//
// Deliberately not a "which modal is on top" stack: nothing needs to
// know. Escape routing is already owned by each modal's own handler
// and works correctly today; this store exists purely so the global
// layer can stand down.

import { get, type Readable } from "svelte/store";

import { guardedDerived, guardedWritable } from "./guardedStore";

const layers = guardedWritable<ReadonlySet<number>>(new Set(), "modalLayers");

let nextID = 1;

// modalDepth is how many layers are currently registered. Exposed for
// the dev frame inspector and tests; most consumers want modalOpen.
export const modalDepth: Readable<number> = guardedDerived(layers, (s) => s.size, "modalDepth", 0);

// modalOpen is the reactive predicate the shortcut layer reads.
// `false` is the fail-safe: a stuck-open global layer would swallow
// every keyboard shortcut on the board, which is worse than a modal
// that has to be dismissed with its own close button.
export const modalOpen: Readable<boolean> = guardedDerived(
  layers,
  (s) => s.size > 0,
  "modalOpen",
  false,
);

// pushModalLayer registers a layer and returns its unregister
// function. Idempotent on the way out: calling the returned function
// twice removes the layer once.
export function pushModalLayer(): () => void {
  const id = nextID++;
  layers.update((prev) => {
    const next = new Set(prev);
    next.add(id);
    return next;
  });
  return () => {
    layers.update((prev) => {
      if (!prev.has(id)) return prev;
      const next = new Set(prev);
      next.delete(id);
      return next;
    });
  };
}

// isModalOpen is the non-reactive read, for call sites outside a
// Svelte reactive scope.
export function isModalOpen(): boolean {
  return get(layers).size > 0;
}

// _resetForTests is the vitest teardown hook. Not for prod use.
export function _resetForTests(): void {
  layers.set(new Set());
  nextID = 1;
}
