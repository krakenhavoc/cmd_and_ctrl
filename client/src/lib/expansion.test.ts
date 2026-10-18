import { describe, it, expect } from "vitest";
import type { CardView, GameView } from "./protocol";
import type { TargetingState } from "./targeting";
import {
  decideReflow,
  decideSeatRendering,
  expansionSettingsFor,
  interactionReason,
  legalDefenderIDs,
  REFLOW_AVATAR_COLUMN,
  reflowAvatarOrder,
  reflowGrid,
  type ReflowCell,
  type ReflowContext,
  type TurnOrderSeat,
  type ReflowSeat,
  seatControlsLegalTarget,
  seatHasAttackersOn,
  type ExpansionSettings,
  type SeatSignals,
  type TableSignals,
} from "./expansion";

// The defaults: an opponent seat, nothing happening, settings as they
// ship. Every test below changes exactly one thing, so a failure names
// the rule that broke rather than "something in the object".
function seat(over: Partial<SeatSignals> = {}): SeatSignals {
  return {
    isSelf: false,
    isActiveSeat: false,
    controlsLegalTarget: false,
    hasAttackersOnViewer: false,
    isLegalDefender: false,
    ...over,
  };
}

function table(over: Partial<TableSignals> = {}): TableSignals {
  return { spectator: false, combatMode: "idle", ...over };
}

function shipped(over: Partial<ExpansionSettings> = {}): ExpansionSettings {
  return { opponentDetail: "summary", expandActivePlayer: true, ...over };
}

function card(over: Partial<CardView> = {}): CardView {
  return { instance_id: "c1", name: "Grizzly Bears", ...over } as CardView;
}

describe("decideSeatRendering", () => {
  it("renders a quiet opponent seat as a summary", () => {
    const d = decideSeatRendering(seat(), table(), shipped());
    expect(d.rendering).toBe("summary");
    expect(d.reason).toBeNull();
  });

  it("never summarises the viewer's own seat", () => {
    // Even with the setting on summary and nothing else true: you
    // cannot play out of a hand you are not being shown.
    expect(decideSeatRendering(seat({ isSelf: true }), table(), shipped())).toEqual({
      rendering: "full",
      reason: "self",
    });
  });

  it("gives spectators the old uniform grid", () => {
    expect(decideSeatRendering(seat(), table({ spectator: true }), shipped())).toEqual({
      rendering: "full",
      reason: "spectator",
    });
  });

  it('opponentDetail "full" restores the pre-feature table exactly', () => {
    // The escape hatch has to beat every expansion trigger, or a
    // player who opted out would still see panels changing size.
    const everything = seat({
      isActiveSeat: true,
      controlsLegalTarget: true,
      hasAttackersOnViewer: true,
      isLegalDefender: true,
    });
    const d = decideSeatRendering(
      everything,
      table({ combatMode: "block" }),
      shipped({ opponentDetail: "full" }),
    );
    expect(d).toEqual({ rendering: "full", reason: "setting" });
  });

  it("expands a seat holding a legal target while a prompt is live", () => {
    expect(decideSeatRendering(seat({ controlsLegalTarget: true }), table(), shipped())).toEqual({
      rendering: "full",
      reason: "targeting",
    });
  });

  it("expands a seat with attackers on the viewer, but only in block mode", () => {
    const s = seat({ hasAttackersOnViewer: true });
    expect(decideSeatRendering(s, table({ combatMode: "block" }), shipped()).reason).toBe(
      "blocking",
    );
    // Idle: the creatures are still attacking, but the viewer has not
    // entered block mode, so nothing has asked them to click one.
    expect(decideSeatRendering(s, table(), shipped()).rendering).toBe("summary");
  });

  it("expands a legal defender, but only in attack mode", () => {
    const s = seat({ isLegalDefender: true });
    expect(decideSeatRendering(s, table({ combatMode: "attack" }), shipped()).reason).toBe(
      "attacking",
    );
    expect(decideSeatRendering(s, table(), shipped()).rendering).toBe("summary");
  });

  it("does not expand a defender during BLOCK mode, or an attacker during ATTACK mode", () => {
    // The two combat triggers are not interchangeable. Blocking looks
    // at who is attacking you; attacking looks at who you may attack.
    expect(
      decideSeatRendering(
        seat({ isLegalDefender: true }),
        table({ combatMode: "block" }),
        shipped(),
      ).rendering,
    ).toBe("summary");
    expect(
      decideSeatRendering(
        seat({ hasAttackersOnViewer: true }),
        table({ combatMode: "attack" }),
        shipped(),
      ).rendering,
    ).toBe("summary");
  });

  it("expands the active player by default and stops when the setting is off", () => {
    const s = seat({ isActiveSeat: true });
    expect(decideSeatRendering(s, table(), shipped()).reason).toBe("active-player");
    expect(decideSeatRendering(s, table(), shipped({ expandActivePlayer: false })).rendering).toBe(
      "summary",
    );
  });

  it("still expands the active player for a live prompt when the setting is off", () => {
    // Turning the convenience off must not turn off the requirement:
    // a prompt you cannot answer is a wedged client.
    const s = seat({ isActiveSeat: true, controlsLegalTarget: true });
    expect(decideSeatRendering(s, table(), shipped({ expandActivePlayer: false })).reason).toBe(
      "targeting",
    );
  });
});

