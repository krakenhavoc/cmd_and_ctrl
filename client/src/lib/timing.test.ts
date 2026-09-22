import { describe, it, expect } from "vitest";

import {
  canActivateLoyalty,
  canActivateSorcerySpeedAbility,
  canCastFromHand,
  hasNonPassMove,
  hasPassMove,
  hasPriority,
  isActivePlayer,
  isMainPhase,
  movesFor,
  stackEmpty,
} from "./timing";
import type {
  CardView,
  GameView,
  LegalMoveView,
  PlayerView,
  TurnView,
  ZoneView,
  StackItemView,
} from "./protocol";

// emptyZone returns a zone with no cards. The kind / owner fields
// don't matter for the timing predicates — they only read counts.
function emptyZone(kind: string, owner = ""): ZoneView {
  return { kind, owner, count: 0, cards: [] };
}

function turn(opts: Partial<TurnView> = {}): TurnView {
  return {
    number: 1,
    active_seat: 0,
    priority_holder: 0,
    phase: "precombat_main",
    step: "precombat_main",
    ...opts,
  };
}

function seat(id: string, idx: number, name = ""): PlayerView {
  return {
    id,
    name: name || `seat ${idx}`,
    seat: idx,
    life: 40,
    library: emptyZone("library", id),
    hand: emptyZone("hand", id),
    graveyard: emptyZone("graveyard", id),
    command: emptyZone("command", id),
    commander_damage: {},
    life_history: [],
  };
}

interface SnapOpts {
  step?: string;
  activeSeat?: number;
  priorityHolder?: number;
  splitSecond?: boolean;
  stackItems?: StackItemView[];
  battlefield?: CardView[];
  // S31: the seat's enumerated move list. UNDEFINED and [] are
  // deliberately different states — undefined is "the server told us
  // nothing", [] is "the server enumerated and this seat has nothing
  // to do" — and most of the suite below turns on the distinction.
  moves?: LegalMoveView[];
}

function snap(o: SnapOpts = {}): GameView {
  const seats = [seat("p0", 0), seat("p1", 1)];
  return {
    id: "g",
    state: "active",
    seats,
    battlefield: { ...emptyZone("battlefield"), cards: o.battlefield ?? [] },
    stack: emptyZone("stack"),
    exile: emptyZone("exile"),
    turn: turn({
      step: o.step ?? "precombat_main",
      active_seat: o.activeSeat ?? 0,
      priority_holder: o.priorityHolder ?? 0,
    }),
    mulligans_open: false,
    stack_items: o.stackItems ?? [],
    split_second_active: o.splitSecond ?? false,
    legal_moves: o.moves,
  };
}

function card(name: string, type: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "c-" + name,
    name,
    owner: "p0",
    controller: "p0",
    type_line: type,
    ...extras,
  };
}

// castMove is the shape the server ships for "you may play this".
function castMove(
  instanceID: string,
  kind: LegalMoveView["kind"] = "cast",
  player = "p0",
): LegalMoveView {
  return {
    type: "cast_spell",
    player,
    kind,
    label: "Cast " + instanceID,
    source: instanceID,
    params: { instance_id: instanceID, from_zone: "hand" },
  };
}

const passMove: LegalMoveView = {
  type: "pass_priority",
  player: "p0",
  kind: "pass",
  label: "Pass priority",
  source: "00000000-0000-0000-0000-000000000000",
};

describe("priority + step predicates", () => {
  it("hasPriority is false during Untap / Cleanup (-1 sentinel)", () => {
    const s = snap({ step: "untap", priorityHolder: -1 });
    expect(hasPriority(s, "p0")).toBe(false);
  });

  it("hasPriority is true when priority_holder seat matches viewer", () => {
    const s = snap();
    expect(hasPriority(s, "p0")).toBe(true);
    expect(hasPriority(s, "p1")).toBe(false);
  });

  it("isActivePlayer reads active_seat", () => {
    const s = snap({ activeSeat: 1, priorityHolder: 0 });
    expect(isActivePlayer(s, "p1")).toBe(true);
    expect(isActivePlayer(s, "p0")).toBe(false);
  });

  it("isMainPhase covers both pre/post combat mains", () => {
    expect(isMainPhase(snap({ step: "precombat_main" }))).toBe(true);
    expect(isMainPhase(snap({ step: "postcombat_main" }))).toBe(true);
    expect(isMainPhase(snap({ step: "draw" }))).toBe(false);
    expect(isMainPhase(snap({ step: "begin_combat" }))).toBe(false);
  });

  it("stackEmpty reads both stack_items + stack.cards", () => {
    expect(stackEmpty(snap())).toBe(true);
    expect(
      stackEmpty(
        snap({
          stackItems: [
            { id: "x", kind: "spell", controller: "p0", owner: "p0", source_card_id: "x" },
          ],
        }),
      ),
    ).toBe(false);
  });
});

