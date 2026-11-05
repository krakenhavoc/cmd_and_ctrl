// ADR 0209 §1 and §3 (#4396, #4416): which cue a log entry is, what a
// board wipe is, the three motion gates, the length caps, a busy table and
// the shapes the cues draw. The rendered half is in
// tableEffects.render.test.ts.

import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

import {
  BURST_CRACKS,
  BURST_SHARDS,
  BUSY_MIN,
  BUSY_WINDOW_MS,
  BusyClock,
  COUNTER_SOUND,
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
  WIPE_VORTEX_MS,
  SACRIFICE_KNIFE_SOUND,
  SHATTER_BREAK,
  TOKEN_SPREAD_MS,
  TOKEN_STAGGER_MS,
  attachKind,
  busyScale,
  counterBurst,
  counterTone,
  cueMs,
  cueSound,
  cueTiming,
  effectLate,
  effectMotion,
  effectsScheduled,
  exitKind,
  FIZZLE_MOTES,
  fizzleCue,
  fizzleMotes,
  headlines,
  isEffectsMode,
  planEffects,
  sacrificeBeats,
  shatterPlan,
  soundBatch,
  spiralPath,
  swirlArms,
  voidMotes,
  windCuts,
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

  it("marks a destroy replaced by exile, so it crumbles a little as it goes (D4)", () => {
    const [replaced] = plan([zone("a", "exile", { cause: "destroy" })]).cues;
    expect(replaced.fromDestroy).toBe(true);
    const [plain] = plan([zone("b", "exile", { cause: "exile_effect" })]).cues;
    expect(plain.fromDestroy).toBeUndefined();
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

describe("a spell or ability fizzles (#4450)", () => {
  const fizzle = (extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "fizzle",
    seat: 1,
    card_id: "bolt",
    stack_item_id: "item-1",
    batch: 12,
    text: "Lightning Bolt fizzled: no legal targets",
    ...extra,
  });

  it("is a fizzle cue on its stack item, with its batch", () => {
    const [q] = plan([fizzle()]).cues;
    expect(q.kind).toBe("fizzle");
    expect(q.stackItemID).toBe("item-1");
    expect(q.cardID).toBe("bolt");
    expect(q.batch).toBe(12);
    expect(q.weight).toBe("full");
    expect(q.delayMs).toBe(0);
  });

  it("an entry from an old server falls back to the spell's card as the item", () => {
    expect(plan([fizzle({ stack_item_id: undefined })]).cues[0].stackItemID).toBe("bolt");
  });

  it("an optional trigger that lapsed (call condition, CR 603.4) is no cue", () => {
    expect(kinds([fizzle({ call: "condition" })])).toEqual([]);
    expect(fizzleCue(fizzle({ call: "condition" }))).toBeNull();
  });

  it("an entry of another kind is never a fizzle", () => {
    expect(fizzleCue({ ...fizzle(), kind: "resolve" })).toBeNull();
    expect(kinds([{ ...fizzle(), kind: "resolve" }])).toEqual([]);
  });

  it("plays the fizzle sound, an effects-category entry with no takes until its file lands", () => {
    expect(EFFECT_SOUND.fizzle).toBe("fizzle");
    expect(cueSound({ kind: "fizzle", weight: "full" })).toEqual({ name: "fizzle", volume: 1 });
    expect(SOUND_CATEGORY.fizzle).toBe("effects");
    expect(SOUND_MANIFEST.fizzle).toEqual([]);
  });

  it("is a soft dissolve inside its sound's 0.8 s, and the same in Short", () => {
    expect(EFFECT_BASE_MS.fizzle).toBe(800);
    expect(cueTiming(plan([fizzle()]).cues[0], "full", 1)).toEqual({ delayMs: 0, ms: 800 });
    expect(cueTiming(plan([fizzle()]).cues[0], "short", 1)).toEqual({ delayMs: 0, ms: 800 });
    expect(cueTiming(plan([fizzle()]).cues[0], "still", 1).ms).toBe(EFFECT_STILL_HOLD_MS);
  });

  it("counts once on a busy table", () => {
    expect(headlines(plan([fizzle()]))).toBe(1);
  });

  it("comes apart into seeded specks over its box, top first, each drifting up", () => {
    const a = fizzleMotes("7", 150, 210);
    expect(a).toEqual(fizzleMotes("7", 150, 210));
    expect(a).not.toEqual(fizzleMotes("8", 150, 210));
    expect(a).toHaveLength(FIZZLE_MOTES);
    for (const m of a) {
      expect(m.x).toBeGreaterThanOrEqual(0);
      expect(m.x).toBeLessThanOrEqual(100);
      expect(m.y).toBeGreaterThanOrEqual(0);
      expect(m.y).toBeLessThanOrEqual(100);
      expect(m.dy).toBeLessThan(0);
      expect(m.at).toBeGreaterThan(0);
      expect(m.at).toBeLessThan(1);
    }
    // The higher half lets go before the lower half, on average.
    const byY = [...a].sort((p, q) => p.y - q.y);
    const mean = (ms: typeof a) => ms.reduce((t, m) => t + m.at, 0) / ms.length;
    expect(mean(byY.slice(0, 7))).toBeLessThan(mean(byY.slice(7)));
  });
});

describe("a token is created (#4441)", () => {
  // The engine logs a created token as a `token` entry (it came from no
  // zone, so there is no `zone` entry for it), with its batch.
  const token = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "token",
    turn: 3,
    seat: 0,
    card_id: id,
    text: "",
    ...extra,
  });

  it("is a token_create cue on the token, with its own sound", () => {
    const p = plan([token("t1", { batch: 7 })]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.batch])).toEqual([["token_create", "t1", 7]]);
    expect(p.wipes).toEqual([]);
    expect(EFFECT_SOUND.token_create).toBe("token_create");
    expect(cueSound(p.cues[0])).toEqual({ name: "token_create", volume: 1 });
    expect(EFFECT_BASE_MS.token_create).toBe(400);
  });

  it("a batch of tokens is one sound, with the pop-ins a stagger apart", () => {
    const p = plan([
      token("t1", { batch: 7 }),
      token("t2", { batch: 7 }),
      token("t3", { batch: 7 }),
    ]);
    expect(p.cues.map((q) => q.delayMs)).toEqual([0, TOKEN_STAGGER_MS, 2 * TOKEN_STAGGER_MS]);
    expect(p.cues.map((q) => q.follows === true)).toEqual([false, true, true]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual(["token_create", null, null]);
    // Every pop-in is a full cue, not a wipe's lighter one.
    for (const q of p.cues) expect(q.weight).toBe("full");
    // And the batch is one moment on a busy table.
    expect(headlines(p)).toBe(1);
  });

  it("two batches are two moments, each with its sound", () => {
    const p = plan([token("t1", { batch: 7 }), token("t2", { batch: 8 })]);
    expect(p.cues.map((q) => q.delayMs)).toEqual([0, 0]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual(["token_create", "token_create"]);
    expect(headlines(p)).toBe(2);
  });

  it("a big batch spreads its pop-ins over TOKEN_SPREAD_MS at most", () => {
    const p = plan(Array.from({ length: 20 }, (_, i) => token(`t${i}`, { batch: 9 })));
    const delays = p.cues.map((q) => q.delayMs);
    for (let i = 1; i < delays.length; i++) expect(delays[i]).toBeGreaterThan(delays[i - 1]);
    expect(delays.at(-1)).toBeCloseTo(TOKEN_SPREAD_MS, 6);
  });

  it("keeps its stagger in its timing, and the whole batch fits the cap in every mode", () => {
    const p = plan(Array.from({ length: 20 }, (_, i) => token(`t${i}`, { batch: 9 })));
    const last = p.cues.at(-1)!;
    for (const speed of [0.5, 1, 2]) {
      for (const mode of ["full", "short"] as EffectMotion[]) {
        const t = cueTiming(last, mode, speed);
        expect(t.delayMs).toBeCloseTo(TOKEN_SPREAD_MS * speed, 6);
        expect(t.ms).toBeCloseTo(EFFECT_BASE_MS.token_create * speed, 6);
        const cap = (mode === "short" ? EFFECT_SHORT_CAP_MS : EFFECT_CAP_MS) * speed;
        expect(t.delayMs + t.ms).toBeLessThanOrEqual(cap + 1e-6);
      }
    }
    // Still: no motion, no wait.
    expect(cueTiming(last, "still", 1)).toEqual({ delayMs: 0, ms: EFFECT_STILL_HOLD_MS });
  });

  it("is not an exit, and a token's later death is its own cue", () => {
    const k = kinds([
      token("t1", { batch: 3 }),
      zone("t1", "graveyard", { cause: "sba", batch: 4 }),
    ]);
    expect(k).toEqual(["token_create", "dies"]);
  });
});

