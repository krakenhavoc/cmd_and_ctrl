import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { buildMovePayload, canManageZone, cardsForZone } from "./zoneBrowser.logic";

// Tiny builders keep each test's setup near where the expectations
// live. We only populate the fields the logic under test reads so
// the tests don't break when unrelated protocol fields move around.

function card(
  instance_id: string,
  owner: string,
  controller?: string,
  extra: Partial<CardView> = {},
): CardView {
  return {
    instance_id,
    name: instance_id,
    owner,
    controller: controller ?? owner,
    ...extra,
  };
}

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, graveyard: CardView[], command: CardView[]): PlayerView {
  return {
    id,
    name,
    seat: 0,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, graveyard),
    command: zone("command", id, command),
    commander_damage: {},
    life_history: [],
  };
}

function view(seats: PlayerView[], exile: CardView[] = [], stack: CardView[] = []): GameView {
  return {
    id: "g1",
    state: "active",
    seats,
    battlefield: zone("battlefield", undefined, []),
    stack: zone("stack", undefined, stack),
    exile: zone("exile", undefined, exile),
    turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "" },
    mulligans_open: false,
  };
}

describe("cardsForZone", () => {
  it("returns the owner's graveyard slice", () => {
    const a = seat("a", "Alice", [card("a1", "a"), card("a2", "a")], []);
    const b = seat("b", "Bob", [card("b1", "b")], []);
    const v = view([a, b]);
    expect(cardsForZone(v, "graveyard", "a").map((c) => c.instance_id)).toEqual(["a1", "a2"]);
    expect(cardsForZone(v, "graveyard", "b").map((c) => c.instance_id)).toEqual(["b1"]);
  });

  it("filters the shared exile zone by owner", () => {
    const a = seat("a", "Alice", [], []);
    const b = seat("b", "Bob", [], []);
    const v = view([a, b], [card("e1", "a"), card("e2", "b"), card("e3", "a")]);
    expect(cardsForZone(v, "exile", "a").map((c) => c.instance_id)).toEqual(["e1", "e3"]);
    expect(cardsForZone(v, "exile", "b").map((c) => c.instance_id)).toEqual(["e2"]);
  });

  it("returns the command zone cards", () => {
    const a = seat("a", "Alice", [], [card("cmd_a", "a")]);
    const v = view([a]);
    expect(cardsForZone(v, "command", "a").map((c) => c.instance_id)).toEqual(["cmd_a"]);
  });

  it("returns the full stack regardless of ownerID", () => {
    const a = seat("a", "Alice", [], []);
    const stackCards = [card("s1", "a"), card("s2", "b")];
    const v = view([a], [], stackCards);
    expect(cardsForZone(v, "stack", "a").map((c) => c.instance_id)).toEqual(["s1", "s2"]);
    expect(cardsForZone(v, "stack", "b").map((c) => c.instance_id)).toEqual(["s1", "s2"]);
  });

  it("returns [] when the owner seat is missing", () => {
    const a = seat("a", "Alice", [card("x", "a")], []);
    const v = view([a]);
    expect(cardsForZone(v, "graveyard", "missing")).toEqual([]);
  });
});

describe("canManageZone", () => {
  it("allows the owner to manage their own graveyard", () => {
    expect(canManageZone("graveyard", "a", "a")).toBe(true);
  });
  it("denies a non-owner", () => {
    expect(canManageZone("graveyard", "b", "a")).toBe(false);
  });
  it("denies a spectator (null viewer)", () => {
    expect(canManageZone("graveyard", null, "a")).toBe(false);
  });
  it("never allows stack management from the browser", () => {
    expect(canManageZone("stack", "a", "a")).toBe(false);
  });
  it("applies to exile + command the same way", () => {
    expect(canManageZone("exile", "a", "a")).toBe(true);
    expect(canManageZone("exile", "b", "a")).toBe(false);
    expect(canManageZone("command", "a", "a")).toBe(true);
    expect(canManageZone("command", "b", "a")).toBe(false);
  });
});

describe("buildMovePayload", () => {
  it("builds a hand move with src + dst owner stamped to the viewer", () => {
    const c = card("x", "a");
    const p = buildMovePayload("graveyard", "a", "a", c, "hand");
    expect(p).toEqual({
      src: { kind: "graveyard", owner: "a" },
      dst: { kind: "hand", owner: "a" },
      instance_id: "x",
    });
  });

  it("omits the dst owner for battlefield (shared zone)", () => {
    const c = card("x", "a");
    const p = buildMovePayload("graveyard", "a", "a", c, "battlefield");
    expect(p).toEqual({
      src: { kind: "graveyard", owner: "a" },
      dst: { kind: "battlefield" },
      instance_id: "x",
    });
  });

  it("stamps the viewer as the dst owner for library moves", () => {
    const c = card("x", "a");
    const p = buildMovePayload("exile", "a", "a", c, "library");
    expect(p).toEqual({
      src: { kind: "exile", owner: "a" },
      dst: { kind: "library", owner: "a" },
      instance_id: "x",
    });
  });

  it("returns null when the viewer is not the owner", () => {
    const c = card("x", "a");
    expect(buildMovePayload("graveyard", "a", "b", c, "hand")).toBeNull();
  });

  it("returns null when the card's controller differs from the viewer", () => {
    // An exiled card the viewer owns but which is controlled by
    // another player (e.g. Thassa's Oracle-style control-change)
    // must not be movable by the viewer — defense in depth.
    const c = card("x", "a", "b");
    expect(buildMovePayload("exile", "a", "a", c, "hand")).toBeNull();
  });

  it("never builds a payload for the stack", () => {
    const c = card("x", "a");
    expect(buildMovePayload("stack", "a", "a", c, "hand")).toBeNull();
  });

  it("returns null when viewerID is empty (spectator sentinel)", () => {
    const c = card("x", "a");
    // buildMovePayload expects a non-empty viewerID — the component
    // guards before calling — but belt-and-braces: pass "" which
    // canManageZone treats as "no viewer" via the !viewerID check.
    expect(buildMovePayload("graveyard", "a", "", c, "hand")).toBeNull();
  });
});
