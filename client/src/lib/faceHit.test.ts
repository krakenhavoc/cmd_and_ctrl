// #4432 — the face_hit decision: which seats a beat hit, and whether the
// life popup stands down for a combat hit.

import { describe, expect, it } from "vitest";

import { faceHitSeats, popupHasCombatHit, popupSound } from "./faceHit";
import type { GameView, LogEvent } from "./protocol";

const dmg = (
  seq: number,
  target_seat: number | undefined,
  amount: number,
  combat: boolean,
  extra: Partial<LogEvent> = {},
): LogEvent =>
  ({ seq, kind: "damage", seat: 0, target_seat, amount, combat, text: "", ...extra }) as LogEvent;

const life = (seq: number, seat: number, amount: number): LogEvent =>
  ({ seq, kind: "life", seat, amount, text: "" }) as LogEvent;

const view = (log: LogEvent[]): GameView => ({ log, seats: [{}, {}] }) as unknown as GameView;

describe("faceHitSeats", () => {
  it("names each player combat damage hit, once", () => {
    const entries = [dmg(1, 1, 3, true), dmg(2, 1, 2, true), dmg(3, 2, 4, true)];
    expect(faceHitSeats(entries)).toEqual([1, 2]);
  });

  it("ignores noncombat damage, damage to a card, and fully prevented damage", () => {
    expect(faceHitSeats([dmg(1, 1, 3, false)])).toEqual([]);
    expect(faceHitSeats([dmg(2, undefined, 3, true, { target: "c1" })])).toEqual([]);
    expect(faceHitSeats([dmg(3, 1, 0, true)])).toEqual([]);
    expect(faceHitSeats([life(4, 1, -2)])).toEqual([]);
  });
});

describe("popupHasCombatHit", () => {
  it("is true when the popup's newest loss is combat damage to the seat", () => {
    expect(popupHasCombatHit(view([dmg(1, 1, 3, true)]), 1, [-3])).toBe(true);
  });

  it("is false for a drain or noncombat damage", () => {
    expect(popupHasCombatHit(view([life(1, 1, -3)]), 1, [-3])).toBe(false);
    expect(popupHasCombatHit(view([dmg(1, 1, 3, false)]), 1, [-3])).toBe(false);
  });

  it("is false when the deltas do not line up with the log", () => {
    expect(popupHasCombatHit(view([dmg(1, 1, 3, true)]), 1, [-2])).toBe(false);
    expect(popupHasCombatHit(view([]), 1, [-3])).toBe(false);
    expect(popupHasCombatHit(view([dmg(1, 1, 3, true)]), -1, [-3])).toBe(false);
  });

  it("is true for a combat hit alongside a drain in the same popup", () => {
    expect(popupHasCombatHit(view([life(1, 1, -1), dmg(2, 1, 3, true)]), 1, [-1, -3])).toBe(true);
  });
});

describe("popupSound", () => {
  it("keeps damage for non-combat loss and heal for gains", () => {
    expect(popupSound("damage", false, true)).toBe("damage");
    expect(popupSound("damage", false, false)).toBe("damage");
    expect(popupSound("heal", false, true)).toBe("heal");
  });

  it("is silent for a combat hit the strike layer plays", () => {
    expect(popupSound("damage", true, true)).toBeNull();
  });

  it("plays face_hit itself when no strike plays", () => {
    expect(popupSound("damage", true, false)).toBe("face_hit");
  });
});