describe("counters are put on a permanent (#4442)", () => {
  // The engine logs each kind's new total on a card as a `counters`
  // entry; `delta` is the signed change, which tells a placement from a
  // removal.
  const counters = (
    id: string,
    label: string,
    delta: number | undefined,
    extra: Partial<LogEvent> = {},
  ): LogEvent => ({
    seq: ++seq,
    kind: "counters",
    turn: 3,
    seat: -1,
    card_id: id,
    label,
    amount: Math.max(0, delta ?? 1),
    delta,
    text: "",
    ...extra,
  });

  it("+1/+1 counters are a counter_add cue on the card, -1/-1 a counter_remove one", () => {
    const p = plan([
      counters("a", "+1/+1", 2, { batch: 4 }),
      counters("b", "-1/-1", 1, { batch: 5 }),
    ]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.counterTone, q.batch])).toEqual([
      ["counters", "a", "plus", 4],
      ["counters", "b", "minus", 5],
    ]);
    expect(p.cues.map((q) => cueSound(q))).toEqual([
      { name: "counter_add", volume: 1 },
      { name: "counter_remove", volume: 1 },
    ]);
    expect(COUNTER_SOUND).toEqual({ plus: "counter_add", minus: "counter_remove", other: null });
    for (const name of [COUNTER_SOUND.plus!, COUNTER_SOUND.minus!]) {
      expect(SOUND_CATEGORY[name]).toBe("effects");
    }
    expect(EFFECT_BASE_MS.counters).toBe(450);
    expect(headlines(p)).toBe(2);
  });

  it("another kind is a pulse with no sound", () => {
    const p = plan([counters("a", "charge", 1, { batch: 4 })]);
    expect(p.cues.map((q) => [q.kind, q.counterTone, q.counterLabel])).toEqual([
      ["counters", "other", "charge"],
    ]);
    expect(cueSound(p.cues[0])).toBeNull();
    expect(counterTone("oil")).toBe("other");
  });

  it("a batch is one sound: a proliferate over many permanents plays once", () => {
    const p = plan([
      counters("a", "charge", 1, { batch: 9 }),
      counters("b", "+1/+1", 1, { batch: 9 }),
      counters("c", "+1/+1", 1, { batch: 9 }),
      counters("d", "-1/-1", 1, { batch: 9 }),
    ]);
    expect(p.cues.map((q) => q.cardID)).toEqual(["a", "b", "c", "d"]);
    // The first cue with a sound leads; every other one follows it.
    expect(p.cues.map((q) => q.follows === true)).toEqual([true, false, true, true]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual([null, "counter_add", null, null]);
    // The pulses play together, and the batch is one moment.
    expect(p.cues.map((q) => q.delayMs)).toEqual([0, 0, 0, 0]);
    expect(headlines(p)).toBe(1);
  });

  it("two batches are two sounds", () => {
    const p = plan([
      counters("a", "+1/+1", 1, { batch: 4 }),
      counters("a", "+1/+1", 1, { batch: 5 }),
    ]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual(["counter_add", "counter_add"]);
  });

  it("one card's entries of one kind in one batch are one pulse", () => {
    const p = plan([
      counters("a", "+1/+1", 1, { batch: 4 }),
      counters("a", "+1/+1", 1, { batch: 4 }),
    ]);
    expect(p.cues).toHaveLength(1);
  });

  it("a removal, the CR 704.5q cancel, a redacted entry and one with no delta are not cues", () => {
    const k = kinds([
      counters("a", "+1/+1", -1, { batch: 4 }),
      counters("a", "-1/-1", -1, { batch: 4 }),
      // A card this viewer may not identify: no kind, no count, no change.
      counters("b", "", undefined, { label: undefined, amount: undefined, batch: 5 }),
      // A server that predates `delta`.
      counters("c", "+1/+1", undefined, { batch: 6 }),
    ]);
    expect(k).toEqual([]);
  });

  it("fits the Short cap, and still is no motion", () => {
    const [cue] = plan([counters("a", "+1/+1", 1, { batch: 4 })]).cues;
    expect(cueTiming(cue, "short", 1).ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS);
    expect(cueTiming(cue, "full", 2)).toEqual({ delayMs: 0, ms: 900 });
    expect(cueTiming(cue, "still", 1)).toEqual({ delayMs: 0, ms: EFFECT_STILL_HOLD_MS });
  });
});

