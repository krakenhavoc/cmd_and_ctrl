import { describe, it, expect } from "vitest";

import { dragVerdict } from "./dragCast";
import { handHasAction, legalActionsOf } from "./legalActions";
import type { CardView, GameView, LegalActionsView, LegalMoveView, PlayerView } from "./protocol";
import { canCastFromHand, LAND_DROPS_SPENT_REASON, landDropsSpent } from "./timing";

// #2203: a land in hand stops looking playable once the turn's land
// plays are used (CR 305.2). Two things draw a hand card (Hand.svelte):
//
//   - the ready ring, from the frame's digest
//     (`legalActionsOf(view).castableFrom(id, "hand")`), and
//   - the dim, `!canCastFromHand(...).legal && !handHasAction(...)`.
//
// With a move list the server answers both: it offers no land move past
// the allowance (internal/legal land_drop_allowance_test.go). With no
// list, canCastFromHand used to answer LEGAL for any card while the
// viewer held priority, a land past its allowance included. It now
// reads the seat's own `land_drops_per_turn` / `lands_played_this_turn`.
// The allowance is the server's (Exploration and one-turn grants are
// summed in); the client only compares the two numbers.

const ME = "p0";
const NIL = "00000000-0000-0000-0000-000000000000";

function zone(kind: string, owner: string, cards: CardView[] = []) {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, idx: number, hand: CardView[], drops?: [number, number]): PlayerView {
  return {
    id,
    name: `seat ${idx}`,
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    ...(drops ? { land_drops_per_turn: drops[0], lands_played_this_turn: drops[1] } : {}),
  };
}

const forest: CardView = {
  instance_id: "c-forest",
  name: "Forest",
  owner: ME,
  controller: ME,
  type_line: "Basic Land — Forest",
};

// A modal DFC whose front is a spell and whose back is a land: its spell
// face stays castable after the land drop is spent.
const mdfc: CardView = {
  instance_id: "c-mdfc",
  name: "Bala Ged Recovery",
  owner: ME,
  controller: ME,
  type_line: "Sorcery",
  layout: "modal_dfc",
  faces: [
    { name: "Bala Ged Recovery", type_line: "Sorcery" },
    { name: "Bala Ged Sanctuary", type_line: "Land" },
  ],
};

const passMove: LegalMoveView = {
  type: "pass_priority",
  player: ME,
  kind: "pass",
  label: "Pass priority",
  source: NIL,
};

function landMove(id: string): LegalMoveView {
  return {
    type: "cast_spell",
    player: ME,
    kind: "land",
    label: "Play " + id,
    source: id,
    params: { instance_id: id, from_zone: "hand" },
  };
}

// frame builds the viewer's frame in its own precombat main phase with
// priority. `offered` is whether the server's list (and digest) offers
// the Forest's land play; `list: false` is a frame with neither.
function frame(opts: {
  drops?: [allowance: number, played: number];
  list: boolean;
  offered?: boolean;
}): GameView {
  const moves = [passMove, ...(opts.offered ? [landMove(forest.instance_id)] : [])];
  const digest: LegalActionsView = {
    pass: true,
    sources: opts.offered
      ? { [forest.instance_id]: { kinds: ["land"], moves: 1, zones: ["hand"] } }
      : undefined,
  };
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, 0, [forest, mdfc], opts.drops), seat("p1", 1, [])],
    battlefield: zone("battlefield", ""),
    stack: zone("stack", ""),
    exile: zone("exile", ""),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "precombat_main",
      step: "precombat_main",
    },
    mulligans_open: false,
    stack_items: [],
    split_second_active: false,
    legal_moves: opts.list ? moves : undefined,
    legal_actions: opts.list ? digest : undefined,
  };
}

// What Hand.svelte draws for the Forest on this frame.
function forestInHand(view: GameView): { ring: boolean; dimmed: boolean; reason?: string } {
  const legal = legalActionsOf(view);
  const leg = canCastFromHand(forest, view, ME);
  return {
    ring: legal.castableFrom(forest.instance_id, "hand"),
    dimmed: !leg.legal && !handHasAction(legal, forest.instance_id),
    reason: leg.reason,
  };
}

