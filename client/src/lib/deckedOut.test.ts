// #4879: the decked-out cue's rules. A player who loses by drawing from an
// empty library (the wire's cause "empty_draw", CR 704.5b) gets the
// library's ghost and player_decked; every other cause keeps the
// elimination cue. The rendered half is deckedOut.render.test.ts.

import { describe, expect, it } from "vitest";

import { SIGNATURE_CAP_MS } from "./fx/signatures";
import type { LogEvent } from "./protocol";
import { SOUND_CATEGORY, SOUND_FALLBACK, SOUND_MANIFEST } from "./sounds";
import {
  DECKED_BASE_MS,
  DECKED_BEATS,
  DECKED_CAP_MS,
  DECKED_CAUSE,
  DECKED_FLING_CAP,
  EFFECT_SHORT_CAP_MS,
  busyScale,
  deckedBadgeDelay,
  deckedFlings,
  deckedSeats,
  deckedTiming,
  deckedWisps,
  effectMotion,
  eliminationCue,
  eliminationSound,
  seatDecked,
} from "./tableEffects";
import { skullKeyframes } from "./deckedOutMotion";

const out = (seq: number, seat: number, cause?: string): LogEvent =>
  ({ seq, kind: "eliminated", seat, cause, turn: 3, text: "" }) as LogEvent;

describe("which cue a departure plays", () => {
  it("is the decked-out cue only for a draw from an empty library", () => {
    expect(DECKED_CAUSE).toBe("empty_draw");
    expect(eliminationCue("empty_draw")).toBe("decked_out");
    for (const cause of ["life", "commander_damage", "poison", "concede", "effect", "other", ""]) {
      expect(eliminationCue(cause)).toBe("eliminated");
    }
    expect(eliminationCue(undefined)).toBe("eliminated");
    expect(eliminationCue(null)).toBe("eliminated");
  });

  it("plays player_decked instead of player_eliminated for that cause alone", () => {
    expect(eliminationSound("empty_draw")).toBe("player_decked");
    for (const cause of ["life", "commander_damage", "poison", "concede", "effect", undefined]) {
      expect(eliminationSound(cause)).toBe("player_eliminated");
    }
  });

  it("finds the decked seats in a frame's log, each once, and no other", () => {
    const log = [
      out(1, 2, "life"),
      out(2, 1, "empty_draw"),
      out(3, 1, "empty_draw"),
      { seq: 4, kind: "draw", seat: 3, turn: 3, text: "" } as LogEvent,
      out(5, 3, "empty_draw"),
      out(6, 0, "concede"),
      out(7, -1, "empty_draw"),
    ];
    expect(deckedSeats(log)).toEqual([1, 3]);
    expect(deckedSeats([out(1, 2, "life"), out(2, 0)])).toEqual([]);
  });

  it("reads a seat as decked from its own loss_cause, once it is out", () => {
    expect(seatDecked({ eliminated: true, loss_cause: "empty_draw" })).toBe(true);
    expect(seatDecked({ eliminated: true, loss_cause: "life" })).toBe(false);
    expect(seatDecked({ eliminated: true })).toBe(false);
    expect(seatDecked({ loss_cause: "empty_draw" })).toBe(false);
  });
});

describe("the sound", () => {
  it("maps player_decked to its two takes, in the effects category", () => {
    expect(SOUND_MANIFEST.player_decked).toEqual([
      "/sounds/player_decked-1.mp3",
      "/sounds/player_decked-2.mp3",
    ]);
    expect(SOUND_CATEGORY.player_decked).toBe("effects");
    expect(SOUND_FALLBACK.player_decked).toBe("player_eliminated");
  });
});