describe("expansionSettingsFor", () => {
  const fullAndActive = { opponentDetail: "full", expandActivePlayer: true } as const;

  it("passes the two settings through in the quadrant and row layouts", () => {
    for (const tableLayout of ["quadrant", "row"] as const) {
      expect(expansionSettingsFor({ ...fullAndActive, tableLayout })).toEqual(fullAndActive);
    }
  });

  it("the focus layout keeps every opponent a summary, on their turn too", () => {
    const focus = expansionSettingsFor({ ...fullAndActive, tableLayout: "focus" });
    expect(focus).toEqual({ opponentDetail: "summary", expandActivePlayer: false });
    expect(decideSeatRendering(seat(), table(), focus).rendering).toBe("summary");
    expect(decideSeatRendering(seat({ isActiveSeat: true }), table(), focus).rendering).toBe(
      "summary",
    );
  });

  it("the focus layout still expands a seat a prompt needs you to click", () => {
    const focus = expansionSettingsFor({ ...fullAndActive, tableLayout: "focus" });
    expect(decideSeatRendering(seat({ controlsLegalTarget: true }), table(), focus).reason).toBe(
      "targeting",
    );
    expect(
      decideSeatRendering(
        seat({ hasAttackersOnViewer: true }),
        table({ combatMode: "block" }),
        focus,
      ).reason,
    ).toBe("blocking");
  });
});

describe("interactionReason", () => {
  it("names the trigger a live prompt needs, and nothing otherwise", () => {
    expect(interactionReason(seat(), table())).toBeNull();
    expect(interactionReason(seat({ controlsLegalTarget: true }), table())).toBe("targeting");
    expect(
      interactionReason(seat({ hasAttackersOnViewer: true }), table({ combatMode: "block" })),
    ).toBe("blocking");
    expect(
      interactionReason(seat({ isLegalDefender: true }), table({ combatMode: "attack" })),
    ).toBe("attacking");
  });

  it("ignores attackers and defenders outside their combat mode", () => {
    expect(interactionReason(seat({ hasAttackersOnViewer: true }), table())).toBeNull();
    expect(interactionReason(seat({ isLegalDefender: true }), table())).toBeNull();
    expect(
      interactionReason(seat({ isLegalDefender: true }), table({ combatMode: "block" })),
    ).toBeNull();
  });

  it("is not the active-player expansion: a turn alone needs no board", () => {
    expect(interactionReason(seat({ isActiveSeat: true }), table())).toBeNull();
  });
});

