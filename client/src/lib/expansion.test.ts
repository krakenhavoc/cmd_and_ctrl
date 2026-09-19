import { describe, it, expect } from "vitest";
import type { CardView, GameView } from "./protocol";
import type { TargetingState } from "./targeting";
import {
  decideSeatRendering,
  legalDefenderIDs,
  nextPinnedSeat,
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
    isPinned: false,
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
      isPinned: true,
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

  it("expands a pinned seat and keeps the pin above the interaction triggers", () => {
    // Ordering test, not a duplicate of the one above: a pinned seat
    // that ALSO holds a legal target must report "pinned", so that
    // when the prompt closes the panel does not collapse.
    expect(
      decideSeatRendering(seat({ isPinned: true, controlsLegalTarget: true }), table(), shipped()),
    ).toEqual({ rendering: "full", reason: "pinned" });
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

describe("nextPinnedSeat", () => {
  it("pins a seat when nothing is pinned", () => {
    expect(nextPinnedSeat(null, "p2")).toBe("p2");
  });

  it("unpins when the pinned seat is clicked again", () => {
    expect(nextPinnedSeat("p2", "p2")).toBeNull();
  });

  it("moves the pin rather than adding a second one", () => {
    expect(nextPinnedSeat("p2", "p3")).toBe("p3");
  });
});
