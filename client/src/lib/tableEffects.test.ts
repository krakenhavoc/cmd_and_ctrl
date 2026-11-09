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
  CRACK_JOINTS,
  CRACK_SOUND,
  CRACK_VOLUME,
  DAY_NIGHT_FLIP_DELAY_MS,
  EFFECT_BASE_MS,
  EFFECT_CAP_MS,
  EFFECT_SHORT_CAP_MS,
  EFFECT_SOUND,
  EFFECT_STILL_HOLD_MS,
  PILE_LAND,
  PILE_SHOWN_MAX,
  PILE_STAGGER,
  PILE_SWEEP_MIN,
  WIPE_CARD_MS,
  WIPE_CARD_VOLUME,
  WIPE_MIN,
  WIPE_SWEEP_MS,
  WIPE_TOTAL_MS,
  WIPE_VORTEX_MS,
  SACRIFICE_KNIFE_SOUND,
  SHATTER_BREAK,
  TRANSFORM_TURN,
  TOKEN_SPREAD_MS,
  TOKEN_STAGGER_MS,
  attachKind,
  busyScale,
  counterBurst,
  counterTone,
  crackKind,
  crackPlan,
  cueMs,
  cueSound,
  cueTiming,
  dayNightOf,
  effectLate,
  effectMotion,
  effectsScheduled,
  exitKind,
  FIZZLE_MOTES,
  fizzleCue,
  drainCues,
  crownCues,
  fizzleMotes,
  headlines,
  isEffectsMode,
  isSelfCost,
  pileBaseMs,
  pileKind,
  pileShown,
  planEffects,
  reanimateTo,
  REANIMATE_HAND_VOLUME,
  RIFFLE_CARDS,
  searchedLibrary,
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
import { pileLandAt, pileWindows } from "./effectCueMotion";
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
    // Into a graveyard from a hand is a discard's cue now (#4445).
    expect(kinds([zone("a", "exile", {}, "hand")])).toEqual([]);
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

describe("cards are milled or discarded into a graveyard (#4445)", () => {
  // A mill is one `mill` entry per player (`amount` the count, `card_ids`
  // the cards the text names); a discard is one `discard` entry per card.
  const mill = (seat: number, ids: string[], extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "mill",
    turn: 3,
    seat,
    amount: ids.length,
    card_ids: ids,
    text: "",
    ...extra,
  });
  const discard = (seat: number, id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "discard",
    turn: 3,
    seat,
    card_id: id,
    old_zone: "hand",
    new_zone: "graveyard",
    cause: "effect",
    text: "",
    ...extra,
  });

  it("a mill is one cue with its cards and its count, and plays the mill sound", () => {
    const p = plan([mill(1, ["x", "y", "z"], { batch: 4 })]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.cardIDs, q.count, q.seat, q.batch])).toEqual([
      ["mill", "x", ["x", "y", "z"], 3, 1, 4],
    ]);
    expect(cueSound(p.cues[0])).toEqual({ name: "mill", volume: 1 });
    expect(EFFECT_SOUND.mill).toBe("mill");
    expect(EFFECT_SOUND.discard).toBe("discard");
    // A big mill names its first five; its count is all of them.
    const big = plan([mill(0, ["a", "b", "c", "d", "e"], { amount: 9, batch: 5 })]);
    expect(big.cues[0].count).toBe(9);
    // A server that predates card_ids still says how many.
    const old = plan([mill(0, [], { amount: 2, card_ids: undefined, batch: 6 })]);
    expect(old.cues.map((q) => [q.kind, q.cardID, q.count])).toEqual([["mill", null, 2]]);
  });

  it("a hand's discards in one batch are one cue, and one sound however many hands", () => {
    // A Windfall: each player discards their hand.
    const p = plan([
      discard(0, "a1", { batch: 8 }),
      discard(0, "a2", { batch: 8 }),
      discard(1, "b1", { batch: 8 }),
    ]);
    expect(p.cues.map((q) => [q.kind, q.seat, q.cardIDs, q.count])).toEqual([
      ["discard", 0, ["a1", "a2"], 2],
      ["discard", 1, ["b1"], 1],
    ]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual(["discard", null]);
    expect(headlines(p)).toBe(1);
    // Two batches are two moments.
    const q = plan([discard(0, "a", { batch: 9 }), discard(0, "b", { batch: 10 })]);
    expect(q.cues.map((c) => cueSound(c)?.name)).toEqual(["discard", "discard"]);
    // A mill and a discard in one batch (a looting that mills) are two sounds.
    const r = plan([mill(0, ["x"], { batch: 11 }), discard(0, "a", { batch: 11 })]);
    expect(r.cues.map((c) => cueSound(c)?.name)).toEqual(["mill", "discard"]);
  });

  it("the cleanup step's discard and a cost's are discards; a cycling is one too", () => {
    expect(kinds([discard(0, "a", { cause: "cleanup", batch: 3 })])).toEqual(["discard"]);
    expect(kinds([discard(0, "a", { cause: "cost", batch: 3 })])).toEqual(["discard"]);
    const cycle: LogEvent = {
      seq: ++seq,
      kind: "cycle",
      turn: 3,
      seat: 0,
      card_id: "monitor",
      old_zone: "hand",
      new_zone: "graveyard",
      text: "",
    };
    expect(plan([cycle]).cues.map((q) => [q.kind, q.cardIDs])).toEqual([["discard", ["monitor"]]]);
  });

  it("a card put into a graveyard from a library or a hand is the same cue; from the battlefield it is a death", () => {
    // Surveil's graveyard half, an explore's nonland, a sandbox move.
    expect(pileKind(zone("s", "graveyard", { seat: 0 }, "library"))).toBe("mill");
    expect(pileKind(zone("h", "graveyard", { seat: 0 }, "hand"))).toBe("discard");
    expect(pileKind(zone("d", "graveyard"))).toBeNull();
    expect(kinds([zone("d", "graveyard", { cause: "destroy" })])).toEqual(["destroy"]);
    // Two cards surveilled into one graveyard are one cue.
    const p = plan([
      zone("s1", "graveyard", { seat: 0, batch: 2 }, "library"),
      zone("s2", "graveyard", { seat: 0, batch: 2 }, "library"),
    ]);
    expect(p.cues.map((q) => [q.kind, q.cardIDs, q.count])).toEqual([["mill", ["s1", "s2"], 2]]);
  });

  it("only into a graveyard: a madness card discarded into exile, or a library to exile, is no cue", () => {
    expect(kinds([discard(0, "m", { new_zone: "exile" })])).toEqual([]);
    expect(kinds([zone("x", "exile", { seat: 0 }, "library")])).toEqual([]);
    expect(kinds([mill(0, [], { amount: 0 })])).toEqual([]);
  });

  it("a milled or discarded commander that goes home is both cues: into the graveyard, then home", () => {
    // CR 903.9a: it is put into its graveyard first, and its owner's
    // choice is a zone entry after that, in the same batch or a later one.
    const c = ctx({ isCommander: (id) => id === "cmdr" });
    const p = plan(
      [
        mill(0, ["cmdr"], { batch: 3 }),
        zone("cmdr", "command", { seat: 0, batch: 3 }, "graveyard"),
        discard(1, "cmdr2", { batch: 4 }),
      ],
      c,
    );
    expect(p.cues.map((q) => [q.kind, cueSound(q)?.name ?? null])).toEqual([
      ["mill", "mill"],
      ["commander_return", "commander_return"],
      ["discard", "discard"],
    ]);
  });

  it("each card's slide is as long as the others, the next a stagger in, and the last lands as the cue ends", () => {
    expect(pileWindows(1)).toEqual([{ a: 0, b: 1 }]);
    const w = pileWindows(4);
    const len = w[0].b - w[0].a;
    for (const [i, x] of w.entries()) {
      expect(x.b - x.a).toBeCloseTo(len, 9);
      expect(x.a).toBeCloseTo(i * PILE_STAGGER * len, 9);
    }
    expect(w[3].b).toBeCloseTo(1, 9);
    // A sweep's cards go together.
    expect(pileWindows(3, true)).toEqual([
      { a: 0, b: 1 },
      { a: 0, b: 1 },
      { a: 0, b: 1 },
    ]);
  });

  it("slides up to four cards one after another, and five or more as one sweep", () => {
    expect([1, 2, 4, 5, 9].map(pileShown)).toEqual([1, 2, PILE_SHOWN_MAX, 3, 3]);
    expect(PILE_SWEEP_MIN).toBe(5);
    expect(pileBaseMs("mill", 1)).toBe(EFFECT_BASE_MS.mill);
    expect(pileBaseMs("discard", 1)).toBe(EFFECT_BASE_MS.discard);
    expect(pileBaseMs("mill", 4)).toBeCloseTo(EFFECT_BASE_MS.mill * (1 + 3 * PILE_STAGGER), 6);
    expect(pileBaseMs("mill", 7)).toBeCloseTo(EFFECT_BASE_MS.mill * 1.25, 6);
    // Every length fits the Short cap, so Full and Short play the same.
    for (const kind of ["mill", "discard"] as const) {
      for (const n of [1, 2, 3, 4, 5, 40]) {
        const full = cueTiming({ kind, weight: "full", delayMs: 0, count: n }, "full", 1);
        expect(full.ms).toBeCloseTo(pileBaseMs(kind, n), 6);
        expect(full.ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS);
        expect(cueTiming({ kind, weight: "full", delayMs: 0, count: n }, "short", 1)).toEqual(full);
      }
      expect(cueTiming({ kind, weight: "full", delayMs: 0, count: 3 }, "still", 1)).toEqual({
        delayMs: 0,
        ms: EFFECT_STILL_HOLD_MS,
      });
    }
  });
});

