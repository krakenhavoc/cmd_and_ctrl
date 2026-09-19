// @vitest-environment jsdom
//
// #740 — the root cause, in isolation.
//
// `metaFor` used to take a cached entry and re-`set()` the store it
// was about to return. Its one caller reached it from a `$derived`
// (`HoverZoomOverlay.svelte`), so that `set()` ran the store's
// subscribers inside the derived's reaction — and the subscriber
// assigns to a `$state`. Svelte answers a `$state` write from inside a
// running derived by throwing `state_unsafe_mutation`, which is the
// error the frozen board reported.
//
// The trigger is narrow, which is why the bug was seen once in four
// tries and never by hand again: the derived has to re-run AND return
// the store it is already subscribed to. That means the hovered card
// changing to a DIFFERENT card with the SAME `scryfall_id`, with no
// null in between — a second copy of the same printing, or the same
// printing hovered again after the first Card unmounted without a
// `pointerleave` (a zone browser closing over it, a permanent leaving
// the battlefield under the cursor).
//
// Pre-#720 that throw escaped into svelte/store's shared
// `subscriber_queue` and stalled every store in the app. The guards
// contain it now, so this test asserts on the recorded error rather
// than on a thrown one: the freeze is fixed, and this is the throw
// that caused it not happening at all.

import { describe, it, expect, beforeEach, afterEach } from "vitest";

import HoverZoomOverlay from "./components/board/HoverZoomOverlay.svelte";
import { hoveredCard } from "./cardTypes";
import { recentClientErrors, resetClientErrors } from "./clientErrors";
import type { CardView, GameView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

// The metadata cache is module-global and deliberately never evicted,
// so each test names its own printing rather than fighting over one.
let printings = 0;
const nextPrinting = (): string => `sf-grizzly-${++printings}`;

const view: GameView = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: {
    number: 4,
    active_seat: 0,
    priority_holder: 0,
    phase: "combat",
    step: "declare_attackers",
  },
} as unknown as GameView;

const copy = (instanceID: string, scryfallID: string): CardView =>
  ({
    instance_id: instanceID,
    name: "Grizzly Bears",
    owner: "me",
    controller: "me",
    known_by_you: true,
    scryfall_id: scryfallID,
    type_line: "Creature — Bear",
    power: 2,
    toughness: 2,
  }) as unknown as CardView;

let realFetch: unknown;
let cardFetches = 0;

beforeEach(() => {
  resetClientErrors();
  hoveredCard.set(null);
  cardFetches = 0;
  const g = globalThis as Record<string, unknown>;
  realFetch = g.fetch;
  g.fetch = (url: unknown) => {
    cardFetches += 1;
    return Promise.resolve({
      ok: true,
      json: () =>
        Promise.resolve({
          id: String(url).split("/").pop(),
          name: "Grizzly Bears",
          type_line: "Creature — Bear",
          oracle_text: "It is a bear.",
        }),
    } as unknown as Response);
  };
});

afterEach(() => {
  cleanup();
  (globalThis as Record<string, unknown>).fetch = realFetch;
});

// settle drains the metadata fetch and flushes the effects it wakes.
async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    flushSync();
  }
}

const unsafeMutations = (): string[] =>
  recentClientErrors()
    .map((e) => e.text)
    .filter((t) => t.includes("state_unsafe_mutation"));

describe("the hover panel's metadata subscription", () => {
  it("does not write state while a derived is running when the printing repeats", async () => {
    const printing = nextPrinting();
    const { container } = render(HoverZoomOverlay as never, { view } as never);

    hoveredCard.set(copy("bear-1", printing));
    await settle();
    expect(container.textContent, "the first hover should have loaded the oracle text").toContain(
      "It is a bear.",
    );

    // The second copy. Same printing, different permanent, and no
    // null in between — the overlay's effect is still subscribed to
    // this exact store when the derived re-runs.
    hoveredCard.set(copy("bear-2", printing));
    await settle();

    expect(unsafeMutations(), "a $state write from inside a $derived").toEqual([]);
    expect(container.textContent, "the panel should still be showing the card").toContain(
      "It is a bear.",
    );
  });

  it("fetches one printing once however many copies are hovered", async () => {
    const printing = nextPrinting();
    render(HoverZoomOverlay as never, { view } as never);

    hoveredCard.set(copy("bear-1", printing));
    await settle();
    hoveredCard.set(copy("bear-2", printing));
    await settle();
    hoveredCard.set(copy("bear-3", printing));
    await settle();

    // The cache is the point of the module. One printing, one request
    // — and the caching branch is the one that used to write.
    expect(cardFetches).toBe(1);
    expect(unsafeMutations()).toEqual([]);
  });

  it("still swaps the panel over when the printing changes", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);

    hoveredCard.set(copy("bear-1", nextPrinting()));
    await settle();
    hoveredCard.set(copy("bolt-1", nextPrinting()));
    await settle();

    expect(cardFetches, "a new printing is a new request").toBe(2);
    expect(container.textContent).toContain("It is a bear.");
    expect(unsafeMutations()).toEqual([]);
  });

  it("clears the panel when the cursor leaves the table", async () => {
    const { container } = render(HoverZoomOverlay as never, { view } as never);

    hoveredCard.set(copy("bear-1", nextPrinting()));
    await settle();
    hoveredCard.set(null);
    await settle();

    expect(container.textContent?.trim() ?? "").toBe("");
    expect(unsafeMutations()).toEqual([]);
  });
});
