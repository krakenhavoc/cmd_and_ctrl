// ADR 0209 §1 and §3 (#4396): which cue a log entry is, what a board wipe
// is, the three motion gates and the length caps. The rendered half is in
// tableEffects.render.test.ts.

import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

import {
  BURST_CRACKS,
  BURST_SHARDS,
  EFFECT_BASE_MS,
  EFFECT_CAP_MS,
  EFFECT_SHORT_CAP_MS,
  EFFECT_SOUND,
  EFFECT_STILL_HOLD_MS,
  WIPE_CARD_MS,
  WIPE_CARD_VOLUME,
  WIPE_MIN,
  WIPE_SWEEP_MS,
  WIPE_TOTAL_MS,
  counterBurst,
  cueMs,
  cueSound,
  cueTiming,
  effectLate,
  effectMotion,
  effectsScheduled,
  exitKind,
  isEffectsMode,
  planEffects,
  slashGeometry,
  soundBatch,
  wipeTiming,
  type EffectKind,
  type EffectMotion,
  type PlanContext,
} from "./tableEffects";
import { SOUND_CATEGORY, SOUND_MANIFEST } from "./sounds";
import type { LogEvent } from "./protocol";

const ctx = (over: Partial<PlanContext> = {}): PlanContext => ({
  isPlaneswalker: () => false,
  combatMotion: false,
  ...over,
});

let seq = 100;
const zone = (
  id: string,
  newZone: string,
  extra: Partial<LogEvent> = {},
  oldZone = "battlefield",
): LogEvent => ({
  seq: ++seq,
  kind: "zone",
  turn: 3,
  seat: -1,
  card_id: id,
  old_zone: oldZone,
  new_zone: newZone,
  text: "",
  ...extra,
});
const sacrifice = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
  seq: ++seq,
  kind: "sacrifice",
  turn: 3,
  seat: 0,
  card_id: id,
  text: "",
  ...extra,
});
const step = (name: string): LogEvent => ({
  seq: ++seq,
  kind: "step",
  step: name,
  turn: 3,
  seat: 0,
  text: name,
});

function plan(entries: LogEvent[], c: PlanContext = ctx()) {
  return planEffects(entries, entries, c);
}
const kinds = (entries: LogEvent[], c?: PlanContext) => plan(entries, c).cues.map((q) => q.kind);

