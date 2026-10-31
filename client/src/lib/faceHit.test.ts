// #4432, #4543 — the face_hit decision: how hard a blow is, which seats a
// beat hit, and whether the life popup stands down for a combat hit.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  faceHitSeats,
  faceHitSound,
  faceHits,
  playStepWarCry,
  popupCombatDamage,
  popupSound,
  stepFaceHits,
  warCryPlan,
} from "./faceHit";
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

describe("faceHitSound", () => {
  it("is light for exactly 1 damage", () => {
    expect(faceHitSound(1)).toBe("face_hit_light");
  });

  it("is medium for 2 and 3 damage", () => {
    expect(faceHitSound(2)).toBe("face_hit_medium");
    expect(faceHitSound(3)).toBe("face_hit_medium");
  });

  it("is heavy for 4 or more damage", () => {
    expect(faceHitSound(4)).toBe("face_hit_heavy");
    expect(faceHitSound(7)).toBe("face_hit_heavy");
    expect(faceHitSound(21)).toBe("face_hit_heavy");
  });

  it("is silent for no damage", () => {
    expect(faceHitSound(0)).toBeNull();
    expect(faceHitSound(-3)).toBeNull();
    expect(faceHitSound(Number.NaN)).toBeNull();
  });
});

describe("faceHits", () => {
  it("names each player combat damage hit, once, with all the damage they took", () => {
    const entries = [dmg(1, 1, 3, true), dmg(2, 1, 2, true), dmg(3, 2, 4, true)];
    expect(faceHits(entries)).toEqual([
      { seat: 1, amount: 5 },
      { seat: 2, amount: 4 },
    ]);
    expect(faceHitSeats(entries)).toEqual([1, 2]);
  });

  it("makes two small hits on one player one bigger blow", () => {
    const hits = faceHits([dmg(1, 1, 2, true), dmg(2, 1, 2, true)]);
    expect(hits).toEqual([{ seat: 1, amount: 4 }]);
    expect(faceHitSound(hits[0].amount)).toBe("face_hit_heavy");
  });

  it("ignores noncombat damage, damage to a card, and fully prevented damage", () => {
    expect(faceHits([dmg(1, 1, 3, false)])).toEqual([]);
    expect(faceHits([dmg(2, undefined, 3, true, { target: "c1" })])).toEqual([]);
    expect(faceHits([dmg(3, 1, 0, true)])).toEqual([]);
    expect(faceHits([life(4, 1, -2)])).toEqual([]);
  });

  it("counts commander damage like any other combat damage", () => {
    // Commander damage is combat damage to the player; the log entry is the same.
    expect(faceHits([dmg(1, 2, 5, true, { card_id: "cmdr" })])).toEqual([{ seat: 2, amount: 5 }]);
  });
});

describe("popupCombatDamage", () => {
  it("is the combat damage when the popup's newest loss is combat damage to the seat", () => {
    expect(popupCombatDamage(view([dmg(1, 1, 3, true)]), 1, [-3])).toBe(3);
  });

  it("is 0 for a drain or noncombat damage", () => {
    expect(popupCombatDamage(view([life(1, 1, -3)]), 1, [-3])).toBe(0);
    expect(popupCombatDamage(view([dmg(1, 1, 3, false)]), 1, [-3])).toBe(0);
  });

  it("is 0 when the deltas do not line up with the log", () => {
    expect(popupCombatDamage(view([dmg(1, 1, 3, true)]), 1, [-2])).toBe(0);
    expect(popupCombatDamage(view([]), 1, [-3])).toBe(0);
    expect(popupCombatDamage(view([dmg(1, 1, 3, true)]), -1, [-3])).toBe(0);
  });

  it("counts only the combat part of a popup that also holds a drain", () => {
    expect(popupCombatDamage(view([life(1, 1, -1), dmg(2, 1, 3, true)]), 1, [-1, -3])).toBe(3);
  });

  it("adds a double strike's two combat hits", () => {
    expect(popupCombatDamage(view([dmg(1, 1, 2, true), dmg(2, 1, 2, true)]), 1, [-2, -2])).toBe(4);
  });
});

describe("popupSound", () => {
  it("keeps damage for non-combat loss and heal for gains", () => {
    expect(popupSound("damage", 0, true)).toBe("damage");
    expect(popupSound("damage", 0, false)).toBe("damage");
    expect(popupSound("heal", 0, true)).toBe("heal");
    expect(popupSound("heal", 3, false)).toBe("heal");
  });

  it("is silent for a combat hit the strike layer plays", () => {
    expect(popupSound("damage", 3, true)).toBeNull();
  });

  it("plays the hit's tier itself when no strike plays", () => {
    expect(popupSound("damage", 1, false)).toBe("face_hit_light");
    expect(popupSound("damage", 2, false)).toBe("face_hit_medium");
    expect(popupSound("damage", 3, false)).toBe("face_hit_medium");
    expect(popupSound("damage", 4, false)).toBe("face_hit_heavy");
    expect(popupSound("damage", 7, false)).toBe("face_hit_heavy");
  });
});

describe("warCryPlan (#4560)", () => {
  it("is at contact and full volume when no player is hit", () => {
    expect(warCryPlan(0)).toEqual({ delayMs: 0, volume: 1 });
  });
  it("follows one hit by 300 ms at 0.6", () => {
    expect(warCryPlan(1)).toEqual({ delayMs: 300, volume: 0.6 });
  });
  it("follows the last of several staggered hits", () => {
    expect(warCryPlan(3)).toEqual({ delayMs: 2 * 90 + 300, volume: 0.6 });
  });
});

describe("stepFaceHits and playStepWarCry, the step-change fallback (#4560)", () => {
  const step = (seq: number): LogEvent => ({ seq, kind: "step", seat: 0, text: "" }) as LogEvent;
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it("reads only the entries after the newest step entry", () => {
    const log = [step(1), dmg(2, 1, 3, true), step(3), dmg(4, undefined, 2, true)];
    expect(stepFaceHits(log)).toEqual([]);
    expect(stepFaceHits([...log, dmg(5, 1, 2, true)])).toEqual([{ seat: 1, amount: 2 }]);
  });

  it("plays the war cry at once, at normal volume, when no player is hit", () => {
    const play = vi.fn();
    expect(playStepWarCry([step(1), dmg(2, undefined, 2, true)], play)).toBeNull();
    expect(play).toHaveBeenCalledExactlyOnceWith("combat_resolve");
  });

  it("delays a quieter war cry by 300 ms when a player is hit", () => {
    const play = vi.fn();
    const fired = vi.fn();
    const t = playStepWarCry([step(1), dmg(2, 1, 3, true)], play, fired);
    expect(t).not.toBeNull();
    vi.advanceTimersByTime(299);
    expect(play).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(play).toHaveBeenCalledExactlyOnceWith("combat_resolve", { volume: 0.6 });
    expect(fired).toHaveBeenCalledWith(t);
  });
});