// The move-list lookup itself. Its one interesting property is that
// "absent" and "empty" are different answers, because a client that
// confuses them greys the player's whole hand on every frame where
// they hold no priority.
describe("movesFor / hasNonPassMove", () => {
  it("returns undefined — not [] — when the frame carries no move list", () => {
    expect(movesFor(snap(), "c-Bolt")).toBeUndefined();
    expect(hasNonPassMove(snap())).toBeUndefined();
    expect(movesFor(null, "c-Bolt")).toBeUndefined();
  });

  it("returns [] when the server enumerated and this card has no move", () => {
    const s = snap({ moves: [passMove] });
    expect(movesFor(s, "c-Bolt")).toEqual([]);
    expect(hasNonPassMove(s)).toBe(false);
  });

  it("filters by source and, optionally, by kind", () => {
    const s = snap({ moves: [passMove, castMove("c-Bolt"), castMove("c-Forest", "land")] });
    expect(movesFor(s, "c-Bolt")).toHaveLength(1);
    expect(movesFor(s, "c-Forest", ["cast"])).toHaveLength(0);
    expect(movesFor(s, "c-Forest", ["cast", "land"])).toHaveLength(1);
    expect(hasNonPassMove(s)).toBe(true);
  });

  it("a pass-only list is not a response", () => {
    expect(hasNonPassMove(snap({ moves: [passMove] }))).toBe(false);
  });

  it("hasPassMove answers from the list, and says nothing when there is none", () => {
    // The keyboard layer's gate for the pass-priority key (ADR 0047).
    expect(hasPassMove(snap({ moves: [passMove] }))).toBe(true);
    expect(hasPassMove(snap({ moves: [castMove("c-Bolt")] }))).toBe(false);
    expect(hasPassMove(snap({ moves: [] }))).toBe(false);
    // No list at all is "no information", not "you can't pass".
    expect(hasPassMove(snap())).toBeUndefined();
    expect(hasPassMove(null)).toBeUndefined();
  });
});

// The verdict. Every case here is the server's answer, which is the
// whole point of S31 sub-PR 2: the client no longer has an opinion
// about sorcery speed, the land drop, flash, split second or faces.
describe("canCastFromHand — the verdict is the server's move list", () => {
  it("a card with a cast move is castable", () => {
    const bolt = card("Bolt", "Instant");
    const s = snap({ moves: [passMove, castMove(bolt.instance_id)] });
    expect(canCastFromHand(bolt, s, "p0").legal).toBe(true);
  });

  it("a card with a land move is playable", () => {
    const forest = card("Forest", "Basic Land — Forest");
    const s = snap({ moves: [passMove, castMove(forest.instance_id, "land")] });
    expect(canCastFromHand(forest, s, "p0").legal).toBe(true);
  });

  it("a card the server did not enumerate is not castable, whatever its type line says", () => {
    const bolt = card("Bolt", "Instant");
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(bolt, s, "p0").legal).toBe(false);
  });

  it("the spent land drop greys the second land — CR 305.2, which the old client never checked", () => {
    const forest = card("Forest", "Basic Land — Forest");
    // Own main phase, empty stack, priority held: every gate the old
    // predicate knew about says yes. The server says no because the
    // drop is spent, and the server is the one that counts.
    const s = snap({ moves: [passMove] });
    const got = canCastFromHand(forest, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBeTruthy();
  });

  it("an unaffordable spell greys — mana the old client could not see", () => {
    const titan = card("Worldspine Wurm", "Creature — Wurm", { mana_cost: "{11}" });
    const s = snap({ moves: [passMove, castMove("c-Something-Else")] });
    expect(canCastFromHand(titan, s, "p0").legal).toBe(false);
  });

  it("a modal DFC is offered exactly when the server enumerated a face for it", () => {
    // The face-by-face type-line walk is gone from the client: the
    // enumerator expands CastableFaces server-side and the move it
    // emits carries the face index in its params.
    const mdfc = card("Malakir Rebirth", "Instant", {
      layout: "modal_dfc",
      faces: [
        { name: "Malakir Rebirth", type_line: "Instant" },
        { name: "Malakir Mire", type_line: "Land" },
      ],
    });
    const offered = snap({
      step: "declare_attackers",
      moves: [passMove, castMove(mdfc.instance_id)],
    });
    expect(canCastFromHand(mdfc, offered, "p0").legal).toBe(true);
    const notOffered = snap({ step: "declare_attackers", moves: [passMove] });
    expect(canCastFromHand(mdfc, notOffered, "p0").legal).toBe(false);
  });
});