describe("a card returns from a graveyard to the battlefield or a hand (#4446)", () => {
  // The server logs each as a `zone` entry from "graveyard": to
  // "battlefield" (Reanimate, unearth, a land put back; `seat` the new
  // controller) or to "hand" (Regrowth). log_reanimate_test.go pins it.
  const back = (id: string, to: string, extra: Partial<LogEvent> = {}): LogEvent =>
    zone(id, to, { seat: 0, ...extra }, "graveyard");
  const token = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "token",
    turn: 3,
    seat: 0,
    card_id: id,
    text: "",
    ...extra,
  });

  it("a card out of a graveyard onto the battlefield or into a hand is a reanimate, with its sound", () => {
    expect(reanimateTo(back("r", "battlefield"))).toBe("battlefield");
    expect(reanimateTo(back("g", "hand"))).toBe("hand");
    const p = plan([back("r", "battlefield", { batch: 3 })]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.cardIDs, q.count, q.reanimateTo])).toEqual([
      ["reanimate", "r", ["r"], 1, "battlefield"],
    ]);
    expect(cueSound(p.cues[0])).toEqual({ name: "reanimate", volume: 1 });
    expect(EFFECT_SOUND.reanimate).toBe("reanimate");
    // To a hand: the quieter variant.
    const h = plan([back("g", "hand", { batch: 4 })]);
    expect(h.cues.map((q) => [q.kind, q.reanimateTo])).toEqual([["reanimate", "hand"]]);
    expect(cueSound(h.cues[0])).toEqual({ name: "reanimate", volume: REANIMATE_HAND_VOLUME });
    expect(REANIMATE_HAND_VOLUME).toBeLessThan(1);
  });

  it("anywhere else is none: a graveyard to exile or a library, a hand or a library to the battlefield", () => {
    for (const e of [
      back("x", "exile"),
      back("l", "library"),
      zone("h", "battlefield", { seat: 0 }, "hand"),
      zone("s", "battlefield", { seat: 0 }, "stack"),
      zone("t", "hand", { seat: 0 }, "library"),
    ]) {
      expect(reanimateTo(e)).toBeNull();
      expect(kinds([e]).filter((k) => k === "reanimate")).toEqual([]);
    }
    // A spell cast from a graveyard (flashback, escape) is a cast.
    const cast: LogEvent = {
      seq: ++seq,
      kind: "cast",
      turn: 3,
      seat: 0,
      card_id: "f",
      old_zone: "graveyard",
      text: "",
    };
    expect(kinds([cast])).toEqual([]);
  });

  it("a token copy of a card in a graveyard is a token, not a reanimation", () => {
    // Embalm, Feldon of the Third Path: the card stays where it is (or is
    // exiled as the cost), and the copy is created, which the log tells
    // as a `token` entry (CR 111.1); a token never comes back (CR 111.8).
    const p = plan([
      zone("cat", "exile", { seat: 0, batch: 5 }, "graveyard"),
      token("copy", { batch: 5 }),
    ]);
    expect(p.cues.map((q) => q.kind)).toEqual(["token_create"]);
  });

  it("one effect's cards are one cue, a stagger apart, with one sound; two batches are two", () => {
    // Splendid Reclamation, Living Death: several cards, several graveyards.
    const p = plan([
      back("l1", "battlefield", { batch: 6 }),
      back("l2", "battlefield", { batch: 6 }),
      back("o1", "battlefield", { batch: 6, seat: 1 }),
    ]);
    expect(p.cues.map((q) => [q.kind, q.cardIDs, q.count])).toEqual([
      ["reanimate", ["l1", "l2", "o1"], 3],
    ]);
    expect(headlines(p)).toBe(1);
    // Unearth twice: two moments.
    const q = plan([
      back("a", "battlefield", { batch: 7 }),
      back("b", "battlefield", { batch: 8 }),
    ]);
    expect(q.cues.map((c) => cueSound(c)?.name)).toEqual(["reanimate", "reanimate"]);
    // One to the battlefield and one to a hand in one batch: one moment.
    const r = plan([back("a", "battlefield", { batch: 9 }), back("b", "hand", { batch: 9 })]);
    expect(r.cues.map((c) => [c.reanimateTo, cueSound(c)?.name ?? null])).toEqual([
      ["battlefield", "reanimate"],
      ["hand", null],
    ]);
    expect(headlines(r)).toBe(1);
  });

  it("a card returned under another player's control is the same cue", () => {
    // Reanimate on an opponent's creature: `seat` is the new controller;
    // the layer reads whose graveyard it rose from off the card's owner.
    const p = plan([back("theirs", "battlefield", { seat: 0, batch: 3 })]);
    expect(p.cues.map((q) => [q.kind, q.seat])).toEqual([["reanimate", 0]]);
  });

  it("a reanimated planeswalker's arrival and an Aura returned attached follow the rise", () => {
    const pw = ctx({ isPlaneswalker: (id) => id === "jace" });
    const p = plan([back("jace", "battlefield", { batch: 4 })], pw);
    expect(p.cues.map((q) => [q.kind, cueSound(q)?.name ?? null])).toEqual([
      ["pw_enter", null],
      ["reanimate", "reanimate"],
    ]);
    expect(headlines(p)).toBe(1);
    const aura = (id: string, batch: number): LogEvent => ({
      seq: ++seq,
      kind: "attach",
      turn: 3,
      seat: 0,
      card_id: id,
      target: "bear",
      label: "Aura",
      batch,
      text: "",
    });
    const q = plan([back("rancor", "battlefield", { batch: 5 }), aura("rancor", 5)]);
    expect(q.cues.map((c) => [c.kind, cueSound(c)?.name ?? null])).toEqual([
      ["reanimate", "reanimate"],
      ["attach_aura", null],
    ]);
    // Another Aura in that batch, not reanimated, still plays its own.
    const r = plan([
      back("rancor", "battlefield", { batch: 6 }),
      aura("rancor", 6),
      aura("other", 6),
    ]);
    expect(r.cues.map((c) => [c.kind, c.cardID, cueSound(c)?.name ?? null])).toEqual([
      ["reanimate", "rancor", "reanimate"],
      ["attach_aura", "rancor", null],
      ["attach_aura", "other", "attach_aura"],
    ]);
  });

  it("a commander going home from its graveyard is commander_return, not this cue", () => {
    const c = ctx({ isCommander: (id) => id === "cmdr" });
    expect(kinds([back("cmdr", "command", { batch: 3 })], c)).toEqual(["commander_return"]);
    // Left there and later reanimated, or regrown (its owner may choose
    // not to send it home, CR 903.9a, 903.9b), it is this cue.
    expect(kinds([back("cmdr", "battlefield", { batch: 4 })], c)).toEqual(["reanimate"]);
    expect(kinds([back("cmdr", "hand", { batch: 5 })], c)).toEqual(["reanimate"]);
  });

  it("rises up to four cards one after another and five or more as one sweep, about 0.8 s for one", () => {
    expect(EFFECT_BASE_MS.reanimate).toBe(800);
    expect(pileBaseMs("reanimate", 1)).toBe(800);
    expect(pileBaseMs("reanimate", 3)).toBeCloseTo(800 * (1 + 2 * PILE_STAGGER), 6);
    expect(pileBaseMs("reanimate", 12)).toBeCloseTo(800 * 1.25, 6);
    for (const n of [1, 2, 4, 5, 12]) {
      const full = cueTiming(
        { kind: "reanimate", weight: "full", delayMs: 0, count: n },
        "full",
        1,
      );
      expect(full.ms).toBeCloseTo(pileBaseMs("reanimate", n), 6);
      // Short caps the longer runs, as for every cue.
      const short = cueTiming(
        { kind: "reanimate", weight: "full", delayMs: 0, count: n },
        "short",
        1,
      );
      expect(short.ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS);
    }
    expect(
      cueTiming({ kind: "reanimate", weight: "full", delayMs: 0, count: 2 }, "still", 1),
    ).toEqual({ delayMs: 0, ms: EFFECT_STILL_HOLD_MS });
    // Each card's tile takes over from its copy as that copy lands.
    expect(pileLandAt(1, 0)).toBeCloseTo(PILE_LAND, 9);
    const w = pileWindows(3);
    for (let i = 0; i < 3; i++) {
      expect(pileLandAt(3, i)).toBeCloseTo(w[i].a + (w[i].b - w[i].a) * PILE_LAND, 9);
    }
    expect(pileLandAt(3, 2, true)).toBeCloseTo(PILE_LAND, 9);
  });
});