// ADR 0077 §6 as amended 2026-10-16 (#4065): the reflow.
describe("decideReflow", () => {
  function opp(id: string, over: Partial<ReflowSeat> = {}): ReflowSeat {
    return { id, isActiveSeat: false, needsBoard: false, ...over };
  }
  function ctx(over: Partial<ReflowContext> = {}): ReflowContext {
    return {
      style: "reflow",
      expandActivePlayer: true,
      opponentsShareRow: true,
      spectator: false,
      ...over,
    };
  }
  // Bob's turn at a four-seat row table.
  const bobsTurn = [opp("bob", { isActiveSeat: true }), opp("cat"), opp("dee")];

  it("on an opponent's turn, that opponent is the focus and the others are avatar cards", () => {
    const r = decideReflow(bobsTurn, ctx())!;
    expect(r.focusSeatID).toBe("bob");
    expect([...r.avatarSeatIDs].sort()).toEqual(["cat", "dee"]);
  });

  it("follows the active seat wherever it sits in the row", () => {
    const r = decideReflow([opp("bob"), opp("cat"), opp("dee", { isActiveSeat: true })], ctx())!;
    expect(r.focusSeatID).toBe("dee");
    expect([...r.avatarSeatIDs].sort()).toEqual(["bob", "cat"]);
  });

  it("does nothing on the viewer's own turn: the row stays equal boards", () => {
    expect(decideReflow([opp("bob"), opp("cat"), opp("dee")], ctx())).toBeNull();
  });

  it("does nothing with the overlay style, which keeps the boards equal", () => {
    expect(decideReflow(bobsTurn, ctx({ style: "overlay" }))).toBeNull();
  });

  it("does nothing with the active-player expansion off", () => {
    expect(decideReflow(bobsTurn, ctx({ expandActivePlayer: false }))).toBeNull();
  });

  it("does nothing in the focus layout, whose settings turn the expansion off", () => {
    const off = expansionSettingsFor({
      opponentDetail: "full",
      expandActivePlayer: true,
      tableLayout: "focus",
    });
    expect(decideReflow(bobsTurn, ctx({ expandActivePlayer: off.expandActivePlayer }))).toBeNull();
  });

  it("does nothing when the opponents do not share one row, or for a spectator", () => {
    expect(decideReflow(bobsTurn, ctx({ opponentsShareRow: false }))).toBeNull();
    expect(decideReflow(bobsTurn, ctx({ spectator: true }))).toBeNull();
  });

  it("does nothing with one opponent: there is nobody to shrink", () => {
    expect(decideReflow([opp("bob", { isActiveSeat: true })], ctx())).toBeNull();
  });

  it("works at three players: one avatar card", () => {
    const r = decideReflow([opp("bob", { isActiveSeat: true }), opp("cat")], ctx())!;
    expect([...r.avatarSeatIDs]).toEqual(["cat"]);
  });

  it("never shrinks a seat a live prompt needs", () => {
    const r = decideReflow(
      [opp("bob", { isActiveSeat: true }), opp("cat", { needsBoard: true }), opp("dee")],
      ctx(),
    )!;
    expect(r.focusSeatID).toBe("bob");
    expect([...r.avatarSeatIDs]).toEqual(["dee"]);
  });
});

// ADR 0077 as amended 2026-10-17 (#4102): one column of avatar cards at
// the row's left end, stacked in play order, the full boards to its right.
describe("reflowGrid", () => {
  // Bob sits next, Cat across, Dee across_next: the row's turn order.
  const cells: ReflowCell[] = [
    { id: "bob", area: "next" },
    { id: "cat", area: "across" },
    { id: "dee", area: "across_next" },
  ];

  it("stacks two avatar cards in one left column, each half the row, the focus beside them", () => {
    const r = { focusSeatID: "bob", avatarSeatIDs: new Set(["cat", "dee"]) };
    const g = reflowGrid(cells, ["cat", "dee"], r);
    expect(g.columns).toBe(`${REFLOW_AVATAR_COLUMN} minmax(0, 2fr)`);
    expect(g.rows).toBe("minmax(0, 0.5fr) minmax(0, 0.5fr) minmax(0, 1fr)");
    expect(g.areas).toBe('"across next" "across_next next" "self self"');
  });

  it("puts the column on the left whatever seat is active, in the order given", () => {
    const r = { focusSeatID: "dee", avatarSeatIDs: new Set(["bob", "cat"]) };
    const g = reflowGrid(cells, ["bob", "cat"], r);
    expect(g.columns).toBe(`${REFLOW_AVATAR_COLUMN} minmax(0, 2fr)`);
    expect(g.areas).toBe('"next across_next" "across across_next" "self self"');
    // The order is the caller's (reflowAvatarOrder's), not the row's.
    expect(reflowGrid(cells, ["cat", "bob"], r).areas).toBe(
      '"across across_next" "next across_next" "self self"',
    );
  });

  it("gives one avatar card the whole row's height, and keeps the layout's split", () => {
    const two: ReflowCell[] = [
      { id: "bob", area: "next" },
      { id: "cat", area: "across" },
    ];
    const r = { focusSeatID: "cat", avatarSeatIDs: new Set(["bob"]) };
    const g = reflowGrid(two, ["bob"], r, [0.7, 1.3]);
    expect(g.columns).toBe(`${REFLOW_AVATAR_COLUMN} minmax(0, 2fr)`);
    expect(g.rows).toBe("minmax(0, 0.7fr) minmax(0, 1.3fr)");
    expect(g.areas).toBe('"next across" "self self"');
    // The quadrant split, shared by two cards.
    const r3 = { focusSeatID: "bob", avatarSeatIDs: new Set(["cat", "dee"]) };
    expect(reflowGrid(cells, ["cat", "dee"], r3, [0.7, 1.3]).rows).toBe(
      "minmax(0, 0.35fr) minmax(0, 0.35fr) minmax(0, 1.3fr)",
    );
  });

  it("keeps a board a prompt needs beside the focus, in turn order, one share", () => {
    const r = { focusSeatID: "dee", avatarSeatIDs: new Set(["bob"]) };
    const g = reflowGrid(cells, ["bob"], r);
    expect(g.columns).toBe(`${REFLOW_AVATAR_COLUMN} minmax(0, 1fr) minmax(0, 2fr)`);
    expect(g.areas).toBe('"next across across_next" "self self self"');
  });

  it("has no column when no seat is an avatar card", () => {
    const r = { focusSeatID: "bob", avatarSeatIDs: new Set<string>() };
    const g = reflowGrid(cells, [], r);
    expect(g.columns).toBe("minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1fr)");
    expect(g.rows).toBe("minmax(0, 1fr) minmax(0, 1fr)");
    expect(g.areas).toBe('"next across across_next" "self self self"');
  });
});