// The reasons. These are tooltip copy, derived from fields the server
// stamps; none of them can flip a verdict.
describe("canCastFromHand — denial reasons", () => {
  it("spectator (no viewerID) cannot cast", () => {
    expect(canCastFromHand(card("Bolt", "Instant"), snap(), null).reason).toBe(
      "Spectator can't cast",
    );
  });

  it("viewer without priority", () => {
    const s = snap({ priorityHolder: 1, moves: [] });
    expect(canCastFromHand(card("Bolt", "Instant"), s, "p0").reason).toBe("Not your priority");
  });

  it("split second", () => {
    const s = snap({ splitSecond: true, moves: [passMove] });
    expect(canCastFromHand(card("Bolt", "Instant"), s, "p0").reason).toBe(
      "Split second on the stack",
    );
  });

  it("targeted spell with no legal target", () => {
    const blade = card("Doom Blade", "Instant", { legal_targets: { cards: [] } });
    const s = snap({ moves: [passMove] });
    const got = canCastFromHand(blade, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBe("No legal target");
  });

  it("targeted spell needing two targets says so", () => {
    const c = card("Twin Bolt", "Instant", { legal_targets: { cards: ["a"], min: 2 } });
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(c, s, "p0").reason).toBe("Needs 2 legal targets");
  });

  it("an alternative cost that drops the target clause rescues the card", () => {
    // Overloaded Cyclonic Rift has no target clause at all, so the
    // printed clause having nothing to point at is not a denial.
    const rift = card("Cyclonic Rift", "Instant", {
      legal_targets: { cards: [] },
      alternative_costs: [{ key: "overload", label: "Overload {6}{U}", mana_cost: "{6}{U}" }],
    });
    const s = snap({ moves: [passMove, castMove(rift.instance_id)] });
    expect(canCastFromHand(rift, s, "p0").legal).toBe(true);
  });

  it("modal spell with no castable mode", () => {
    const stuck = card("Charm", "Instant", {
      modes: {
        prompt: "Choose one",
        min: 1,
        max: 1,
        options: [
          { label: "A", target_mode: "permanent", legal_targets: { cards: [] } },
          { label: "B", target_mode: "player", legal_targets: { players: [] } },
        ],
      },
    });
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(stuck, s, "p0").reason).toBe("No castable mode");
  });

  it("additional cost: nothing to sacrifice", () => {
    const rites = card("Village Rites", "Instant", {
      additional_cost: { sacrifice_options: { cards: [] } },
    });
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(rites, s, "p0").reason).toBe("Nothing to sacrifice");
  });

  it("additional cost: fewer permanents than a sacrifice-N clause needs (#747)", () => {
    const rites = card("Two-Creature Rites", "Sorcery", {
      additional_cost: { sacrifice_options: { cards: ["c1"], min: 2, max: 2 } },
    });
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(rites, s, "p0").reason).toBe("Needs 2 permanents to sacrifice");
  });

  it("additional cost: no card to discard", () => {
    const c = card("Thrill", "Sorcery", { additional_cost: { discard_cards: 1 } });
    const s = snap({ moves: [passMove] });
    // The viewer's hand holds only the spell itself, which doesn't
    // count — it's on the stack by the time costs are paid.
    s.seats[0].hand = { ...emptyZone("hand", "p0"), count: 1, cards: [c] };
    expect(canCastFromHand(c, s, "p0").reason).toBe("No card to discard");
  });

  // #1185: `castIsForbidden` (targeting.ts) had no caller — a card
  // refused by its own printed `cant_cast` clause fell through to the
  // generic fallbacks below instead of naming its own reason.
  it("a cant_cast clause denies with the printed clause itself", () => {
    const c = card("Bolt", "Instant", {
      cant_cast: "Each player can't cast more than one spell each turn",
    });
    const s = snap({ moves: [passMove] });
    const got = canCastFromHand(c, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBe("Each player can't cast more than one spell each turn");
  });

  it("cant_cast is checked before target / mode / cost, since it refuses outright", () => {
    const c = card("Doom Blade", "Instant", {
      cant_cast: "Cast this spell only if you control a legendary creature or planeswalker",
      legal_targets: { cards: [] },
    });
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(c, s, "p0").reason).toBe(
      "Cast this spell only if you control a legendary creature or planeswalker",
    );
  });

  it("falls back to the sorcery-speed window as the hint", () => {
    const wrath = card("Wrath", "Sorcery");
    const s = snap({ step: "declare_attackers", moves: [passMove] });
    expect(canCastFromHand(wrath, s, "p0").reason).toBe("Only at sorcery speed");
  });

  it("with the window open and nothing else to say, stays generic", () => {
    const wrath = card("Wrath", "Sorcery");
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(wrath, s, "p0").reason).toBe("Can't play this right now");
  });
});