const READY = { ring: true, dimmed: false, reason: undefined };
const SPENT_WITH_DIGEST = { ring: false, dimmed: true, reason: LAND_DROPS_SPENT_REASON };
// No digest means no ring at all (ADR 0105 §3); the dim is the answer.
const PLAYABLE_NO_DIGEST = { ring: false, dimmed: false, reason: undefined };
const SPENT_NO_DIGEST = { ring: false, dimmed: true, reason: LAND_DROPS_SPENT_REASON };

describe("landDropsSpent", () => {
  it("compares the seat's own counts", () => {
    expect(landDropsSpent(frame({ drops: [1, 0], list: false }), ME)).toBe(false);
    expect(landDropsSpent(frame({ drops: [1, 1], list: false }), ME)).toBe(true);
    expect(landDropsSpent(frame({ drops: [2, 1], list: false }), ME)).toBe(false);
    expect(landDropsSpent(frame({ drops: [2, 2], list: false }), ME)).toBe(true);
  });

  it("says nothing when the seat carries no counts, or there is no viewer", () => {
    expect(landDropsSpent(frame({ list: false }), ME)).toBe(false);
    expect(landDropsSpent(frame({ drops: [1, 1], list: false }), null)).toBe(false);
    expect(landDropsSpent(null, ME)).toBe(false);
  });
});

describe("a hand land after the turn's land drops (#2203)", () => {
  // Each case: the seat's allowance and lands played, and whether the
  // server's list offers the land (it does exactly while one is left).
  const cases: { name: string; drops: [number, number]; left: boolean }[] = [
    { name: "allowance 1, none played", drops: [1, 0], left: true },
    { name: "allowance 1, one played", drops: [1, 1], left: false },
    { name: "Exploration (allowance 2), one played", drops: [2, 1], left: true },
    { name: "Exploration (allowance 2), two played", drops: [2, 2], left: false },
    { name: "a one-turn grant (allowance 2), one played", drops: [2, 1], left: true },
    { name: "a one-turn grant (allowance 2), two played", drops: [2, 2], left: false },
  ];

  for (const c of cases) {
    it(`${c.name}: with a digest`, () => {
      const view = frame({ drops: c.drops, list: true, offered: c.left });
      expect(forestInHand(view)).toEqual(c.left ? READY : SPENT_WITH_DIGEST);
    });

    it(`${c.name}: without a digest`, () => {
      const view = frame({ drops: c.drops, list: false });
      expect(forestInHand(view)).toEqual(c.left ? PLAYABLE_NO_DIGEST : SPENT_NO_DIGEST);
    });
  }

  it("a frame with no counts stays permissive, as before", () => {
    expect(forestInHand(frame({ list: false }))).toEqual(PLAYABLE_NO_DIGEST);
  });

  it("a modal DFC with a spell face is not refused for the spent land drop", () => {
    const view = frame({ drops: [1, 1], list: false });
    expect(canCastFromHand(mdfc, view, ME)).toEqual({ legal: true });
  });

  it("the land drop is not the reason off the viewer's priority", () => {
    const view = { ...frame({ drops: [1, 1], list: false }) };
    view.turn = { ...view.turn, priority_holder: 1 };
    expect(canCastFromHand(forest, view, ME).reason).toBe("Not your priority");
  });
});

describe("dragging a land past the allowance (#1920, #2203)", () => {
  it("is the red, refused verdict with the land-drop reason, with a digest", () => {
    const view = frame({ drops: [1, 1], list: true, offered: false });
    expect(dragVerdict(forest, canCastFromHand(forest, view, ME), null)).toEqual({
      castable: false,
      reason: LAND_DROPS_SPENT_REASON,
    });
  });

  it("is the red, refused verdict with the land-drop reason, without a digest", () => {
    const view = frame({ drops: [2, 2], list: false });
    expect(dragVerdict(forest, canCastFromHand(forest, view, ME), null)).toEqual({
      castable: false,
      reason: LAND_DROPS_SPENT_REASON,
    });
  });

  it("stays gold while a land drop is left", () => {
    expect(
      dragVerdict(forest, canCastFromHand(forest, frame({ drops: [2, 1], list: false }), ME), null),
    ).toEqual({ castable: true });
    expect(
      dragVerdict(
        forest,
        canCastFromHand(forest, frame({ drops: [1, 0], list: true, offered: true }), ME),
        null,
      ),
    ).toEqual({ castable: true });
  });
});