describe("reflowAvatarOrder", () => {
  // You, Bot 1, Bot 2 and Bot 3, seated (and so playing) in that order.
  const table = (out: string[] = []): TurnOrderSeat[] =>
    [
      { id: "me", name: "You" },
      { id: "b1", name: "Bot 1" },
      { id: "b2", name: "Bot 2" },
      { id: "b3", name: "Bot 3" },
    ].map((s) => ({ ...s, eliminated: out.includes(s.id) }));
  const others = (active: string, ids = ["b1", "b2", "b3"]) => ids.filter((id) => id !== active);
  const read = (o: ReturnType<typeof reflowAvatarOrder>) =>
    o.map((c) => `${c.id}:${c.out ? "out" : c.chip}`);

  it("on Bot 1's turn: Bot 2 is next up, then Bot 3 after Bot 2", () => {
    expect(read(reflowAvatarOrder(table(), "b1", "me", others("b1")))).toEqual([
      "b2:Next up",
      "b3:After Bot 2",
    ]);
  });

  it("on Bot 2's turn: Bot 3 is next up, then Bot 1 after you (the amendment's example)", () => {
    expect(read(reflowAvatarOrder(table(), "b2", "me", others("b2")))).toEqual([
      "b3:Next up",
      "b1:After you",
    ]);
  });

  it("on Bot 3's turn: Bot 1 after you, then Bot 2 after Bot 1 (the amendment's example)", () => {
    expect(read(reflowAvatarOrder(table(), "b3", "me", others("b3")))).toEqual([
      "b1:After you",
      "b2:After Bot 1",
    ]);
  });

  it("at three players: the one card is next up or after you", () => {
    const three = table().slice(0, 3);
    expect(read(reflowAvatarOrder(three, "b1", "me", ["b2"]))).toEqual(["b2:Next up"]);
    expect(read(reflowAvatarOrder(three, "b2", "me", ["b1"]))).toEqual(["b1:After you"]);
  });

  it("sorts the viewer's seat anywhere at the table, not only first", () => {
    // The viewer is Bot 2's neighbour: seated third.
    const seats: TurnOrderSeat[] = [
      { id: "b1", name: "Bot 1" },
      { id: "b2", name: "Bot 2" },
      { id: "me", name: "You" },
      { id: "b3", name: "Bot 3" },
    ];
    expect(read(reflowAvatarOrder(seats, "b1", "me", ["b2", "b3"]))).toEqual([
      "b2:Next up",
      "b3:After you",
    ]);
    expect(read(reflowAvatarOrder(seats, "b3", "me", ["b1", "b2"]))).toEqual([
      "b1:Next up",
      "b2:After Bot 1",
    ]);
  });

  it("sorts an eliminated opponent last, with no chip, and skips it when naming who plays before", () => {
    // Bot 1 is out: on Bot 3's turn, Bot 2 now plays straight after you.
    expect(read(reflowAvatarOrder(table(["b1"]), "b3", "me", others("b3")))).toEqual([
      "b2:After you",
      "b1:out",
    ]);
    // Bot 2 is out: on Bot 1's turn, Bot 3 is next up.
    expect(read(reflowAvatarOrder(table(["b2"]), "b1", "me", others("b1")))).toEqual([
      "b3:Next up",
      "b2:out",
    ]);
  });

  it("skips an eliminated viewer, who plays no turn, so no card reads after you", () => {
    expect(read(reflowAvatarOrder(table(["me"]), "b3", "me", others("b3")))).toEqual([
      "b1:Next up",
      "b2:After Bot 1",
    ]);
    expect(read(reflowAvatarOrder(table(["me"]), "b2", "me", others("b2")))).toEqual([
      "b3:Next up",
      "b1:After Bot 3",
    ]);
  });

  it("with no viewer seat (a spectator), names every seat before by its name", () => {
    expect(read(reflowAvatarOrder(table(), "b3", null, ["me", "b1", "b2"]))).toEqual([
      "me:Next up",
      "b1:After You",
      "b2:After Bot 1",
    ]);
  });

  it("drops the active seat and seats not at the table", () => {
    expect(read(reflowAvatarOrder(table(), "b1", "me", ["b1", "zz", "b2"]))).toEqual([
      "b2:Next up",
    ]);
    expect(reflowAvatarOrder(table(), "zz", "me", ["b1"])).toEqual([]);
  });
});