describe("a player searches a library (#4447)", () => {
  // The server writes one entry per search with `library_seat`: its
  // `search` line, or the `reveal` line that took its place when the
  // search revealed what it took (a fetch land, Bribery).
  // log_library_search_test.go pins it.
  const search = (extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "search",
    turn: 3,
    seat: 0,
    target_seat: 0,
    library_seat: 0,
    label: "search_shuffled",
    batch: 5,
    text: "",
    ...extra,
  });
  const reveal = (extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "reveal",
    turn: 3,
    seat: 0,
    old_zone: "library",
    amount: 1,
    batch: 5,
    text: "",
    ...extra,
  });

  it("a search, found or not, is a riffle on the searched library, with its sound", () => {
    expect(searchedLibrary(search())).toBe(0);
    const p = plan([search({ batch: 7 })]);
    expect(p.cues.map((q) => [q.kind, q.seat, q.librarySeat, q.cardID, q.batch])).toEqual([
      ["library_search", 0, 0, null, 7],
    ]);
    expect(cueSound(p.cues[0])).toEqual({ name: "library_search", volume: 1 });
    expect(EFFECT_SOUND.library_search).toBe("library_search");
    expect(SOUND_CATEGORY.library_search).toBe("effects");
    expect(SOUND_MANIFEST.library_search).toEqual([]);
    // A search that did not shuffle, or one that found nothing, is the
    // same cue: the log says neither what nor whether.
    expect(kinds([search({ label: "" })])).toEqual(["library_search"]);
  });

  it("a fetch land's reveal carries the search, and a reveal no search made is none", () => {
    // "P1 fetched Forest from their library onto the battlefield".
    const fetched = reveal({ library_seat: 0, batch: 9 });
    expect(searchedLibrary(fetched)).toBe(0);
    expect(kinds([fetched])).toEqual(["library_search"]);
    // Dark Confidant, a "reveal the top card": no library_seat.
    expect(searchedLibrary(reveal())).toBeNull();
    expect(kinds([reveal()])).toEqual([]);
  });

  it("a plain shuffle is no search, and neither is anything else", () => {
    const shuffle = search({ label: "shuffle", library_seat: undefined });
    expect(searchedLibrary(shuffle)).toBeNull();
    expect(kinds([shuffle])).toEqual([]);
    expect(searchedLibrary(zone("a", "hand", {}, "library"))).toBeNull();
    expect(searchedLibrary({ ...sacrifice("f"), library_seat: 0 })).toBeNull();
  });

  it("a search line from an older server reads its target_seat", () => {
    const old = search({ library_seat: undefined, target_seat: 1, seat: 0 });
    expect(searchedLibrary(old)).toBe(1);
    expect(plan([old]).cues.map((q) => [q.seat, q.librarySeat])).toEqual([[0, 1]]);
  });

  it("another player's library riffles that library, not the searcher's", () => {
    // Bribery: P1 searches P2's library and reveals what it takes.
    const p = plan([reveal({ seat: 0, library_seat: 1, batch: 4 })]);
    expect(p.cues.map((q) => [q.kind, q.seat, q.librarySeat])).toEqual([["library_search", 0, 1]]);
  });

  it("one effect's searches are one sound; one library searched twice in a batch is one riffle", () => {
    // "Each player searches their library": one riffle per library, one sound.
    const p = plan([
      search({ seat: 0, library_seat: 0, batch: 6 }),
      search({ seat: 1, library_seat: 1, batch: 6 }),
    ]);
    expect(p.cues.map((q) => [q.librarySeat, cueSound(q)?.name ?? null])).toEqual([
      [0, "library_search"],
      [1, null],
    ]);
    expect(headlines(p)).toBe(1);
    // CR 701.23h: searched again before the shuffle, still one search.
    expect(kinds([search({ batch: 8 }), search({ batch: 8 })])).toEqual(["library_search"]);
    // Two tutors are two moments.
    const q = plan([search({ batch: 10 }), search({ batch: 11 })]);
    expect(q.cues.map((c) => cueSound(c)?.name)).toEqual(["library_search", "library_search"]);
  });

  it("the found card's own cues play beside it, each with its own sound", () => {
    // A fetch land resolving: the land's arrival is a zone entry the
    // riffle does not touch. A found planeswalker keeps its arrival and
    // its sound; a found Aura entering attached keeps its attach cue; a
    // found commander its owner sends to the command zone instead of a
    // hand (CR 903.9b) is commander_return.
    const pw = ctx({
      isPlaneswalker: (id) => id === "jace",
      isCommander: (id) => id === "cmdr",
    });
    const p = plan(
      [
        reveal({ library_seat: 0, batch: 3 }),
        zone("jace", "battlefield", { seat: 0, batch: 3 }, "library"),
        {
          seq: ++seq,
          kind: "attach",
          turn: 3,
          seat: 0,
          card_id: "rancor",
          target: "bear",
          label: "Aura",
          batch: 3,
          text: "",
        },
        zone("cmdr", "command", { seat: 0, batch: 3 }, "library"),
      ],
      pw,
    );
    expect(p.cues.map((q) => [q.kind, cueSound(q)?.name ?? null])).toEqual([
      ["library_search", "library_search"],
      ["pw_enter", "pw_enter"],
      ["attach_aura", "attach_aura"],
      ["commander_return", "commander_return"],
    ]);
    // A fetch land's sacrifice is its cost, in the activation's batch.
    expect(
      kinds([sacrifice("wilds", { batch: 2 }), zone("wilds", "graveyard", { batch: 2 })]),
    ).toEqual(["sacrifice"]);
  });

  it("riffles in about half a second, the same in Short, and is still when motion is off", () => {
    expect(EFFECT_BASE_MS.library_search).toBe(500);
    expect(RIFFLE_CARDS).toBeGreaterThanOrEqual(3);
    for (const mode of ["full", "short"] as const) {
      expect(cueTiming({ kind: "library_search", weight: "full", delayMs: 0 }, mode, 1)).toEqual({
        delayMs: 0,
        ms: 500,
      });
    }
    expect(cueTiming({ kind: "library_search", weight: "full", delayMs: 0 }, "still", 1)).toEqual({
      delayMs: 0,
      ms: EFFECT_STILL_HOLD_MS,
    });
    // The riffled backs go one after another, the last settling as it ends.
    const w = pileWindows(RIFFLE_CARDS);
    expect(w[0].a).toBe(0);
    expect(w[RIFFLE_CARDS - 1].b).toBeCloseTo(1, 9);
  });
});