// #1168: the target / mode / additional-cost gates used to read the
// card's own top-level block, which is face 0's — always the ACTIVE
// face for a card in hand. A modal DFC whose FRONT targets and can't,
// or an adventure card whose CREATURE half targets and can't, denied
// with "No legal target" even when the other half needs no target at
// all. The verdict is still the server's "no" here (moves is [] in
// every case below, same as the single-face suite above) — a face's
// local pass can't speak for mana or timing, which stay the server's
// alone (see this file's docblock) — but the REASON must stop blaming
// a target clause a castable face doesn't have.
//
// Every fixture below mirrors its front face at the TOP level too,
// exactly as the server stamps it (cardAsFace's docblock: "for the
// face that is up the two blocks are the same answer") — which is
// what makes these tests sensitive to the fix: the OLD code read that
// top-level mirror directly and never looked at `faces` at all.
describe("canCastFromHand — a multi-face card's OTHER half (#1168)", () => {
  function frontTargetsBackDoesNot(): CardView {
    return card("Twin Path", "Instant", {
      layout: "modal_dfc",
      legal_targets: { cards: [] },
      faces: [
        { name: "Twin Path", type_line: "Instant", legal_targets: { cards: [] } },
        { name: "Twin Path's Reverse", type_line: "Instant" },
      ],
    });
  }

  it("a modal DFC whose front targets and whose back does not isn't denied for the front's target", () => {
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(frontTargetsBackDoesNot(), s, "p0").reason).toBe(
      "Can't play this right now",
    );
  });

  function creatureTargetsAdventureDoesNot(): CardView {
    return card("Ambush Wolf", "Creature — Wolf", {
      layout: "adventure",
      legal_targets: { cards: [] },
      faces: [
        { name: "Ambush Wolf", type_line: "Creature — Wolf", legal_targets: { cards: [] } },
        {
          name: "Pounce",
          type_line: "Instant — Adventure",
          legal_targets: { cards: ["bear"], min: 1 },
        },
      ],
    });
  }

  it("an adventure card whose Adventure half is the only castable one isn't denied for the creature's target", () => {
    const s = snap({ moves: [passMove] });
    expect(canCastFromHand(creatureTargetsAdventureDoesNot(), s, "p0").reason).toBe(
      "Can't play this right now",
    );
  });

  it("a card castable on no face stays greyed, and the reason names which half it's about", () => {
    const bothBlocked = card("Twin Path", "Instant", {
      layout: "modal_dfc",
      legal_targets: { cards: [] },
      faces: [
        { name: "Twin Path", type_line: "Instant", legal_targets: { cards: [] } },
        { name: "Twin Path's Reverse", type_line: "Instant", legal_targets: { cards: [], min: 2 } },
      ],
    });
    const s = snap({ moves: [passMove] });
    const got = canCastFromHand(bothBlocked, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBe("Twin Path: No legal target");
  });

  // The single-face cases are unchanged: none of the fixtures in the
  // "denial reasons" and "verdict" suites above carry `faces`, so
  // `castableFaces` collapses to `[card]` and `named()` never
  // prefixes — every reason there reads exactly as it did before.

  // #1185: `cant_cast` lives on `CastSurfaceView`, so it travels with
  // the rest of the per-face block `cardAsFace` swaps in — a face
  // whose own printed clause refuses it is named, not the card's front.
  it("a modal DFC whose back face is the one cant_cast blocks is named for that face", () => {
    const c = card("Twin Path", "Instant", {
      layout: "modal_dfc",
      faces: [
        { name: "Twin Path", type_line: "Instant" },
        {
          name: "Twin Path's Reverse",
          type_line: "Instant",
          cant_cast: "Cast this spell only during combat",
        },
      ],
    });
    const s = snap({ moves: [passMove] });
    // The FRONT face has no cant_cast, so it alone rescues the card —
    // the verdict here is the generic fallback (the server's move
    // list still says no), not the back's own clause, exactly as
    // #1168's own "front targets, back doesn't" case reads.
    expect(canCastFromHand(c, s, "p0").reason).toBe("Can't play this right now");
  });

  it("a card refused by cant_cast on every castable face names the first blocked one", () => {
    const c = card("Twin Path", "Instant", {
      layout: "modal_dfc",
      cant_cast: "Cast this spell only during combat",
      faces: [
        {
          name: "Twin Path",
          type_line: "Instant",
          cant_cast: "Cast this spell only during combat",
        },
        {
          name: "Twin Path's Reverse",
          type_line: "Instant",
          cant_cast: "Cast this spell only if you control a legendary creature",
        },
      ],
    });
    const s = snap({ moves: [passMove] });
    const got = canCastFromHand(c, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBe("Twin Path: Cast this spell only during combat");
  });
});

// The compatibility half. A frame with no move list is NO
// INFORMATION, not a refusal — an older server, or simply a frame
// where this seat owes no decision. Greying the hand on those would
// be the worst possible regression.
describe("canCastFromHand — permissive when the server said nothing", () => {
  it("stays legal with priority and no move list", () => {
    const s = snap();
    expect(canCastFromHand(card("Wrath", "Sorcery"), s, "p0").legal).toBe(true);
    expect(canCastFromHand(card("Forest", "Basic Land — Forest"), s, "p0").legal).toBe(true);
  });

  it("still reports the two facts the frame does carry", () => {
    expect(canCastFromHand(card("Bolt", "Instant"), snap({ priorityHolder: 1 }), "p0").reason).toBe(
      "Not your priority",
    );
    expect(canCastFromHand(card("Bolt", "Instant"), snap({ splitSecond: true }), "p0").reason).toBe(
      "Split second on the stack",
    );
  });
});

// The last rules derivation in the file, and the reason it survives:
// `activate_loyalty` is a sandbox verb the enumerator does not
// enumerate, so there is no move list to look it up in.
describe("canActivateSorcerySpeedAbility", () => {
  it("open on your own main phase with an empty stack", () => {
    expect(canActivateSorcerySpeedAbility(snap(), "p0").legal).toBe(true);
  });

  it("shut outside a main phase", () => {
    expect(canActivateSorcerySpeedAbility(snap({ step: "upkeep" }), "p0").reason).toBe(
      "Sorcery-speed only",
    );
  });

  it("shut with a non-empty stack", () => {
    const s = snap({
      stackItems: [{ id: "x", kind: "spell", controller: "p1", owner: "p1", source_card_id: "x" }],
    });
    expect(canActivateSorcerySpeedAbility(s, "p0").reason).toBe("Stack isn't empty");
  });

  it("shut on someone else's turn", () => {
    expect(
      canActivateSorcerySpeedAbility(snap({ activeSeat: 1, priorityHolder: 0 }), "p0").reason,
    ).toBe("Not your turn");
  });

  it("shut without priority, and for spectators", () => {
    expect(canActivateSorcerySpeedAbility(snap({ priorityHolder: 1 }), "p0").reason).toBe(
      "Not your priority",
    );
    expect(canActivateSorcerySpeedAbility(snap(), null).reason).toBe("Spectator can't activate");
  });

  it("shut under split second", () => {
    expect(canActivateSorcerySpeedAbility(snap({ splitSecond: true }), "p0").reason).toBe(
      "Split second on the stack",
    );
  });
});

describe("canActivateLoyalty", () => {
  const pwID = "pw-1";
  function snapWithPW(o: SnapOpts = {}): GameView {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    return snap({ ...o, battlefield: [c] });
  }

  it("legal on own main phase, stack empty, not yet activated", () => {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(canActivateLoyalty(c, snapWithPW(), "p0", false).legal).toBe(true);
  });

  it("rejects on opponent's turn", () => {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(
      canActivateLoyalty(c, snapWithPW({ activeSeat: 1, priorityHolder: 0 }), "p0", false).reason,
    ).toBe("Not your turn");
  });

  it("rejects when already activated this turn, from the flag or the override", () => {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(canActivateLoyalty(c, snapWithPW(), "p0", true).reason).toBe(
      "Already activated this turn",
    );
    const stamped = card("Jace", "Legendary Planeswalker — Jace", {
      instance_id: pwID,
      loyalty_activated: true,
    });
    expect(canActivateLoyalty(stamped, snapWithPW(), "p0", false).reason).toBe(
      "Already activated this turn",
    );
  });

  it("rejects when source not on battlefield", () => {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(canActivateLoyalty(c, snap(), "p0", false).reason).toBe(
      "Planeswalker not on the battlefield",
    );
  });

  it("does NOT consult legal_moves — activate_loyalty is not enumerated", () => {
    // An empty move list must not grey a loyalty row: the enumerator
    // deliberately skips the sandbox loyalty verb, so its silence
    // says nothing about it.
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(canActivateLoyalty(c, snapWithPW({ moves: [passMove] }), "p0", false).legal).toBe(true);
  });
});