describe("seatControlsLegalTarget", () => {
  // `legal` is the server-computed set; targeting.ts prefers it over
  // the mode heuristics, so these tests pin the real path.
  function prompt(players: string[], cards: string[]): TargetingState {
    return {
      card: card({ instance_id: "src" }),
      mode: "creature",
      legal: { players: new Set(players), cards: new Set(cards) },
    } as TargetingState;
  }

  it("is false with no prompt open", () => {
    expect(seatControlsLegalTarget(null, "p2", [card()])).toBe(false);
  });

  it("is true when the seat itself is a legal player target", () => {
    expect(seatControlsLegalTarget(prompt(["p2"], []), "p2", [])).toBe(true);
  });

  it("is true when the seat controls a legal card target", () => {
    const c = card({ instance_id: "bear" });
    expect(seatControlsLegalTarget(prompt([], ["bear"]), "p2", [c])).toBe(true);
  });

  it("is false when the legal card belongs to someone else", () => {
    // The seat's own slice is what decides — a legal target on
    // another board must not expand this one.
    const c = card({ instance_id: "bear" });
    expect(seatControlsLegalTarget(prompt([], ["ogre"]), "p2", [c])).toBe(false);
  });
});

describe("seatHasAttackersOn", () => {
  it("is false for a spectator with no seat", () => {
    expect(seatHasAttackersOn(null, [card({ attacking_target: "p1" })])).toBe(false);
  });

  it("finds a creature attacking the viewer", () => {
    expect(seatHasAttackersOn("p1", [card({ attacking_target: "p1" })])).toBe(true);
  });

  it("ignores a creature attacking somebody else", () => {
    expect(seatHasAttackersOn("p1", [card({ attacking_target: "p3" })])).toBe(false);
  });

  it("ignores an unattacking board", () => {
    expect(seatHasAttackersOn("p1", [card(), card({ instance_id: "c2" })])).toBe(false);
  });
});

describe("legalDefenderIDs", () => {
  function view(activeSeat: number, targets: { id: string; kind: string }[]): GameView {
    return {
      seats: [{ id: "p1" }, { id: "p2" }],
      turn: { active_seat: activeSeat, attack_targets: targets },
    } as unknown as GameView;
  }

  it("is empty when the viewer is not the active player", () => {
    // attack_targets is published for the active seat only, so reading
    // it as anyone else would be reading someone else's legal set.
    expect(legalDefenderIDs(view(1, [{ id: "p2", kind: "player" }]), "p1").size).toBe(0);
  });

  it("is empty for a spectator", () => {
    expect(legalDefenderIDs(view(0, [{ id: "p2", kind: "player" }]), null).size).toBe(0);
  });

  it("returns the player rows for the active viewer", () => {
    const ids = legalDefenderIDs(view(0, [{ id: "p2", kind: "player" }]), "p1");
    expect([...ids]).toEqual(["p2"]);
  });

  it("drops planeswalker and battle rows", () => {
    // Those are permanents inside a panel, not seats. Expanding a seat
    // because one of its planeswalkers is attackable is a different
    // rule, and not one this ADR makes.
    const ids = legalDefenderIDs(
      view(0, [
        { id: "p2", kind: "player" },
        { id: "walker", kind: "planeswalker" },
        { id: "siege", kind: "battle" },
      ]),
      "p1",
    );
    expect([...ids]).toEqual(["p2"]);
  });

  it("survives a snapshot with no attack_targets at all", () => {
    const v = { seats: [{ id: "p1" }], turn: { active_seat: 0 } } as unknown as GameView;
    expect(legalDefenderIDs(v, "p1").size).toBe(0);
  });
});