describe("cue selection", () => {
  it("a destroyed permanent is a destroy", () => {
    expect(kinds([zone("a", "graveyard", { cause: "destroy" })])).toEqual(["destroy"]);
  });

  it("a death by state-based action is the existing crumble, with no new sound", () => {
    const k = kinds([zone("a", "graveyard", { cause: "sba" })]);
    expect(k).toEqual(["dies"]);
    expect(EFFECT_SOUND.dies).toBeNull();
    expect(cueSound({ kind: "dies", weight: "full" })).toBeNull();
  });

  it("an exile effect is an exile, and so is any move to exile", () => {
    expect(kinds([zone("a", "exile", { cause: "exile_effect" })])).toEqual(["exile"]);
  });

  it("reads new_zone before cause: a destroy replaced by exile is an exile, never a death", () => {
    const k = kinds([zone("a", "exile", { cause: "destroy" })]);
    expect(k).toEqual(["exile"]);
    // And a sacrifice that was exiled instead is still the exile.
    const s = [sacrifice("b"), zone("b", "exile", { cause: "sacrifice" })];
    expect(kinds(s)).toEqual(["exile"]);
  });

  it("a sacrifice entry is a sacrifice, once", () => {
    expect(kinds([sacrifice("a")])).toEqual(["sacrifice"]);
  });

  it("a card that went to a hand, or whose cause is bounce, is a bounce", () => {
    expect(kinds([zone("a", "hand", { cause: "bounce" })])).toEqual(["bounce"]);
    expect(kinds([zone("a", "hand", { cause: "other" })])).toEqual(["bounce"]);
  });

  it("a counter entry is a counterspell on its stack item", () => {
    const e: LogEvent = {
      seq: ++seq,
      kind: "counter",
      seat: 1,
      target: "bolt",
      stack_item_id: "item-1",
      text: "",
    };
    const q = plan([e]).cues[0];
    expect(q.kind).toBe("counter");
    expect(q.stackItemID).toBe("item-1");
    expect(q.cardID).toBe("bolt");
  });

  it("a counter entry from an old server falls back to the target as the item", () => {
    const e: LogEvent = { seq: ++seq, kind: "counter", seat: 1, target: "bolt", text: "" };
    expect(plan([e]).cues[0].stackItemID).toBe("bolt");
  });

  it("a battlefield exit to a library or the command zone has no cue", () => {
    expect(kinds([zone("a", "library", { cause: "other" })])).toEqual([]);
    expect(kinds([zone("a", "command", { cause: "other" })])).toEqual([]);
  });

  it("an entry that is not a battlefield exit is none", () => {
    expect(kinds([zone("a", "graveyard", {}, "hand")])).toEqual([]);
    expect(kinds([zone("a", "stack", {}, "hand")])).toEqual([]);
  });

  it("an old server's exit with no cause is a death by default", () => {
    expect(kinds([zone("a", "graveyard")])).toEqual(["dies"]);
  });

  it("a planeswalker that enters or dies gets its own cues, and an exiled one stays an exile", () => {
    const pw = ctx({ isPlaneswalker: (id) => id === "jace" });
    expect(kinds([zone("jace", "battlefield", {}, "stack")], pw)).toEqual(["pw_enter"]);
    expect(kinds([zone("bear", "battlefield", {}, "stack")], pw)).toEqual([]);
    expect(kinds([zone("jace", "graveyard", { cause: "destroy" })], pw)).toEqual(["pw_leave"]);
    expect(kinds([zone("jace", "graveyard", { cause: "sba" })], pw)).toEqual(["pw_leave"]);
    expect(kinds([zone("jace", "exile", { cause: "exile_effect" })], pw)).toEqual(["exile"]);
    expect(kinds([zone("jace", "hand", { cause: "bounce" })], pw)).toEqual(["bounce"]);
  });

  it("does not draw combat's deaths a second time while the strike layer is playing", () => {
    const entries = [
      step("combat_damage"),
      zone("a", "graveyard", { cause: "sba" }),
      zone("b", "graveyard", { cause: "destroy" }),
    ];
    // A destroy in the damage step is still ours; the crumble is the strike layer's.
    expect(kinds(entries, ctx({ combatMotion: true }))).toEqual(["destroy"]);
    // With combat motion off nobody crumbles it, so we do.
    expect(kinds(entries, ctx({ combatMotion: false }))).toEqual(["dies", "destroy"]);
  });

  it("a state-based death outside a combat damage step is ours with strikes on", () => {
    const entries = [step("main1"), zone("a", "graveyard", { cause: "sba" })];
    expect(kinds(entries, ctx({ combatMotion: true }))).toEqual(["dies"]);
  });

  it("only fresh entries cue", () => {
    const old = zone("a", "graveyard", { cause: "destroy" });
    const next = zone("b", "graveyard", { cause: "destroy" });
    const p = planEffects([next], [old, next], ctx());
    expect(p.cues.map((q) => q.cardID)).toEqual(["b"]);
  });

  it("exitKind reads the zone entry's own new_zone first", () => {
    const z = zone("a", "exile", { cause: "destroy" });
    expect(exitKind({ zone: z, sacrificed: false, step: "main1" }, ctx())).toBe("exile");
  });
});