describe("a commander is cast from the command zone, or goes home (#4443)", () => {
  // The engine logs a cast with its origin in `old_zone` when that is not
  // a hand, and every move into the command zone as a `zone` entry: the
  // state-based action's (CR 903.9a, from a graveyard or exile) and the
  // replacement's (CR 903.9b, instead of a hand or a library).
  const cast = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "cast",
    turn: 3,
    seat: 0,
    card_id: id,
    old_zone: "command",
    text: "",
    ...extra,
  });
  const commanders = (...ids: string[]) => ctx({ isCommander: (id) => ids.includes(id) });

  it("a cast from the command zone is a commander_cast on its stack item, with its own sound", () => {
    const p = plan([cast("cmdr", { batch: 4 })]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.stackItemID, q.batch])).toEqual([
      ["commander_cast", "cmdr", "cmdr", 4],
    ]);
    expect(EFFECT_SOUND.commander_cast).toBe("commander_cast");
    expect(cueSound(p.cues[0])).toEqual({ name: "commander_cast", volume: 1 });
    expect(EFFECT_BASE_MS.commander_cast).toBe(800);
  });

  it("a cast from anywhere else is none", () => {
    expect(kinds([cast("a", { old_zone: undefined })])).toEqual([]);
    expect(kinds([cast("a", { old_zone: "graveyard" })])).toEqual([]);
    expect(kinds([cast("a", { old_zone: "exile" })])).toEqual([]);
  });

  it("a commander put into the command zone from anywhere is a commander_return, from that zone", () => {
    for (const from of ["graveyard", "exile", "battlefield", "hand", "library", "stack"]) {
      const p = plan([zone("cmdr", "command", { batch: 5 }, from)], commanders("cmdr"));
      expect(p.cues.map((q) => [q.kind, q.cardID, q.commanderFrom])).toEqual([
        ["commander_return", "cmdr", from],
      ]);
      expect(cueSound(p.cues[0])).toEqual({ name: "commander_return", volume: 1 });
    }
    expect(EFFECT_BASE_MS.commander_return).toBe(800);
  });

  it("a bounce or a tuck its owner sent home instead (CR 903.9b) is the return, never a bounce", () => {
    const p = plan([zone("cmdr", "command", { cause: "bounce", batch: 6 })], commanders("cmdr"));
    expect(p.cues.map((q) => q.kind)).toEqual(["commander_return"]);
  });

  it("a card put into the command zone that is not a commander is no return", () => {
    // An Attraction going to its junkyard (CR 717.6a), or no context at all.
    expect(kinds([zone("attr", "command", {}, "battlefield")], commanders("cmdr"))).toEqual([]);
    expect(kinds([zone("cmdr", "command", {}, "graveyard")])).toEqual([]);
  });

  it("a commander that dies, then goes home, is two cues: its death, then its return", () => {
    const k = kinds(
      [
        zone("cmdr", "graveyard", { cause: "destroy", batch: 7 }),
        zone("cmdr", "command", { batch: 8 }, "graveyard"),
      ],
      commanders("cmdr"),
    );
    expect(k).toEqual(["destroy", "commander_return"]);
  });

  it("two commanders of one batch are one moment: one sound, one headline", () => {
    const p = plan(
      [
        zone("a", "command", { cause: "bounce", batch: 9 }),
        zone("b", "command", { cause: "bounce", batch: 9 }),
      ],
      commanders("a", "b"),
    );
    expect(p.cues.map((q) => q.follows === true)).toEqual([false, true]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual(["commander_return", null]);
    expect(headlines(p)).toBe(1);
    // Different batches, or a cast and a return, are each their own.
    const q = plan(
      [cast("a", { batch: 10 }), zone("b", "command", { batch: 10 }, "graveyard")],
      commanders("a", "b"),
    );
    expect(q.cues.map((c) => cueSound(c)?.name)).toEqual(["commander_cast", "commander_return"]);
    expect(headlines(q)).toBe(2);
  });

  it("plays at once, the same in Full and Short, and still is the hold", () => {
    for (const kind of ["commander_cast", "commander_return"] as EffectKind[]) {
      for (const mode of ["full", "short"] as EffectMotion[]) {
        expect(cueTiming({ kind, weight: "full", delayMs: 0 }, mode, 1)).toEqual({
          delayMs: 0,
          ms: 800,
        });
      }
      expect(cueTiming({ kind, weight: "full", delayMs: 0 }, "still", 1)).toEqual({
        delayMs: 0,
        ms: EFFECT_STILL_HOLD_MS,
      });
    }
  });
});