describe("a double-faced card transforms, or day and night changes (#4448)", () => {
  // The engine's log, as log_transform_test.go pins it: a `transform`
  // entry per permanent turned over (`label` the face it was), a `meld`
  // entry naming the melded permanent, and a `day_night` entry with what
  // the game became; night falling and the werewolves it turns over share
  // one batch, the change first.
  const transform = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "transform",
    turn: 3,
    seat: 0,
    card_id: id,
    label: "Front Face",
    text: "Front Face transformed into Back Face",
    ...extra,
  });
  const meld = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "meld",
    turn: 3,
    seat: 0,
    card_id: id,
    label: "Midnight Scavengers and Graf Rats",
    text: "Midnight Scavengers and Graf Rats melded into Chittering Host",
    ...extra,
  });
  const dayNight = (label: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "day_night",
    turn: 4,
    seat: -1,
    label,
    text: `It becomes ${label}`,
    ...extra,
  });
  const counters = (id: string, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "counters",
    turn: 3,
    seat: -1,
    card_id: id,
    label: "+1/+1",
    amount: 1,
    delta: 1,
    text: "",
    ...extra,
  });

  it("a transform is one flip on its card, with the transform sound", () => {
    const p = plan([transform("a", { batch: 4 })]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.batch, q.delayMs, q.follows])).toEqual([
      ["transform", "a", 4, 0, undefined],
    ]);
    expect(cueSound(p.cues[0])).toEqual({ name: "transform", volume: 1 });
    expect(EFFECT_SOUND.transform).toBe("transform");
    // A convert is the same entry, marked converted (CR 701.28a).
    expect(kinds([transform("b", { converted: true, batch: 5 })])).toEqual(["transform"]);
  });

  it("several permanents turned over at once are one sound, the first's", () => {
    const p = plan([
      transform("a", { batch: 6 }),
      transform("b", { batch: 6 }),
      transform("c", { batch: 6 }),
      transform("d", { batch: 7 }),
    ]);
    expect(p.cues.map((q) => [q.cardID, q.follows === true])).toEqual([
      ["a", false],
      ["b", true],
      ["c", true],
      ["d", false],
    ]);
    expect(p.cues.filter((q) => cueSound(q) !== null)).toHaveLength(2);
    expect(headlines(p)).toBe(2);
  });

  it("night falling is one wash and one sound, the bell; the werewolves it turns over follow, a beat later", () => {
    const p = plan([
      dayNight("night", { batch: 9 }),
      transform("wolf-1", { batch: 9 }),
      transform("wolf-2", { batch: 9 }),
    ]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.sky, q.follows === true, q.delayMs])).toEqual([
      ["day_night", null, "night", false, 0],
      ["transform", "wolf-1", undefined, true, DAY_NIGHT_FLIP_DELAY_MS],
      ["transform", "wolf-2", undefined, true, DAY_NIGHT_FLIP_DELAY_MS],
    ]);
    expect(p.cues.map((q) => cueSound(q))).toEqual([{ name: "day_night", volume: 1 }, null, null]);
    expect(headlines(p)).toBe(1);
    // Day breaking is the same, with no werewolf to turn as well.
    expect(plan([dayNight("day", { batch: 10 })]).cues.map((q) => [q.kind, q.sky])).toEqual([
      ["day_night", "day"],
    ]);
    expect(dayNightOf(dayNight("dusk"))).toBeNull();
    expect(kinds([dayNight("dusk", { batch: 11 })])).toEqual([]);
  });

  it("a flip after night falls still ends inside the Short cap", () => {
    const t = cueTiming(
      { kind: "transform", weight: "full", delayMs: DAY_NIGHT_FLIP_DELAY_MS },
      "short",
      1,
    );
    expect(t.delayMs + t.ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS);
    expect(EFFECT_BASE_MS.transform).toBe(600);
    expect(cueTiming({ kind: "day_night", weight: "full", delayMs: 0 }, "short", 1).ms).toBe(1000);
  });

  it("a meld is one flip on the melded permanent, unless it is a planeswalker, whose arrival it is", () => {
    const entries = [
      zone("rats", "exile", { batch: 12 }),
      zone("host", "battlefield", { batch: 12 }, "exile"),
      meld("host", { batch: 12 }),
    ];
    const p = plan(entries);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.melded])).toEqual([
      ["exile", "rats", undefined],
      ["transform", "host", true],
    ]);
    // One moment, one sound: the cards it was melded from leave in the
    // exile's void, silently, and the flip plays.
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual([null, "transform"]);
    // Urza, Planeswalker: its pw_enter is the arrival, and no flip.
    const urza = plan(
      [zone("urza", "battlefield", { batch: 13 }, "exile"), meld("urza", { batch: 13 })],
      ctx({ isPlaneswalker: (id) => id === "urza" }),
    );
    expect(urza.cues.map((q) => q.kind)).toEqual(["pw_enter"]);
    expect(cueSound(urza.cues[0])?.name).toBe("pw_enter");
    // An exile in a batch with no meld keeps its sound.
    expect(cueSound(plan([zone("x", "exile", { batch: 19 })]).cues[0])?.name).toBe("exile");
  });

  it("an entry with no card is no flip, and a card's entries in one batch are one", () => {
    expect(kinds([transform("a", { card_id: undefined, batch: 14 })])).toEqual([]);
    expect(kinds([transform("a", { batch: 15 }), transform("a", { batch: 15 })])).toEqual([
      "transform",
    ]);
  });

  it("what is not a transform is no flip: an entry at night, a disturb cast, a face turned up, a return transformed", () => {
    // A daybound card entering at night, a disturbed creature entering
    // back face up, a card returned transformed: each only a zone move.
    expect(kinds([zone("wolf", "battlefield", { batch: 16 }, "hand")])).toEqual([]);
    expect(
      kinds([
        {
          seq: ++seq,
          kind: "cast",
          turn: 3,
          seat: 0,
          card_id: "drifter",
          old_zone: "graveyard",
          text: "",
        },
      ]),
    ).toEqual([]);
    expect(
      kinds([
        { seq: ++seq, kind: "turn_face_up", turn: 3, seat: 0, card_id: "morph", text: "" },
        { seq: ++seq, kind: "flip_permanent", turn: 3, seat: 0, card_id: "tok", text: "" },
      ]),
    ).toEqual([]);
  });

  it("counters put on a card that flips in the same batch pulse once it turns edge-on", () => {
    const p = plan([
      transform("a", { batch: 17 }),
      counters("a", { batch: 17 }),
      counters("b", { batch: 17 }),
    ]);
    const pulse = (id: string) => p.cues.find((q) => q.kind === "counters" && q.cardID === id)!;
    expect(pulse("a").delayMs).toBe(TRANSFORM_TURN * EFFECT_BASE_MS.transform);
    expect(pulse("b").delayMs).toBe(0);
    // Each its own moment, its own sound: the flip's, then the counter's.
    expect(cueSound(p.cues[0])).toEqual({ name: "transform", volume: 1 });
    expect(cueSound(pulse("a"))).toEqual({ name: "counter_add", volume: 1 });
    // After night falls the pulse waits for the later flip, and still
    // ends inside the Short cap.
    const n = plan([
      dayNight("night", { batch: 18 }),
      transform("w", { batch: 18 }),
      counters("w", { batch: 18 }),
    ]);
    const c = n.cues.find((q) => q.kind === "counters")!;
    expect(c.delayMs).toBe(DAY_NIGHT_FLIP_DELAY_MS + TRANSFORM_TURN * EFFECT_BASE_MS.transform);
    const t = cueTiming(c, "short", 1);
    expect(t.delayMs + t.ms).toBeLessThanOrEqual(EFFECT_SHORT_CAP_MS);
  });

  it("the sounds are effects-category entries with no takes yet", () => {
    for (const name of ["transform", "day_night"] as const) {
      expect(SOUND_CATEGORY[name]).toBe("effects");
      expect(SOUND_MANIFEST[name]).toEqual([]);
    }
  });
});