describe("board wipes", () => {
  const wipe = (n: number, over: Partial<LogEvent> = {}, newZone = "graveyard") =>
    Array.from({ length: n }, (_, i) =>
      zone(`c${i}`, newZone, { cause: "destroy", batch: 7, source: "damnation", ...over }),
    );

  it("three permanents in one batch from one source are one wipe", () => {
    const p = plan(wipe(WIPE_MIN));
    expect(p.wipes).toHaveLength(1);
    expect(p.wipes[0]).toMatchObject({ batch: 7, source: "damnation", count: 3, tone: "dark" });
    expect(p.cues).toHaveLength(3);
    for (const q of p.cues) expect(q.weight).toBe("light");
  });

  it("two permanents are not a wipe", () => {
    const p = plan(wipe(2));
    expect(p.wipes).toHaveLength(0);
    expect(p.cues.every((q) => q.weight === "full")).toBe(true);
  });

  it("three in one batch from no source, or from different sources, are not a wipe", () => {
    expect(plan(wipe(4, { source: undefined })).wipes).toHaveLength(0);
    expect(plan(wipe(4, { source: "" })).wipes).toHaveLength(0);
    const mixed = [
      zone("a", "graveyard", { cause: "destroy", batch: 7, source: "x" }),
      zone("b", "graveyard", { cause: "destroy", batch: 7, source: "y" }),
      zone("c", "graveyard", { cause: "destroy", batch: 7, source: "z" }),
    ];
    expect(plan(mixed).wipes).toHaveLength(0);
  });

  it("three in different batches are not a wipe", () => {
    const sep = [
      zone("a", "graveyard", { cause: "destroy", batch: 1, source: "x" }),
      zone("b", "graveyard", { cause: "destroy", batch: 2, source: "x" }),
      zone("c", "graveyard", { cause: "destroy", batch: 3, source: "x" }),
    ];
    expect(plan(sep).wipes).toHaveLength(0);
  });

  it("a sacrifice wipe is the same over sacrifice entries", () => {
    const s = Array.from({ length: 4 }, (_, i) =>
      sacrifice(`s${i}`, { batch: 9, source: "tribute" }),
    );
    const p = plan(s);
    expect(p.wipes).toHaveLength(1);
    expect(p.cues.map((q) => q.kind)).toEqual(["sacrifice", "sacrifice", "sacrifice", "sacrifice"]);
  });

  it("an exile wipe is cold, and a mass bounce is cool", () => {
    expect(plan(wipe(3, { cause: "exile_effect" }, "exile")).wipes[0].tone).toBe("cold");
    expect(plan(wipe(3, { cause: "bounce" }, "hand")).wipes[0].tone).toBe("cool");
  });

  it("staggers the per-card cues after the sweep begins, inside the cap", () => {
    const p = plan(wipe(12));
    const delays = p.cues.map((q) => q.delayMs);
    expect(delays[0]).toBeGreaterThan(0);
    expect([...delays].sort((a, b) => a - b)).toEqual(delays);
    expect(delays[delays.length - 1] + WIPE_CARD_MS).toBeLessThanOrEqual(WIPE_TOTAL_MS);
  });

  it("a sba death in a wipe still counts toward it", () => {
    const entries = [
      zone("a", "graveyard", { cause: "destroy", batch: 5, source: "wipe" }),
      zone("b", "graveyard", { cause: "sba", batch: 5, source: "wipe" }),
      zone("c", "graveyard", { cause: "sba", batch: 5, source: "wipe" }),
    ];
    expect(plan(entries).wipes[0].count).toBe(3);
  });
});

describe("the sound", () => {
  it("each cue's sound is an effects-category manifest entry with takes", () => {
    for (const k of Object.keys(EFFECT_SOUND) as EffectKind[]) {
      const name = EFFECT_SOUND[k];
      if (!name) continue;
      expect(SOUND_CATEGORY[name]).toBe("effects");
      expect(SOUND_MANIFEST[name].length).toBeGreaterThan(0);
    }
    expect(SOUND_CATEGORY.board_wipe).toBe("effects");
  });

  it("a wipe's per-card sound is quieter", () => {
    expect(cueSound({ kind: "destroy", weight: "light" })?.volume).toBe(WIPE_CARD_VOLUME);
    expect(cueSound({ kind: "destroy", weight: "full" })?.volume).toBe(1);
  });

  it("caps one sound per batch per game, and leaves an unbatched one uncapped", () => {
    expect(soundBatch("g1", 7)).toBe("g1:7");
    expect(soundBatch("g2", 7)).not.toBe(soundBatch("g1", 7));
    expect(soundBatch("g1", null)).toBeUndefined();
  });
});

describe("the motion gates", () => {
  const base = { enabled: true, effects: "full" as const, reduceMotion: false };

  it("plays in full and short with every signal clear", () => {
    expect(effectMotion(base)).toBe("full");
    expect(effectMotion({ ...base, effects: "short" })).toBe("short");
  });

  it("is still when effects are off", () => {
    expect(effectMotion({ ...base, effects: "off" })).toBe("still");
  });

  it("is still under each of the three reduced-motion signals", () => {
    // 1. accessibility.reduceMotion
    expect(effectMotion({ ...base, reduceMotion: true })).toBe("still");
    // 2. animations.enabled
    expect(effectMotion({ ...base, enabled: false })).toBe("still");
    // 3. the OS signal
    expect(effectMotion({ ...base, osReduceMotion: true })).toBe("still");
    expect(effectMotion({ ...base, osReduceMotion: false })).toBe("full");
  });

  it("validates the stored mode", () => {
    expect(isEffectsMode("full")).toBe(true);
    expect(isEffectsMode("short")).toBe(true);
    expect(isEffectsMode("off")).toBe(true);
    expect(isEffectsMode("loud")).toBe(false);
    expect(isEffectsMode(undefined)).toBe(false);
  });

  it("draws no motion in a hidden tab, and drops a late cue", () => {
    expect(effectsScheduled("visible")).toBe(true);
    expect(effectsScheduled(undefined)).toBe(true);
    expect(effectsScheduled("hidden")).toBe(false);
    expect(effectLate(1000, 0, 1500)).toBe(false);
    expect(effectLate(1000, 0, 2500)).toBe(true);
    // A staggered cue is late against its own start, not the frame's.
    expect(effectLate(1000, 600, 2500)).toBe(false);
  });
});