describe("its length and its motion", () => {
  it("runs about 2.2 s at Full, under the signature cap", () => {
    expect(DECKED_CAP_MS).toBe(SIGNATURE_CAP_MS);
    expect(DECKED_BASE_MS).toBe(2200);
    expect(deckedTiming("full", 1)).toEqual({ ms: 2200 });
    expect(deckedTiming("full", 1)!.ms).toBeLessThanOrEqual(SIGNATURE_CAP_MS);
  });

  it("halves in Short, within Short's cap", () => {
    const short = deckedTiming("short", 1)!.ms;
    expect(short).toBe(Math.min(DECKED_BASE_MS / 2, EFFECT_SHORT_CAP_MS));
    expect(short).toBeLessThan(deckedTiming("full", 1)!.ms / 1.9);
  });

  it("scales with animations.speed and a busy table", () => {
    expect(deckedTiming("full", 2)).toEqual({ ms: 4400 });
    const busy = deckedTiming("full", busyScale(4))!.ms;
    expect(busy).toBeCloseTo(2200 * busyScale(4));
    expect(busy).toBeLessThan(2200);
  });

  it("has no motion when still: the grey seat and the badge only", () => {
    expect(deckedTiming("still", 1)).toBeNull();
    expect(deckedBadgeDelay("still", 1)).toBe(0);
  });

  it("is still under reduced motion, either signal, and with effects Off", () => {
    const base = { enabled: true, effects: "full" as const, reduceMotion: false };
    expect(deckedTiming(effectMotion({ ...base, reduceMotion: true }), 1)).toBeNull();
    expect(deckedTiming(effectMotion({ ...base, osReduceMotion: true }), 1)).toBeNull();
    expect(deckedTiming(effectMotion({ ...base, effects: "off" }), 1)).toBeNull();
    expect(deckedTiming(effectMotion({ ...base, enabled: false }), 1)).toBeNull();
    expect(deckedTiming(effectMotion({ ...base, effects: "short" }), 1)).not.toBeNull();
  });

  it("settles the stone late in the cue, after the hand has flown", () => {
    expect(deckedBadgeDelay("full", 1)).toBe(Math.round(2200 * DECKED_BEATS.badge));
    expect(DECKED_BEATS.badge).toBeGreaterThan(DECKED_BEATS.fill);
    expect(DECKED_BEATS.fling[0]).toBe(DECKED_BEATS.fill);
  });
});

describe("its shapes", () => {
  it("flings each hand card a different way, spinning, capped", () => {
    const f = deckedFlings("bob", 7);
    expect(f).toHaveLength(7);
    expect(deckedFlings("bob", 7)).toEqual(f);
    const angles = f.map((c) => Math.atan2(c.dy, c.dx));
    expect(new Set(angles.map((a) => a.toFixed(2))).size).toBe(7);
    for (const c of f) {
      expect(Math.abs(c.spin)).toBeGreaterThanOrEqual(360);
      expect(c.reach).toBeGreaterThan(0.7);
    }
    expect(deckedFlings("bob", 40)).toHaveLength(DECKED_FLING_CAP);
    expect(deckedFlings("bob", 0)).toEqual([]);
  });

  it("raises wisps from across the pile, seeded", () => {
    const w = deckedWisps("bob");
    expect(w.length).toBeGreaterThan(4);
    expect(deckedWisps("bob")).toEqual(w);
    for (const x of w) {
      expect(x.x).toBeGreaterThanOrEqual(0);
      expect(x.x).toBeLessThanOrEqual(1);
      expect(x.rise).toBeGreaterThan(1);
    }
  });

  it("starts the skull at the pile's size and place, and fills the mat at the fill beat", () => {
    const lib = { x: 20, y: 300, w: 60, h: 84 };
    const mat = { w: 600, h: 400 };
    const skull = Math.min(mat.w, mat.h / 1.2) * 0.96;
    const k = skullKeyframes({ lib, mat, skull });
    // From the mat's centre to the pile's: (50 - 300, 342 - 200).
    expect(k[0].transform).toBe(
      `translate(-250px, 142px) scale(${Math.round((84 / (skull * 1.2)) * 100) / 100}) rotate(0deg)`,
    );
    expect(k[0].opacity).toBe(0);
    const full = k.find((f) => f.transform === "translate(0px, 0px) scale(1) rotate(0deg)");
    expect(full).toBeTruthy();
    const [r0] = DECKED_BEATS.rise;
    expect(full!.offset).toBeCloseTo((DECKED_BEATS.fill - r0) / (1 - r0));
    // It sways, then dissolves upward with a blur.
    expect(k.some((f) => String(f.transform).includes("rotate(-5deg)"))).toBe(true);
    const last = k[k.length - 1];
    expect(last.opacity).toBe(0);
    expect(last.filter).toBe("blur(12px)");
    expect(String(last.transform)).toMatch(/translate\(0px, -\d/);
  });
});