describe("an Equipment is attached, or an Aura enchants a permanent (#4444)", () => {
  // The engine logs every attach as an `attach` entry: `card_id` the
  // attachment, `target` the host card (`target_seat` a host player) and
  // `label` what the attachment is.
  const attach = (id: string, host: string | null, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "attach",
    turn: 3,
    seat: 0,
    card_id: id,
    ...(host ? { target: host } : {}),
    label: "Equipment",
    text: "",
    ...extra,
  });
  const token = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "token",
    turn: 3,
    seat: 0,
    card_id: id,
    text: "",
    ...extra,
  });

  it("an Equipment is attach_equip and an Aura attach_aura, on the attachment, naming its host", () => {
    const p = plan([
      attach("sword", "bear", { batch: 4 }),
      attach("pacifism", "ogre", { label: "Aura", batch: 5 }),
      attach("fort", "land", { label: "Fortification", batch: 6 }),
    ]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.hostID, q.batch])).toEqual([
      ["attach_equip", "sword", "bear", 4],
      ["attach_aura", "pacifism", "ogre", 5],
      ["attach_equip", "fort", "land", 6],
    ]);
    expect(cueSound(p.cues[0])).toEqual({ name: "attach_equip", volume: 1 });
    expect(cueSound(p.cues[1])).toEqual({ name: "attach_aura", volume: 1 });
    expect(EFFECT_SOUND.attach_equip).toBe("attach_equip");
    expect(EFFECT_SOUND.attach_aura).toBe("attach_aura");
    expect(EFFECT_BASE_MS.attach_equip).toBe(400);
    expect(EFFECT_BASE_MS.attach_aura).toBe(450);
  });

  it("a Curse enchanting a player, a redacted label and a permanent that is none of the three are no cue", () => {
    expect(
      kinds([attach("curse", null, { label: "Aura", target_seat: 1 } as Partial<LogEvent>)]),
    ).toEqual([]);
    expect(kinds([attach("a", "bear", { label: undefined })])).toEqual([]);
    expect(kinds([attach("a", "bear", { label: "Creature" })])).toEqual([]);
    expect(attachKind(attach("a", "bear"))).toBe("attach_equip");
    expect(attachKind(zone("a", "graveyard"))).toBeNull();
  });

  it("one kind in one batch is one moment; an Equipment and an Aura are two", () => {
    const p = plan([
      attach("s1", "bear", { batch: 9 }),
      attach("s2", "bear", { batch: 9 }),
      attach("aura", "bear", { label: "Aura", batch: 9 }),
    ]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual([
      "attach_equip",
      null,
      "attach_aura",
    ]);
    expect(headlines(p)).toBe(2);
    // Two batches are two moments.
    const q = plan([attach("s1", "bear", { batch: 10 }), attach("s1", "ogre", { batch: 11 })]);
    expect(q.cues.map((c) => cueSound(c)?.name)).toEqual(["attach_equip", "attach_equip"]);
  });

  it("a token created attached is one arrival: its attach cue plays the sound, its token cue only pops it in", () => {
    // A Role, or a token copy of an Aura spell: `token` and `attach` share
    // the batch. A plain token made beside it keeps the token sound.
    const p = plan([
      token("role", { batch: 12 }),
      attach("role", "bear", { label: "Aura", batch: 12 }),
      token("germ", { batch: 12 }),
    ]);
    expect(p.cues.map((q) => [q.kind, q.cardID, cueSound(q)?.name ?? null])).toEqual([
      ["token_create", "role", null],
      ["attach_aura", "role", "attach_aura"],
      ["token_create", "germ", "token_create"],
    ]);
    // Its token pop keeps the batch's stagger.
    expect(p.cues.filter((q) => q.kind === "token_create").map((q) => q.delayMs)).toEqual([
      0,
      TOKEN_STAGGER_MS,
    ]);
    expect(headlines(p)).toBe(2);
    // A batch of only Roles: one sound, the attach's.
    const r = plan([
      token("r1", { batch: 13 }),
      attach("r1", "a", { label: "Aura", batch: 13 }),
      token("r2", { batch: 13 }),
      attach("r2", "b", { label: "Aura", batch: 13 }),
    ]);
    expect(r.cues.map((c) => cueSound(c)?.name ?? null)).toEqual([null, "attach_aura", null, null]);
    expect(headlines(r)).toBe(1);
  });

  it("a living weapon's Germ pops in with its sound, and the Equipment slides onto it with its own", () => {
    // The Germ is the host, not the attachment: two arrivals, two sounds.
    const p = plan([token("germ", { batch: 14 }), attach("sword", "germ", { batch: 14 })]);
    expect(p.cues.map((q) => [q.kind, cueSound(q)?.name ?? null])).toEqual([
      ["token_create", "token_create"],
      ["attach_equip", "attach_equip"],
    ]);
  });

  it("plays at once, the same in Full and Short, and still is the hold", () => {
    for (const kind of ["attach_equip", "attach_aura"] as EffectKind[]) {
      for (const mode of ["full", "short"] as EffectMotion[]) {
        expect(cueTiming({ kind, weight: "full", delayMs: 0 }, mode, 1)).toEqual({
          delayMs: 0,
          ms: EFFECT_BASE_MS[kind],
        });
      }
      expect(cueTiming({ kind, weight: "full", delayMs: 0 }, "still", 1)).toEqual({
        delayMs: 0,
        ms: EFFECT_STILL_HOLD_MS,
      });
    }
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
  it("each cue's sound is an effects-category manifest entry, with takes once its files land", () => {
    for (const k of Object.keys(EFFECT_SOUND) as EffectKind[]) {
      const name = EFFECT_SOUND[k];
      if (!name) continue;
      expect(SOUND_CATEGORY[name]).toBe("effects");
    }
    // The ADR 0209 §3 cues have their files. A later cue may have none
    // yet; sounds.test.ts holds each of those to its NOT_MADE_YET list.
    const made: EffectKind[] = [
      "destroy",
      "exile",
      "sacrifice",
      "bounce",
      "counter",
      "pw_enter",
      "pw_leave",
    ];
    for (const k of made) expect(SOUND_MANIFEST[EFFECT_SOUND[k]!].length).toBeGreaterThan(0);
    expect(SOUND_CATEGORY.board_wipe).toBe("effects");
    // A sacrifice's knife (#4416) is one too.
    expect(SOUND_CATEGORY[SACRIFICE_KNIFE_SOUND]).toBe("effects");
    expect(SOUND_MANIFEST[SACRIFICE_KNIFE_SOUND].length).toBeGreaterThan(0);
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

  it("no base cue is longer than the 2 s cap", () => {
    expect(EFFECT_CAP_MS).toBe(2000);
    expect(EFFECT_SHORT_CAP_MS).toBe(1000);
    for (const k of allKinds) expect(EFFECT_BASE_MS[k]).toBeLessThanOrEqual(EFFECT_CAP_MS);
    expect(WIPE_SWEEP_MS).toBeLessThanOrEqual(EFFECT_CAP_MS);
    expect(WIPE_TOTAL_MS).toBeLessThanOrEqual(EFFECT_CAP_MS);
    expect(WIPE_VORTEX_MS).toBeLessThanOrEqual(EFFECT_CAP_MS);
  });

  it("every cue is about 1.25 to 1.5 times the first build's length (#4416)", () => {
    // The first build's lengths. A sacrifice gained an altar and a knife
    // and a counterspell holds its cracks for the stack linger, so they
    // may run longer; none may run shorter. The S80 cues (#4441 on) were
    // not in the first build: their issues give their lengths.
    const first: Partial<Record<EffectKind, number>> = {
      destroy: 900,
      exile: 1100,
      dies: 420,
      sacrifice: 800,
      bounce: 900,
      counter: 900,
      pw_enter: 900,
      pw_leave: 900,
    };
    for (const k of Object.keys(first) as EffectKind[]) {
      const r = EFFECT_BASE_MS[k] / first[k]!;
      expect(r).toBeGreaterThanOrEqual(1.25);
      if (k !== "sacrifice" && k !== "counter" && k !== "pw_leave") {
        expect(r).toBeLessThanOrEqual(1.5);
      }
    }
    expect(WIPE_SWEEP_MS / 700).toBeGreaterThanOrEqual(1.25);
    expect(WIPE_TOTAL_MS / 1350).toBeGreaterThanOrEqual(1.25);
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
        expect(wipeTiming(mode, speed, "cold").ms).toBeLessThanOrEqual(cap);
      }
    }
  });

  it("scales with animations.speed", () => {
    expect(cueMs(900, "full", 2)).toBe(1800);
    expect(cueMs(900, "full", 0.5)).toBe(450);
    expect(cueMs(900, "full", 0)).toBe(900);
    expect(cueMs(900, "full", Number.NaN)).toBe(900);
  });

  it("Short cuts a long cue to 1 s and leaves a shorter one alone", () => {
    expect(cueMs(1500, "short", 1)).toBe(EFFECT_SHORT_CAP_MS);
    expect(cueMs(560, "short", 1)).toBe(560);
  });

  it("an exile wipe's vortex stays open for the whole wipe; a destroy's cuts are quicker", () => {
    expect(wipeTiming("full", 1, "cold").ms).toBe(WIPE_VORTEX_MS);
    expect(wipeTiming("full", 1, "dark").ms).toBe(WIPE_SWEEP_MS);
    expect(WIPE_VORTEX_MS).toBeGreaterThan(WIPE_TOTAL_MS);
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

describe("a busy table (#4416)", () => {
  it("plays one cue in full and compresses as more land close together", () => {
    expect(busyScale(1)).toBe(1);
    expect(busyScale(0)).toBe(1);
    let last = 1;
    for (let n = 2; n <= 10; n++) {
      const k = busyScale(n);
      expect(k).toBeLessThanOrEqual(last);
      expect(k).toBeGreaterThanOrEqual(BUSY_MIN);
      last = k;
    }
    // Four at once play about as long as the first build's cues did.
    expect(busyScale(4)).toBeCloseTo(1 / 1.54, 2);
    expect(busyScale(20)).toBe(BUSY_MIN);
  });

  it("counts the cues of the last two seconds, and forgets the older ones", () => {
    const c = new BusyClock();
    expect(c.frame(1, 0)).toBe(1);
    expect(c.frame(1, 500)).toBe(busyScale(2));
    expect(c.frame(2, 900)).toBe(busyScale(4));
    // Past the window, the first three have gone.
    expect(c.frame(1, 900 + BUSY_WINDOW_MS)).toBe(1);
    // A frame with no cues is not compressed and adds nothing.
    expect(c.frame(0, 900 + BUSY_WINDOW_MS)).toBe(1);
    c.reset();
    expect(c.frame(1, 1000 + BUSY_WINDOW_MS)).toBe(1);
  });

  it("counts a board wipe once, not each card it took", () => {
    const p = plan(
      Array.from({ length: 6 }, (_, i) =>
        zone(`w${i}`, "graveyard", { cause: "destroy", batch: 3, source: "wipe" }),
      ),
    );
    expect(headlines(p)).toBe(1);
    expect(headlines(p, 1)).toBe(2);
    expect(headlines(plan([zone("a", "graveyard", { cause: "destroy" })]))).toBe(1);
  });
});

describe("shapes", () => {
  it("the shatter breaks from an impact near the centre, inner pieces first (D1)", () => {
    const s = shatterPlan("card-1", 80, 112);
    expect(s).toEqual(shatterPlan("card-1", 80, 112));
    expect(s).not.toEqual(shatterPlan("card-2", 80, 112));
    expect(Math.abs(s.ix - 40)).toBeLessThan(80 * 0.07);
    expect(Math.abs(s.iy - 56)).toBeLessThan(112 * 0.07);
    expect(s.spokes.length).toBeGreaterThan(6);
    expect(s.rings.length).toBeGreaterThan(0);
    const byRing = new Map<number, number[]>();
    for (const sh of s.shards) {
      expect(sh.clip.startsWith("polygon(")).toBe(true);
      expect(sh.at).toBeGreaterThanOrEqual(SHATTER_BREAK);
      expect(sh.at).toBeLessThan(1);
      byRing.set(sh.ring, [...(byRing.get(sh.ring) ?? []), sh.at]);
    }
    // Ring by ring outward: every inner piece lets go before the outer ring's last.
    const rings = [...byRing.keys()].sort();
    for (let k = 1; k < rings.length; k++) {
      expect(Math.min(...byRing.get(rings[k - 1])!)).toBeLessThan(
        Math.min(...byRing.get(rings[k])!),
      );
    }
    // A wipe's lighter shatter has fewer pieces.
    expect(shatterPlan("card-1", 80, 112, { light: true }).shards.length).toBeLessThan(
      s.shards.length,
    );
  });

  it("the shatter's pieces cover the whole card, corners included", () => {
    const s = shatterPlan("card-3", 80, 112);
    const pts = s.shards.flatMap((sh) =>
      sh.clip
        .slice("polygon(".length, -1)
        .split(", ")
        .map((p) => p.split(" ").map((v) => Number.parseFloat(v))),
    );
    for (const [cx, cy] of [
      [0, 0],
      [100, 0],
      [100, 100],
      [0, 100],
    ]) {
      expect(pts.some(([x, y]) => x === cx && y === cy)).toBe(true);
    }
  });

  it("a tapped card's pieces still fall down the screen", () => {
    const upright = shatterPlan("card-4", 80, 112);
    const tapped = shatterPlan("card-4", 80, 112, { rot: 90 });
    // On screen, (dx, dy) in the card's frame turned 90 degrees is (-dy, dx).
    const fall = (p: typeof upright, rot: number) =>
      p.shards.reduce((t, sh) => {
        const a = (rot * Math.PI) / 180;
        return t + sh.dx * Math.sin(a) + sh.dy * Math.cos(a);
      }, 0);
    expect(fall(upright, 0)).toBeGreaterThan(0);
    expect(fall(tapped, 90)).toBeGreaterThan(0);
  });

  it("an exile's specks start outside the card and spiral in (D3)", () => {
    const m = voidMotes("e1", 80, 112);
    expect(m).toEqual(voidMotes("e1", 80, 112));
    for (const o of m) {
      expect(o.r).toBeGreaterThan(Math.hypot(80, 112) / 2);
      expect(o.spin).toBeGreaterThan(180);
    }
  });

  it("a vortex's spiral starts where the card stood and ends at the centre (D11)", () => {
    const from = { x: 100, y: 40 };
    const to = { x: 400, y: 240 };
    const p = spiralPath(from, to);
    expect(p[0]).toEqual({ x: 0, y: 0 });
    expect(p.at(-1)!.x).toBeCloseTo(300, 0);
    expect(p.at(-1)!.y).toBeCloseTo(200, 0);
    // It closes on the centre as it goes.
    const dist = p.map((q) => Math.hypot(q.x + from.x - to.x, q.y + from.y - to.y));
    for (let i = 1; i < dist.length; i++) expect(dist[i]).toBeLessThan(dist[i - 1]);
    expect(swirlArms(5)).toHaveLength(5);
  });

  it("a destroy wipe's cuts cross the whole board at spread-out heights (D10)", () => {
    const c = windCuts("wipe-9", 7);
    expect(c).toHaveLength(7);
    expect(c).toEqual(windCuts("wipe-9", 7));
    for (const cut of c) {
      expect(cut.y).toBeGreaterThan(0);
      expect(cut.y).toBeLessThan(100);
      expect(cut.at + cut.len).toBeLessThanOrEqual(1);
    }
    expect(new Set(c.map((x) => Math.round(x.y / 10))).size).toBeGreaterThan(4);
  });

  it("a sacrifice's knife lands before the card shatters, sooner inside a wipe (D5)", () => {
    for (const light of [false, true]) {
      const b = sacrificeBeats(light);
      expect(b.stab).toBeLessThan(b.shatter);
      expect(b.shatter).toBeLessThan(0.5);
    }
    expect(sacrificeBeats(true).stab).toBeLessThan(sacrificeBeats(false).stab);
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