describe("length caps", () => {
  const speeds = [0.5, 1, 1.5, 2];
  const modes: EffectMotion[] = ["full", "short"];
  const allKinds = Object.keys(EFFECT_BASE_MS) as EffectKind[];

  it("no base cue is longer than the 1.5 s cap", () => {
    for (const k of allKinds) expect(EFFECT_BASE_MS[k]).toBeLessThanOrEqual(EFFECT_CAP_MS);
    expect(WIPE_SWEEP_MS).toBeLessThanOrEqual(EFFECT_CAP_MS);
    expect(WIPE_TOTAL_MS).toBeLessThanOrEqual(EFFECT_CAP_MS);
  });

  it("every cue, in every mode and speed, ends within the cap scaled by speed", () => {
    for (const speed of speeds) {
      for (const mode of modes) {
        const cap = (mode === "short" ? EFFECT_SHORT_CAP_MS : EFFECT_CAP_MS) * speed;
        for (const kind of allKinds) {
          const full = cueTiming({ kind, weight: "full", delayMs: 0 }, mode, speed);
          expect(full.delayMs + full.ms).toBeLessThanOrEqual(cap + 1e-6);
          // The last per-card cue of the biggest wipe, staggered.
          const last = plan(
            Array.from({ length: 12 }, (_, i) =>
              zone(`w${i}`, "graveyard", { cause: "destroy", batch: 3, source: "wipe" }),
            ),
          ).cues.at(-1)!;
          const light = cueTiming({ kind, weight: "light", delayMs: last.delayMs }, mode, speed);
          expect(light.delayMs + light.ms).toBeLessThanOrEqual(cap + 1e-6);
        }
        expect(wipeTiming(mode, speed).ms).toBeLessThanOrEqual(cap);
      }
    }
  });

  it("scales with animations.speed", () => {
    expect(cueMs(900, "full", 2)).toBe(1800);
    expect(cueMs(900, "full", 0.5)).toBe(450);
    expect(cueMs(900, "full", 0)).toBe(900);
    expect(cueMs(900, "full", Number.NaN)).toBe(900);
  });

  it("Short cuts a long cue to 0.8 s and leaves a shorter one alone", () => {
    expect(cueMs(1100, "short", 1)).toBe(EFFECT_SHORT_CAP_MS);
    expect(cueMs(420, "short", 1)).toBe(420);
  });

  it("a still cue is the badge's fixed reading time, which speed does not scale", () => {
    expect(cueMs(900, "still", 2)).toBe(EFFECT_STILL_HOLD_MS);
    expect(wipeTiming("still", 0.5).ms).toBe(EFFECT_STILL_HOLD_MS);
    expect(cueTiming({ kind: "destroy", weight: "full", delayMs: 0 }, "still", 2)).toEqual({
      delayMs: 0,
      ms: EFFECT_STILL_HOLD_MS,
    });
  });
});

describe("shapes", () => {
  it("the slash runs along the card's diagonal, a little longer than it", () => {
    const s = slashGeometry(80, 112);
    expect(s.length).toBeGreaterThan(Math.hypot(80, 112));
    expect(s.angle).toBeCloseTo((Math.atan2(112, 80) * 180) / Math.PI, 1);
  });

  it("the counterspell burst is deterministic by seed, and made of cracks and shards", () => {
    const a = counterBurst("101", 120, 160);
    const b = counterBurst("101", 120, 160);
    const c = counterBurst("102", 120, 160);
    expect(a).toEqual(b);
    expect(a).not.toEqual(c);
    expect(a.cracks).toHaveLength(BURST_CRACKS);
    expect(a.shards).toHaveLength(BURST_SHARDS);
    for (const d of a.cracks) expect(d.startsWith("M50 50")).toBe(true);
  });
});

describe("the layer", () => {
  const src = readFileSync(
    new URL("./components/board/TableEffects.svelte", import.meta.url),
    "utf8",
  );
  it("sits at z 57 on the ladder, takes no pointer events and is hidden from assistive tech", () => {
    expect(src).toMatch(/z-index:\s*57;/);
    expect(src).toMatch(/pointer-events:\s*none;/);
    expect(src).toMatch(/aria-hidden="true"/);
  });
  it("carries its own prefers-reduced-motion block", () => {
    expect(src).toMatch(/@media \(prefers-reduced-motion: reduce\)/);
  });
});
