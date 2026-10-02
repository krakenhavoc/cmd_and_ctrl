// tutorialBus: the tutorial's client event bus (ADR 0076 §2.5, #1077).
//
// Only six of the tutorial's eleven steps can read their completion out of the
// game snapshot. A hover, and a card-local menu opening, never reach the wire,
// so the three components that see them say so here.
//
// THIS IS NOT A TELEMETRY OR ANALYTICS BUS AND MUST NOT GROW INTO ONE. It
// carries exactly the three names in TutorialEvent. Anything that can be read
// from the snapshot is read from the snapshot. A fourth event is a decision for
// the ADR, not a default: do not widen the union to make a PR easier.
//
// It retains nothing. An emit with no subscriber (every game that is not the
// tutorial) calls no one and stores nothing, and a subscriber hears only what is
// emitted after it subscribed, so a step never completes on a stale event.

export type TutorialEvent = "ability-menu-opened" | "hand-hovered" | "pile-hovered";

type Handler = (name: TutorialEvent) => void;

const handlers = new Set<Handler>();

/** Announce that `name` just happened. Free when nobody is listening. */
export function emit(name: TutorialEvent): void {
  if (handlers.size === 0) return;
  // Copy first: a handler may unsubscribe itself (a step that just completed).
  for (const h of [...handlers]) h(name);
}

/** Listen for events from now on. Returns the unsubscribe function. */
export function subscribe(handler: Handler): () => void {
  handlers.add(handler);
  return () => {
    handlers.delete(handler);
  };
}

export const tutorialBus = { emit, subscribe };
