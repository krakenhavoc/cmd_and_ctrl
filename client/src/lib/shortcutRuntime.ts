// shortcutRuntime — the seam between the one global key listener and
// whatever route can actually service an action.
//
// There is exactly ONE keydown listener for shortcuts, mounted at the
// app shell (ShortcutLayer.svelte). That is the whole point: a second
// global listener is how two features end up both claiming a key and
// neither one winning reliably. But most of the keymap is about a
// game table, and the shell has no snapshot, no GameClient and no
// seat. So the game route publishes two things here while it is
// mounted — what the actions DO, and what the seat's situation IS —
// and takes them back down on destroy.
//
// When nothing is registered the context is `idleContext()`, every
// game-scoped action reports "only at a game table", and the global
// rows (settings, help, mute) keep working from the lobby.
//
// Registration is token-guarded. Svelte can mount the next route
// before the previous one's cleanup runs (the keyed `{#key gameID}`
// remount in App.svelte does exactly this), and an unguarded
// unregister would then tear down the NEW route's handlers.

import { get, type Readable, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { idleContext, type ShortcutContext, type ShortcutID } from "./shortcuts";

export type ShortcutHandlers = Partial<Record<ShortcutID, () => void>>;

const handlersStore = guardedWritable<ShortcutHandlers>({}, "shortcutHandlers");
const contextStore = guardedWritable<ShortcutContext>(idleContext(), "shortcutContext");

export const shortcutHandlers: Readable<ShortcutHandlers> = {
  subscribe: handlersStore.subscribe,
};
export const shortcutContext: Readable<ShortcutContext> = {
  subscribe: contextStore.subscribe,
};

let token = 0;

// registerShortcutHandlers publishes a route's action handlers and
// returns the teardown. Call from onMount; the returned function is
// safe to hand straight back to Svelte.
export function registerShortcutHandlers(h: ShortcutHandlers): () => void {
  const mine = ++token;
  handlersStore.set(h);
  return () => {
    if (token !== mine) return; // a newer route already took over
    handlersStore.set({});
    contextStore.set(idleContext());
  };
}

// setShortcutContext publishes the seat's current situation. Cheap
// enough to call from an $effect on every snapshot — the payload is a
// dozen booleans and the store only notifies when the object changes
// identity, which it does once per frame at most.
export function setShortcutContext(ctx: ShortcutContext): void {
  contextStore.set(ctx);
}

// currentContext / currentHandlers are the non-reactive reads the
// dispatcher uses inside its event handler.
export function currentContext(): ShortcutContext {
  return get(contextStore);
}

export function currentHandlers(): ShortcutHandlers {
  return get(handlersStore);
}

// shortcutsHelpOpen drives the `?` overlay. Lives here rather than in
// the overlay component so any surface can open it — the Settings
// panel's "show the cheat sheet" link does.
export const shortcutsHelpOpen: Writable<boolean> = guardedWritable(false, "shortcutsHelpOpen");

export function openShortcutsHelp(): void {
  shortcutsHelpOpen.set(true);
}

export function closeShortcutsHelp(): void {
  shortcutsHelpOpen.set(false);
}

// _resetForTests is the vitest teardown hook. Not for prod use.
export function _resetForTests(): void {
  token = 0;
  handlersStore.set({});
  contextStore.set(idleContext());
  shortcutsHelpOpen.set(false);
}
