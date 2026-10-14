// #3381 — a spell travelling to the stack, as pure functions: priming,
// the origin read off the cast entry, a spell gone by the frame that
// logs it, the burst cap, and where each flight starts.

import { describe, expect, it } from "vitest";

import {
  CAST_FLIGHT_MAX,
  castOriginSelectors,
  emptyCastTracker,
  stackItemSelector,
  trackCasts,
  type CastTracker,
} from "./castFlight";
import type { CardView, GameView, LogEvent } from "./protocol";

const card = (id: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: id,
    owner: "b",
    controller: "b",
    scryfall_id: `sf-${id}`,
    ...extra,
  }) as CardView;

function view(log: LogEvent[], stack: CardView[] = []): GameView {
  return {
    stack: { kind: "stack", count: stack.length, cards: stack },
    seats: [{ id: "a" }, { id: "b" }],
    log,
  } as unknown as GameView;
}

const step = (seq: number): LogEvent => ({ seq, kind: "step", seat: 0, text: "" });
const cast = (seq: number, card_id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
  seq,
  kind: "cast",
  seat: 1,
  card_id,
  text: "",
  ...extra,
});

function primed(): CastTracker {
  const res = trackCasts(emptyCastTracker(), view([step(1)]));
  expect(res.flights).toEqual([]);
  return res.tracker;
}

describe("trackCasts", () => {
  it("flies nothing on the first frame, even with a cast in the log", () => {
    const res = trackCasts(emptyCastTracker(), view([cast(1, "bolt")], [card("bolt")]));
    expect(res.flights).toEqual([]);
  });

  it("flies a new cast from the caster's hand, with its face", () => {
    const res = trackCasts(primed(), view([step(1), cast(2, "bolt")], [card("bolt")]));
    expect(res.flights).toEqual([
      {
        itemID: "bolt",
        seat: "b",
        from: "hand",
        image: "/cards/sf-bolt/image?size=normal",
        seq: 2,
      },
    ]);
  });

  it("reads a non-hand origin off old_zone", () => {
    const t = primed();
    const v = view(
      [
        cast(2, "fb", { old_zone: "graveyard" }),
        cast(3, "warp", { old_zone: "exile" }),
        cast(4, "cmdr", { old_zone: "command" }),
        cast(5, "odd", { old_zone: "sideboard" }),
      ],
      [card("fb"), card("warp"), card("cmdr"), card("odd")],
    );
    // Four casts, within CAST_FLIGHT_MAX: all fly.
    expect(trackCasts(t, v).flights.map((f) => f.from)).toEqual([
      "graveyard",
      "exile",
      "command",
      "other",
    ]);
  });

  it("draws a card back for a spell this viewer may not see", () => {
    const res = trackCasts(
      primed(),
      view([cast(2, "morph")], [card("morph", { scryfall_id: undefined })]),
    );
    expect(res.flights[0].image).toBe("/card-back.jpg");
  });

  it("skips a spell already off the stack, and other entries", () => {
    const res = trackCasts(
      primed(),
      view([cast(2, "gone"), { seq: 3, kind: "resolve", seat: 1, card_id: "x", text: "" }]),
    );
    expect(res.flights).toEqual([]);
  });

  it("is quiet on an undo", () => {
    const after = trackCasts(primed(), view([step(5)]));
    const undone = trackCasts(after.tracker, view([cast(3, "bolt")], [card("bolt")]));
    expect(undone.flights).toEqual([]);
  });

  it("keeps only the newest casts of a burst", () => {
    const ids = Array.from({ length: CAST_FLIGHT_MAX + 2 }, (_, i) => `s${i}`);
    const res = trackCasts(
      primed(),
      view(
        ids.map((id, i) => cast(i + 2, id)),
        ids.map((id) => card(id)),
      ),
    );
    expect(res.flights.map((f) => f.itemID)).toEqual(ids.slice(2));
  });
});

describe("castOriginSelectors", () => {
  it("starts at the hand, a pile, or the seat", () => {
    expect(castOriginSelectors({ seat: "b", from: "hand" })).toEqual([
      '[data-hand-owner="b"]',
      '[data-seat-id="b"]',
    ]);
    expect(castOriginSelectors({ seat: "b", from: "graveyard" })).toEqual([
      '[data-pile="graveyard"][data-pile-owner="b"]',
      '[data-seat-id="b"]',
    ]);
    expect(castOriginSelectors({ seat: "b", from: "command" })).toEqual(['[data-seat-id="b"]']);
  });

  it("names the stack item", () => {
    expect(stackItemSelector("bolt")).toBe('[data-stack-item-id="bolt"]');
  });
});