describe("a permanent is sacrificed for its own ability (#4427)", () => {
  // The server writes a `sacrifice` entry with `cause` "cost" for a
  // sacrifice paid as a cost, and `source` the permanent itself when the
  // cost was its own ability's. log_self_sacrifice_test.go pins it.
  const self = (id: string, extra: Partial<LogEvent> = {}): LogEvent =>
    sacrifice(id, { cause: "cost", source: id, batch: 4, ...extra });
  const lines: Record<string, string> = {
    wilds: "Land",
    vista: "Land",
    t1: "Token Artifact — Treasure",
    t2: "Token Artifact — Treasure",
    t3: "Token Artifact — Treasure",
    clue: "Token Artifact — Clue",
    food: "Token Artifact — Food",
    blood: "Token Artifact — Blood",
    petal: "Artifact",
    elder: "Creature — Snake Shaman",
    spawn: "Token Creature — Eldrazi Spawn",
  };
  const seen = ctx({ typeLineOf: (id) => lines[id] });
  const search = (extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "reveal",
    turn: 3,
    seat: 0,
    old_zone: "library",
    amount: 1,
    library_seat: 0,
    text: "",
    ...extra,
  });

  it("tells a sacrifice paid for its own ability from every other", () => {
    expect(isSelfCost(self("t1"))).toBe(true);
    // A creature fed to Viscera Seer: a cost, of another permanent's ability.
    expect(isSelfCost(sacrifice("bear", { cause: "cost" }))).toBe(false);
    // An effect's sacrifice of the card itself (Ball Lightning's end step).
    expect(isSelfCost(sacrifice("bolt", { source: "bolt" }))).toBe(false);
    // An effect's, a sandbox's, and a zone line.
    expect(isSelfCost(sacrifice("x", { source: "wipe" }))).toBe(false);
    expect(isSelfCost(sacrifice("x"))).toBe(false);
    expect(isSelfCost(zone("t1", "graveyard", { cause: "cost", source: "t1" }))).toBe(false);
    // A viewer who may not identify the card has no source: the altar.
    expect(isSelfCost(sacrifice("t1", { cause: "cost" }))).toBe(false);
  });

  it("reads what was spent off its type line; a creature keeps the altar", () => {
    expect(crackKind("Land")).toBe("land");
    expect(crackKind("Basic Land — Forest")).toBe("land");
    expect(crackKind("Artifact Land")).toBe("land");
    expect(crackKind("Token Artifact — Treasure")).toBe("treasure");
    expect(crackKind("Token Artifact — Clue")).toBe("clue");
    expect(crackKind("Token Artifact — Food")).toBe("food");
    expect(crackKind("Token Artifact — Blood")).toBe("other");
    expect(crackKind("Token Artifact — Map")).toBe("other");
    expect(crackKind("Artifact")).toBe("other");
    expect(crackKind("Enchantment")).toBe("other");
    expect(crackKind(undefined)).toBe("other");
    expect(crackKind("Creature — Snake Shaman")).toBeNull();
    expect(crackKind("Token Creature — Eldrazi Spawn")).toBeNull();
    expect(crackKind("Land Creature — Forest Dryad")).toBeNull();
  });

  it("a fetch land, a Treasure, a Clue and a Food crack, each with its own sound, and no altar", () => {
    for (const [id, crack, sound] of [
      ["wilds", "land", "fetch_crack"],
      ["t1", "treasure", "treasure_crack"],
      ["clue", "clue", "clue_crack"],
      ["food", "food", "food_crack"],
      ["blood", "other", "treasure_crack"],
      ["petal", "other", "treasure_crack"],
    ] as const) {
      const p = plan([self(id, { batch: 20 + id.length })], seen);
      expect(p.cues.map((q) => [q.kind, q.cardID, q.crackOf])).toEqual([["crack", id, crack]]);
      expect(p.wipes).toEqual([]);
      expect(cueSound(p.cues[0])).toEqual({ name: sound, volume: CRACK_VOLUME });
      expect(CRACK_SOUND[crack]).toBe(sound);
    }
    expect(EFFECT_SOUND.crack).toBeNull();
    // A little quieter than the other effect cues.
    expect(CRACK_VOLUME).toBeLessThan(1);
    expect(CRACK_VOLUME).toBeGreaterThan(0.5);
    // About 0.6 s, the same in Short.
    expect(EFFECT_BASE_MS.crack).toBe(600);
    expect(cueTiming({ kind: "crack", weight: "full", delayMs: 0 }, "short", 1).ms).toBe(600);
  });

  it("one sacrifice is one cue: a crack never plays the altar and the knife too", () => {
    const p = plan([self("wilds")], seen);
    expect(p.cues.map((q) => q.kind)).toEqual(["crack"]);
    expect(p.cues.some((q) => q.kind === "sacrifice")).toBe(false);
    // The knife's sound is the sacrifice cue's alone (TableEffects plays
    // it for a `sacrifice` cue), so a crack has neither.
    expect(cueSound(p.cues[0])?.name).not.toBe("sacrifice");
    expect(cueSound(p.cues[0])?.name).not.toBe(SACRIFICE_KNIFE_SOUND);
    // A creature sacrificed for its own ability (Sakura-Tribe Elder, an
    // Eldrazi Spawn) keeps the altar: that is what it is for.
    expect(kinds([self("elder")], seen)).toEqual(["sacrifice"]);
    expect(kinds([self("spawn")], seen)).toEqual(["sacrifice"]);
    // So does a creature fed to another permanent's ability, and an
    // effect's sacrifice.
    expect(kinds([sacrifice("bear", { cause: "cost", batch: 4 })], seen)).toEqual(["sacrifice"]);
    expect(kinds([sacrifice("t1", { batch: 4 })], seen)).toEqual(["sacrifice"]);
    // The exit's zone line, when the card did not go to the graveyard (a
    // Rest in Peace), is the same exit: one cue, read off new_zone first.
    expect(
      kinds(
        [self("t2", { batch: 8 }), zone("t2", "exile", { cause: "sacrifice", batch: 8 })],
        seen,
      ),
    ).toEqual(["exile"]);
  });

  it("several Treasures in one payment are one sound, and never a board wipe", () => {
    // The auto-tapper spends three Treasures on one spell as it is cast
    // (CR 601.2g): three lines, one batch, one frame, and the cast.
    const cast: LogEvent = {
      seq: 0,
      kind: "cast",
      turn: 3,
      seat: 0,
      card_id: "spell",
      batch: 9,
      text: "",
    };
    const entries = [self("t1", { batch: 9 }), self("t2", { batch: 9 }), self("t3", { batch: 9 })];
    cast.seq = ++seq;
    const p = plan([...entries, cast], seen);
    expect(p.wipes).toEqual([]);
    expect(p.cues.map((q) => [q.kind, q.cardID, q.follows === true])).toEqual([
      ["crack", "t1", false],
      ["crack", "t2", true],
      ["crack", "t3", true],
    ]);
    expect(p.cues.map((q) => cueSound(q)?.name ?? null)).toEqual(["treasure_crack", null, null]);
    // One payment counts once toward a busy table.
    expect(headlines(p)).toBe(1);
    // Its sound is capped on the payment, not on the batch: a Treasure
    // cracked later in the same priority window, in another frame, is a
    // moment of its own.
    const later = plan([self("t1", { batch: 9 })], seen).cues[0];
    const again = plan([self("t2", { batch: 9 })], seen).cues[0];
    expect(p.cues[0].soundKey).toBe(p.cues[1].soundKey);
    expect(later.soundKey).not.toBe(again.soundKey);
    expect(soundBatch("g1", 9, later.soundKey)).not.toBe(soundBatch("g1", 9, again.soundKey));
    expect(soundBatch("g1", 9, later.soundKey)).not.toBe(soundBatch("g1", 9));
  });

  it("a Treasure spent beside a creature sacrifice keeps each its own cue", () => {
    const p = plan(
      [self("t1", { batch: 3 }), sacrifice("bear", { cause: "cost", batch: 3 })],
      seen,
    );
    expect(p.cues.map((q) => [q.kind, q.cardID])).toEqual([
      ["crack", "t1"],
      ["sacrifice", "bear"],
    ]);
    expect(p.cues.map((q) => cueSound(q)?.name)).toEqual(["treasure_crack", "sacrifice"]);
    // A Treasure spent on a commander cast from the command zone: the
    // payment and the cast are two moments, each with its own sound.
    const paid = self("t1", { batch: 11 });
    const cast: LogEvent = {
      seq: ++seq,
      kind: "cast",
      turn: 3,
      seat: 0,
      card_id: "cmdr",
      old_zone: "command",
      batch: 11,
      text: "",
    };
    const c = plan([paid, cast], seen);
    expect(c.cues.map((q) => [q.kind, cueSound(q)?.name])).toEqual([
      ["crack", "treasure_crack"],
      ["commander_cast", "commander_cast"],
    ]);
  });

  it("a fetch land's crack comes as it is activated, and the search when it resolves", () => {
    // Frame one: the activation's batch, the sacrifice and the activate line.
    const activate: LogEvent = {
      seq: ++seq,
      kind: "activate",
      turn: 3,
      seat: 0,
      card_id: "wilds",
      batch: 4,
      text: "",
    };
    const one = plan([self("wilds", { batch: 4 }), activate], seen);
    expect(one.cues.map((q) => q.kind)).toEqual(["crack"]);
    // Frame two, a later batch: the "fetched" reveal and the land's arrival.
    const two = plan(
      [search({ batch: 5 }), zone("forest", "battlefield", { batch: 5 }, "library")],
      seen,
    );
    expect(two.cues.map((q) => q.kind)).toEqual(["library_search"]);
    expect(two.cues.some((q) => q.kind === "crack" || q.kind === "sacrifice")).toBe(false);
    // Each its own sound.
    expect(cueSound(one.cues[0])?.name).toBe("fetch_crack");
    expect(cueSound(two.cues[0])?.name).toBe("library_search");
  });

  it("splits the card along one jagged line, the halves falling apart and down", () => {
    const w = 80;
    const h = 112;
    const c = crackPlan("t1", w, h);
    expect(crackPlan("t1", w, h)).toEqual(c);
    expect(crackPlan("t2", w, h)).not.toEqual(c);
    // The line runs the card's length, end to end.
    const pts = c.path
      .slice(1)
      .split(" L")
      .map((p) => p.split(" ").map(Number));
    expect(pts).toHaveLength(CRACK_JOINTS);
    expect(pts[0][1]).toBe(0);
    expect(pts[pts.length - 1][1]).toBe(h);
    for (const [x] of pts) {
      expect(x).toBeGreaterThan(w * 0.25);
      expect(x).toBeLessThan(w * 0.75);
    }
    const [left, right] = c.halves;
    expect(left.clip.startsWith("polygon(")).toBe(true);
    // Upright: they part sideways, both fall, and turn outward.
    expect(left.dx).toBeLessThan(0);
    expect(right.dx).toBeGreaterThan(0);
    expect(left.dy).toBeGreaterThan(0);
    expect(right.dy).toBeGreaterThan(0);
    expect(left.rot).toBeLessThan(0);
    expect(right.rot).toBeGreaterThan(0);
    expect(left.ox).toBeLessThan(0);
    expect(right.ox).toBeGreaterThan(0);
    // A tapped land splits along its length, which lies across the
    // screen: the halves part up and down, and still fall down.
    const [up, down] = crackPlan("wilds", w, h, 90).halves;
    expect(Math.abs(up.dx)).toBeLessThan(1);
    expect(up.oy).toBeLessThan(0);
    expect(down.oy).toBeGreaterThan(0);
    expect(down.dy).toBeGreaterThan(up.dy);
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

describe("a table-wide drain (#4450)", () => {
  const loss = (seat: number, extra: Partial<LogEvent> = {}): LogEvent => ({
    seq: ++seq,
    kind: "life",
    seat,
    card_id: "konrad",
    amount: -1,
    batch: 21,
    text: "",
    ...extra,
  });

  it("one source's losses for several players in one batch are one cue", () => {
    const cues = plan([loss(0), loss(1), loss(2)]).cues;
    expect(cues.map((c) => c.kind)).toEqual(["drain_table"]);
    expect(cues[0].cardID).toBe("konrad");
    expect(cues[0].drainSeats).toEqual([0, 1, 2]);
    expect(cues[0].batch).toBe(21);
    expect(headlines(plan([loss(0), loss(1), loss(2)]))).toBe(1);
  });

  it("is no cue for one player, a gain, another batch, no source, or a damage spell", () => {
    expect(kinds([loss(1)])).toEqual([]);
    expect(kinds([loss(0, { amount: 2 }), loss(1, { amount: 2 })])).toEqual([]);
    expect(kinds([loss(0), loss(1, { batch: 22 })])).toEqual([]);
    expect(kinds([loss(0, { card_id: undefined }), loss(1, { card_id: undefined })])).toEqual([]);
    expect(kinds([loss(0, { batch: undefined }), loss(1, { batch: undefined })])).toEqual([]);
    const dmg: LogEvent = {
      seq: ++seq,
      kind: "damage",
      seat: -1,
      card_id: "konrad",
      batch: 21,
      text: "",
    };
    expect(kinds([dmg, loss(0), loss(1)])).toEqual([]);
  });

  it("two sources in one batch are two cues", () => {
    const cues = drainCues([
      loss(0),
      loss(1),
      loss(0, { card_id: "z" }),
      loss(2, { card_id: "z" }),
    ]);
    expect(cues.map((c) => c.cardID)).toEqual(["konrad", "z"]);
  });

  it("is a 0.8 s line with a low siphon, the same in Short", () => {
    expect(EFFECT_BASE_MS.drain_table).toBe(800);
    expect(EFFECT_SOUND.drain_table).toBe("drain_table");
    expect(SOUND_CATEGORY.drain_table).toBe("effects");
    expect(SOUND_MANIFEST.drain_table).toEqual([]);
    const c = drainCues([loss(0), loss(1)])[0];
    expect(cueTiming(c, "short", 1).ms).toBe(800);
  });
});

describe("the monarch and the initiative (#4450)", () => {
  const crown = (kind: "monarch" | "initiative", seat = 2): LogEvent => ({
    seq: ++seq,
    kind,
    seat,
    batch: 30,
    text: "",
  });

  it("each is a cue on the new holder's seat, with its own sound", () => {
    const [m] = plan([crown("monarch")]).cues;
    expect(m.kind).toBe("monarch");
    expect(m.seat).toBe(2);
    const [i] = plan([crown("initiative", 3)]).cues;
    expect(i.kind).toBe("initiative");
    expect(i.seat).toBe(3);
    for (const k of ["monarch", "initiative"] as const) {
      expect(EFFECT_SOUND[k]).toBe(k);
      expect(SOUND_CATEGORY[k]).toBe("effects");
      expect(SOUND_MANIFEST[k]).toEqual([]);
      expect(EFFECT_BASE_MS[k]).toBeLessThanOrEqual(1000);
    }
  });

  it("an entry with no seat is no cue", () => {
    expect(crownCues([crown("monarch", -1)])).toEqual([]);
  });
});
